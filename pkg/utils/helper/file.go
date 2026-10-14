package helper

import (
	"fmt"
	"strings"
	"time"
)

func ParseContentFilename(filename string) (time.Time, time.Time, error) {
	const layout = "20060102T150405Z"

	name := strings.TrimSuffix(filename, ".mp4")

	parts := strings.Split(name, "_")
	if len(parts) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid content filename")
	}

	start, err := time.Parse(layout, parts[0])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start time: %w", err)
	}

	end, err := time.Parse(layout, parts[1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end time: %w", err)
	}

	return start, end, nil
}
