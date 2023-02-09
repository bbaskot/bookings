package repository

import (
	"time"

	"github.com/atom91/bookings/internal/models"
)

type DatabaseRepo interface {
	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(r *models.RoomRestriction) error
	SearchAvailabilityByDatesByRoomId(start, end time.Time,roomId int)(bool, error)
	SearchAvailabilityOfAllRooms(start, end time.Time)([]models.Room, error)
	SearchRoomById(id int)models.Room
	GetUserById(id int)(models.User, error)
	UpdateUser(u models.User)error
	Authenticate(email, testPassword string)(int, string, error)
	AllReservations() ([]models.Reservation, error)
	NewReservations() ([]models.Reservation, error)
	GetReservationById(id int)(models.Reservation, error)
	UpdateReservation(u models.Reservation)error
	DeleteReservation(id int)error
	UpdateProcessed(id,processed int)error
	AllRooms() ([]models.Room, error)
	GetRestrictionsForRoomByDate(roomId int, start, end time.Time) ([]models.RoomRestriction, error)
	DeleteBlocks(id int)error
	InsertBlock(date time.Time, roomId int) (error)
}
