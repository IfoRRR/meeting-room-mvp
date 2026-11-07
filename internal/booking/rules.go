package booking

import (
	"errors"
	"strings"
	"time"
)

var ErrInvalidBooking = errors.New("invalid booking")

func ValidateCreate(input CreateBooking) (time.Time, time.Time, error) {
	if input.RoomID <= 0 || strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Organizer) == "" {
		return time.Time{}, time.Time{}, ErrInvalidBooking
	}
	start, err := time.Parse(time.RFC3339, input.StartsAt)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidBooking
	}
	end, err := time.Parse(time.RFC3339, input.EndsAt)
	if err != nil || !end.After(start) {
		return time.Time{}, time.Time{}, ErrInvalidBooking
	}
	return start.UTC(), end.UTC(), nil
}
