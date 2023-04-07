package repository

import (
	"github.com/JustAPotato0916/bookings/internal/models"
	"time"
)

type DatabaseRepo interface {
	AllUsers() bool

	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(r models.RoomRestriction)
	SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error)
}
