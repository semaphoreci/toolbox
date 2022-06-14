package storage

import (
	"io/fs"
	"sort"
	"time"

	"github.com/pkg/sftp"
)

func (s *SFTPStorage) List() ([]CacheKey, error) {
	files, err := s.SFTPClient.ReadDir(".")
	if err != nil {
		return nil, err
	}

	keys := []CacheKey{}
	for _, file := range files {
		storedAt := file.ModTime()
		keys = append(keys, CacheKey{
			Name:           file.Name(),
			Size:           file.Size(),
			StoredAt:       &storedAt,
			LastAccessedAt: findLastAccessedAt(file),
		})
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].LastAccessedAt.After(*keys[j].LastAccessedAt)
	})

	return keys, nil
}

// If we can't figure out the access time of the file,
// we fallback to the modification time.
func findLastAccessedAt(fileInfo fs.FileInfo) *time.Time {
	stat, ok := fileInfo.Sys().(*sftp.FileStat)

	if !ok {
		mtime := fileInfo.ModTime()
		return &mtime
	}

	if stat.Atime == 0 {
		mtime := fileInfo.ModTime()
		return &mtime
	}

	atime := time.Unix(int64(stat.Atime), 0)
	return &atime
}
