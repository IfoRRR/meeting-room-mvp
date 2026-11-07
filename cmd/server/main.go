package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"meeting-room-mvp/internal/booking"
)

type server struct{ db *pgxpool.Pool }

func main() {
	ctx := context.Background()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://meetings:meetings@localhost:5433/meetings"
	}
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatal("database: ", err)
	}
	if err := applySchema(ctx, db); err != nil {
		log.Fatal("schema: ", err)
	}

	s := &server{db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/rooms", s.rooms)
	mux.HandleFunc("GET /api/bookings", s.bookings)
	mux.HandleFunc("POST /api/bookings", s.createBooking)
	mux.Handle("/", http.FileServer(http.Dir("web")))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	address := ":" + port
	log.Println("meeting-room-mvp listening on http://localhost:" + port)
	log.Fatal(http.ListenAndServe(address, logging(mux)))
}

func applySchema(ctx context.Context, db *pgxpool.Pool) error {
	data, err := os.ReadFile(filepath.Join("internal", "postgres", "schema.sql"))
	if err != nil {
		return err
	}
	if _, err := db.Exec(ctx, string(data)); err != nil {
		return err
	}
	_, err = db.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS btree_gist`)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, `ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_no_overlap`)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, `ALTER TABLE bookings ADD CONSTRAINT bookings_no_overlap EXCLUDE USING gist (room_id WITH =, tstzrange(starts_at, ends_at, '[)') WITH &&)`)
	return err
}

func (s *server) rooms(w http.ResponseWriter, r *http.Request) {
	minCapacity, _ := strconv.Atoi(r.URL.Query().Get("minCapacity"))
	rows, err := s.db.Query(r.Context(), `SELECT id, name, floor, capacity, equipment FROM rooms WHERE capacity >= $1 ORDER BY capacity, name`, minCapacity)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	result := []booking.Room{}
	for rows.Next() {
		var item booking.Room
		if err := rows.Scan(&item.ID, &item.Name, &item.Floor, &item.Capacity, &item.Equipment); err != nil {
			writeError(w, err)
			return
		}
		result = append(result, item)
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *server) bookings(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	rows, err := s.db.Query(r.Context(), `SELECT b.id, b.room_id, r.name, b.title, b.organizer, b.starts_at, b.ends_at FROM bookings b JOIN rooms r ON r.id = b.room_id WHERE b.starts_at < ($1::date + interval '1 day') AND b.ends_at > $1::date ORDER BY b.starts_at`, date)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	result := []booking.Booking{}
	for rows.Next() {
		var item booking.Booking
		if err := rows.Scan(&item.ID, &item.RoomID, &item.RoomName, &item.Title, &item.Organizer, &item.StartsAt, &item.EndsAt); err != nil {
			writeError(w, err)
			return
		}
		result = append(result, item)
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *server) createBooking(w http.ResponseWriter, r *http.Request) {
	var input booking.CreateBooking
	if err := readJSON(r, &input); err != nil {
		writeErrorStatus(w, http.StatusBadRequest, booking.ErrInvalidBooking)
		return
	}
	start, end, err := booking.ValidateCreate(input)
	if err != nil {
		writeErrorStatus(w, http.StatusBadRequest, err)
		return
	}
	var created booking.Booking
	err = s.db.QueryRow(r.Context(), `INSERT INTO bookings (room_id, title, organizer, starts_at, ends_at) VALUES ($1, $2, $3, $4, $5) RETURNING id, room_id, title, organizer, starts_at, ends_at`, input.RoomID, strings.TrimSpace(input.Title), strings.TrimSpace(input.Organizer), start, end).Scan(&created.ID, &created.RoomID, &created.Title, &created.Organizer, &created.StartsAt, &created.EndsAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
			writeErrorStatus(w, http.StatusConflict, errors.New("room is already booked for this interval"))
			return
		}
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func readJSON(r *http.Request, target any) error { return json.NewDecoder(r.Body).Decode(target) }
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, err error) {
	writeErrorStatus(w, http.StatusInternalServerError, err)
}
func writeErrorStatus(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

var _ fs.FS
var _ = pgx.ErrNoRows
var _ = fmt.Sprintf
