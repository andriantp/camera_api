package helper

import (
	"camera_api/pkg/utils/response"
	"net/http"
	"time"
)

func ConvertTime(start, end time.Time) (*time.Time, *time.Time, *time.Location, error) {
	tz1, err := GetTimezone(start)
	if err != nil {
		return nil, nil, nil, err
	}

	tz2, err := GetTimezone(end)
	if err != nil {
		return nil, nil, nil, err
	}

	if tz1 != tz2 {
		return nil, nil, nil, response.APIError{
			StatusCode: http.StatusConflict,
			Message:    "Start and end timezone must match",
		}
	}

	loc, err := time.LoadLocation(tz1)
	if err != nil {
		return nil, nil, nil, response.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Can't load timezone location",
		}
	}

	newStart := start.UTC()
	newEnd := end.UTC()

	return &newStart, &newEnd, loc, nil
}

func GetTimezone(timestamp time.Time) (string, error) {
	_, offset := timestamp.Zone()

	switch offset {
	case 7 * 3600:
		return "Asia/Jakarta", nil
	case 8 * 3600:
		return "Asia/Makassar", nil
	case 9 * 3600:
		return "Asia/Jayapura", nil
	case 0:
		return "UTC", nil
	default:
		return "", response.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Unsupported timezone",
		}
	}
}
