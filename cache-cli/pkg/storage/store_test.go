package storage

// import (
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// 	"os/exec"
// 	"runtime"
// 	"strconv"
// 	"strings"
// 	"testing"
// 	"time"

// 	assert "github.com/stretchr/testify/assert"
// )

// func Test__Store(t *testing.T) {
// 	runTestForAllStorageTypes(t, func(storageType string, storage Storage) {
// 		t.Run(fmt.Sprintf("%s stored objects can be listed", storageType), func(t *testing.T) {
// 			_ = storage.Clear()

// 			file, _ := ioutil.TempFile(os.TempDir(), "*")
// 			err := storage.Store("abc001", file.Name())
// 			assert.Nil(t, err)

// 			keys, err := storage.List()
// 			assert.Nil(t, err)

// 			if assert.Len(t, keys, 1) {
// 				key := keys[0]
// 				assert.Equal(t, key.Name, "abc001")
// 				assert.NotNil(t, key.StoredAt)
// 				assert.NotNil(t, key.Size)
// 			}

// 			os.Remove(file.Name())
// 		})

// 		t.Run(fmt.Sprintf("%s stored objects can be restored", storageType), func(t *testing.T) {
// 			_ = storage.Clear()

// 			file, _ := ioutil.TempFile(os.TempDir(), "*")
// 			file.WriteString("stored objects can be restored")

// 			err := storage.Store("abc002", file.Name())
// 			assert.Nil(t, err)

// 			restoredFile, err := storage.Restore("abc002")
// 			assert.Nil(t, err)

// 			content, err := ioutil.ReadFile(restoredFile.Name())
// 			assert.Nil(t, err)
// 			assert.Equal(t, "stored objects can be restored", string(content))

// 			os.Remove(file.Name())
// 			os.Remove(restoredFile.Name())
// 		})

// 		/*
// 		 * To assert that concurrent writes do not lead to the remote file having bytes from all files being
// 		 * concurrently uploaded, we create two big files and upload them concurrently. The files need to be
// 		 * big because we need to make sure both uploads happen at the same time. Each file has a different string per line.
// 		 *
// 		 * To assert that only the bytes from the bigger file (the one that finishes writing last) are the ones
// 		 * that end up being used for the remote file, we look at the remote file and check that it doesn't
// 		 * have any lines from the smaller file.
// 		 */
// 		t.Run(fmt.Sprintf("%s concurrent writes keep the file that finished writing last", storageType), func(t *testing.T) {
// 			if runtime.GOOS == "windows" {
// 				t.Skip()
// 			}

// 			_ = storage.Clear()

// 			smallerFile := fmt.Sprintf("%s/smaller.tmp", os.TempDir())
// 			err := createBigTempFile(smallerFile, 300*1000*1000) // 300M
// 			assert.Nil(t, err)

// 			// this one is bigger so it will take longer to finish
// 			biggerFile := fmt.Sprintf("%s/bigger.tmp", os.TempDir())
// 			err = createBigTempFile(biggerFile, 600*1000*1000) // 600M
// 			assert.Nil(t, err)

// 			go func() {
// 				_ = storage.Store("abc003", smallerFile)
// 			}()

// 			_ = storage.Store("abc003", biggerFile)

// 			restoredFile, err := storage.Restore("abc003")
// 			assert.Nil(t, err)
// 			assert.Zero(t, countLines(restoredFile.Name(), smallerFile))

// 			os.Remove(smallerFile)
// 			os.Remove(biggerFile)
// 			os.Remove(restoredFile.Name())
// 		})
// 	})

// 	if runtime.GOOS != "windows" {
// 		runTestForSingleStorageType("sftp", 1024, t, func(storage Storage) {
// 			t.Run("sftp storage deletes old keys if no space left to store", func(t *testing.T) {
// 				_ = storage.Clear()

// 				file1, _ := ioutil.TempFile(os.TempDir(), "*")
// 				file1.WriteString(strings.Repeat("x", 400))
// 				storage.Store("abc001", file1.Name())

// 				time.Sleep(time.Second)

// 				file2, _ := ioutil.TempFile(os.TempDir(), "*")
// 				file2.WriteString(strings.Repeat("x", 400))
// 				storage.Store("abc002", file2.Name())

// 				time.Sleep(time.Second)

// 				file3, _ := ioutil.TempFile(os.TempDir(), "*")
// 				file3.WriteString(strings.Repeat("x", 400))
// 				storage.Store("abc003", file3.Name())

// 				keys, _ := storage.List()
// 				assert.Len(t, keys, 2)

// 				firstKey := keys[0]
// 				assert.Equal(t, "abc003", firstKey.Name)
// 				secondKey := keys[1]
// 				assert.Equal(t, "abc002", secondKey.Name)

// 				os.Remove(file1.Name())
// 				os.Remove(file2.Name())
// 				os.Remove(file3.Name())
// 			})
// 		})
// 	}
// }

// func createBigTempFile(fileName string, size int64) error {
// 	command := fmt.Sprintf("yes '%s' | head -c %d > %s", fileName, size, fileName)
// 	cmd := exec.Command("bash", "-c", command)
// 	return cmd.Run()
// }

// func countLines(fileName, line string) int64 {
// 	command := fmt.Sprintf("cat %s | grep '%s' | wc -l", fileName, line)
// 	cmd := exec.Command("bash", "-c", command)
// 	output, err := cmd.Output()
// 	if err != nil {
// 		return -1
// 	}

// 	count := strings.TrimSuffix(string(output), "\n")
// 	value, err := strconv.ParseInt(count, 10, 64)
// 	if err != nil {
// 		return -1
// 	}

// 	return value
// }
