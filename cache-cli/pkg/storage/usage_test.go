package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test__Usage(t *testing.T) {
	runTestForAllStorageTypes(t, func(storageType string, storage Storage) {
		t.Run(fmt.Sprintf("%s no usage", storageType), func(t *testing.T) {
			_ = storage.Clear()
			usage, err := storage.Usage()
			assert.Nil(t, err)
			assert.Equal(t, int64(0), usage.Used)

			switch storageType {
			case "s3":
				assert.Equal(t, int64(-1), usage.Free)
			case "sftp":
				assert.Equal(t, storage.Config().MaxSpace, usage.Free)
			}
		})

		t.Run(fmt.Sprintf("%s some usage", storageType), func(t *testing.T) {
			_ = storage.Clear()

			fileContents := "usage - some usage"
			file, _ := ioutil.TempFile(os.TempDir(), "*")
			file.WriteString(fileContents)
			_ = storage.Store("abc001", file.Name())

			usage, err := storage.Usage()
			assert.Nil(t, err)
			assert.Equal(t, int64(len(fileContents)), usage.Used)

			switch storageType {
			case "s3":
				assert.Equal(t, int64(-1), usage.Free)
			case "sftp":
				free := storage.Config().MaxSpace - int64(len(fileContents))
				assert.Equal(t, free, usage.Free)
			}

			os.Remove(file.Name())
		})
	})
}
