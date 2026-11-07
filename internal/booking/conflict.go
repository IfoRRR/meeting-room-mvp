package booking

import "time"

// HasConflict returns true when two bookings overlap in the same room.
// Meetings that end exactly when another one starts do not conflict.
func HasConflict(roomID int64, startsAt, endsAt time.Time, existing Booking) bool {
	if roomID != existing.RoomID {
		return false
	}

	return startsAt.Before(existing.EndsAt) && existing.StartsAt.Before(endsAt)
}
