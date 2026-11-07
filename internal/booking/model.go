package booking

import "time"

type Room struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Floor     string   `json:"floor"`
	Capacity  int      `json:"capacity"`
	Equipment []string `json:"equipment"`
}

type Booking struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"roomId"`
	RoomName  string    `json:"roomName,omitempty"`
	Title     string    `json:"title"`
	Organizer string    `json:"organizer"`
	StartsAt  time.Time `json:"startsAt"`
	EndsAt    time.Time `json:"endsAt"`
}

type CreateBooking struct {
	RoomID    int64  `json:"roomId"`
	Title     string `json:"title"`
	Organizer string `json:"organizer"`
	StartsAt  string `json:"startsAt"`
	EndsAt    string `json:"endsAt"`
}
