package request

import "time"

type Pagination struct {
	Target string
	From   time.Time
	To     time.Time
	Limit  int
	Offset int
	Order  string
}

type SearchContent struct {
	From   time.Time
	To     time.Time
	Limit  int
	Offset int
}
