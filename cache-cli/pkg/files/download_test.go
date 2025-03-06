package files

import (
	"os"
	"testing"
	"runtime"
	"github.com/semaphoreci/toolbox/cache-cli/pkg/storage"
	"github.com/stretchr/testify/require"
)

func Test__DownloadFromHTTP(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	sftpStorage, err := storage.NewSFTPStorage(storage.SFTPStorageOptions{
		URL:            "sftp-server:22",
		Username:       "tester",
		PrivateKeyPath: "/root/.ssh/semaphore_cache_key",
		Config: storage.StorageConfig{
			MaxSpace:   1024,
			SortKeysBy: storage.SortBySize,
		},
	})

	require.NoError(t, err)
	require.NoError(t, sftpStorage.Clear())
	require.NoError(t, sftpStorage.Store("abc", "testdata/test.txt"))

	t.Run("download works", func(t *testing.T) {
		f, err := DownloadFromHTTP("http://sftp-server:80", "test", "test", "abc")
		require.NoError(t, err)
		require.FileExists(t, f.Name())

		content, err := os.ReadFile(f.Name())
		require.NoError(t, err)
		require.Equal(t, "Test 123", string(content))
	})

	t.Run("download fails if URL is not reachable", func(t *testing.T) {
		_, err := DownloadFromHTTP("http://sftp-server:801", "test", "test", "abc")
		require.ErrorContains(t, err, "connection refused")
	})

	t.Run("download fails if username is invalid", func(t *testing.T) {
		_, err := DownloadFromHTTP("http://sftp-server:80", "wrong", "test", "abc")
		require.ErrorContains(t, err, "failed to download file: 401 Unauthorized")
	})

	t.Run("download fails if password is wrong", func(t *testing.T) {
		_, err := DownloadFromHTTP("http://sftp-server:80", "test", "wrong", "abc")
		require.ErrorContains(t, err, "failed to download file: 401 Unauthorized")
	})

	t.Run("download fails if file does not exist", func(t *testing.T) {
		_, err := DownloadFromHTTP("http://sftp-server:80", "test", "test", "does-not-exist")
		require.ErrorContains(t, err, "failed to download file: 404 Not Found")
	})
}
