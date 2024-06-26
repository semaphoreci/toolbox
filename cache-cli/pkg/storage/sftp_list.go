package storage

import (
	"context"
	"io/fs"
	"sort"
	"time"

	"github.com/pkg/sftp"
)

func (s *SFTPStorage) List(ctx context.Context) ([]CacheKey, error) {
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

	return s.sortKeys(keys), nil
}

func (s *SFTPStorage) sortKeys(keys []CacheKey) []CacheKey {
	switch s.Config().SortKeysBy {
	case SortBySize:
		sort.SliceStable(keys, func(i, j int) bool {
			return keys[i].Size > keys[j].Size
		})
	case SortByAccessTime:
		sort.SliceStable(keys, func(i, j int) bool {
			return keys[i].LastAccessedAt.After(*keys[j].LastAccessedAt)
		})
	default:
		sort.SliceStable(keys, func(i, j int) bool {
			return keys[i].StoredAt.After(*keys[j].StoredAt)
		})
	}

	return keys
}

// If we can't figure out the access time of the file,
// we fallback to the modification time.
func findLastAccessedAt(fileInfo fs.FileInfo) *time.Time {
	mtime := fileInfo.ModTime()

	// Try to get the underlying data source; if nil, fallback to mtime.
	ds := fileInfo.Sys()
	if ds == nil {
		return &mtime
	}

	// Try to cast the underlying data source to something we understand; if nil, fallback to mtime.
	stat, ok := ds.(*sftp.FileStat)
	if !ok {
		return &mtime
	}

	// atime can also be unset; fallback to mtime.
	if stat.Atime == 0 {
		mtime := fileInfo.ModTime()
		return &mtime
	}

	// Otherwise, we use atime.
	atime := time.Unix(int64(stat.Atime), 0)
	return &atime
}
