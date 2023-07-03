package common

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// TruncateTimeToMs takes a time and truncates it to millisecond precision.
func TruncateTimeToMs(tm time.Time) time.Time {
	return tm.Truncate(time.Millisecond)
}

// PgErrToCommon converts a raw Postgres error code value to a known error code.
func PgErrToCommon(err error) string {
	code := ErrcdRepoOther // default

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		code = pgErr.Code
	}

	switch {
	case code == "23505":
		return ErrcdRepoDupeRow
	case code[0:2] == "08":
		return ErrcdRepoConnException
	default:
		return ErrcdRepoOther
	}
}

// matchTime helps go-sqlmock match time.Time data.
type MatchTime struct {
	Value time.Time
}

// Match compares the value it receives as a time to the time in the receiver and returns true if they are the same time.
//
// Mutates receiver: no (read only)
func (mt MatchTime) Match(v driver.Value) bool {
	v1, ok := v.(time.Time)
	if ok && v1.Compare(mt.Value) == 0 {
		return true
	}
	return false
}
