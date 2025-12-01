package parser

import (
	"fmt"
	"time"
)

var layouts = []string{
	time.RFC1123,
	time.RFC1123Z,
	"Mon 02 Jan 2006 03:04:05 PM MST",
	"Mon Jan 2 15:04:05 MST 2006",
	"Mon Jan 2 03:04:05 PM MST 2006",
	"Mon Jan 2 15:04:05 UTC 2006",
}

func ParseTimeAuto(input string) (time.Time, string, error) {
	for _, layout := range layouts {
		t, err := time.Parse(layout, input)
		if err == nil {
			return t, layout, nil
		}
	}
	return time.Time{}, "", fmt.Errorf("no matching layout found for: %s", input)
}
