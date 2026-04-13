package repo

import (
	"errors"
	"strconv"
)

// ErrNotFound is returned when a query finds no matching row.
var ErrNotFound = errors.New("not found")

func itoa(i int) string { return strconv.Itoa(i) }
