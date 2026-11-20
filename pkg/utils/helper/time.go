package helper

import (
	"camera_api/pkg/api/response"
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

// ConvertTimeRange is like ConvertTime but either bound may be zero, meaning
// "unbounded" on that side. Set bounds are converted to UTC; zero bounds stay
// zero. The location comes from whichever bound is set, or UTC when neither
// is. When both are set they must share a timezone and from must not be after
// to.
func ConvertTimeRange(from, to time.Time) (time.Time, time.Time, *time.Location, error) {
	switch {
	case from.IsZero() && to.IsZero():
		return time.Time{}, time.Time{}, time.UTC, nil

	case from.IsZero() || to.IsZero():
		set := from
		if set.IsZero() {
			set = to
		}

		tz, err := GetTimezone(set)
		if err != nil {
			return time.Time{}, time.Time{}, nil, err
		}

		loc, err := time.LoadLocation(tz)
		if err != nil {
			return time.Time{}, time.Time{}, nil, response.APIError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Can't load timezone location",
			}
		}

		if !from.IsZero() {
			from = from.UTC()
		}

		if !to.IsZero() {
			to = to.UTC()
		}

		return from, to, loc, nil

	default:
		if from.After(to) {
			return time.Time{}, time.Time{}, nil, response.APIError{
				StatusCode: http.StatusBadRequest,
				Message:    "from must not be after to",
			}
		}

		start, end, loc, err := ConvertTime(from, to)
		if err != nil {
			return time.Time{}, time.Time{}, nil, err
		}

		return *start, *end, loc, nil
	}
}
