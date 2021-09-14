package storage

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (s *LFTPStorage) List() ([]CacheKey, error) {
	output, err := s.ExecuteCommand("cls --sort=date -l")
	if err != nil {
		fmt.Printf("Error executing command: %v", err)
		return nil, err
	}

	lines := filterEmpty(strings.Split(output, "\n"))
	keys := []CacheKey{}

	for _, line := range lines {
		key, err := lineToCacheKey(line)
		if err != nil {
			return nil, err
		}

		keys = append(keys, *key)
	}

	return keys, nil
}

func lineToCacheKey(line string) (*CacheKey, error) {
	fields := filterEmpty(strings.Split(line, " "))

	if len(fields) != 9 {
		return nil, fmt.Errorf("unrecognized number of fields %d", len(fields))
	}

	size, err := strconv.ParseInt(fields[4], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing file size %s: %v", fields[4], err)
	}

	date := fmt.Sprintf("%s %s %s", fields[5], fields[6], fields[7])
	storedAt, err := time.Parse("Jan 2 15:04", date)
	if err != nil {
		return nil, fmt.Errorf("error parsing file date %s: %v", date, err)
	}

	return &CacheKey{
		Name:     fields[8],
		Size:     size,
		StoredAt: &storedAt,
	}, nil
}

func filterEmpty(fields []string) []string {
	filteredFields := []string{}
	for _, field := range fields {
		if field != "" {
			filteredFields = append(filteredFields, field)
		}
	}

	return filteredFields
}
