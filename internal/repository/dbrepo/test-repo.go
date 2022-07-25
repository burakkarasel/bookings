package dbrepo

import (
	"errors"
	"time"

	"github.com/burakkarasel/bookings/internal/models"
)

// for now i only need this functions to exist, so I can make my unit test with other packages

func (repo *testDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation into database
func (repo *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// here we created a context to cancel this func with a timeout of 3 seconds, because we don't want it to run
	// 5 minutes as we specified in our driver package
	if res.RoomID == 2 {
		return 0, errors.New("some error")
	}

	return 1, nil
}

// InsertRoomRestriction insert a new room restriction to DB
func (repo *testDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	if r.RoomID == 0 {
		return errors.New("some error")
	}
	return nil
}

// SearchAvailabilityByDatesByRoomID returns true if availability exist for roomID and false if no availability exist
func (repo *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	if roomID == 17 {
		return false, errors.New("some error")
	}

	return true, nil
}

// SearchAvailabilityForAllRooms checks for all rooms restriction's in a given period of time and returns available rooms
func (repo *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	if start.Format("2006-01-02") == "2023-02-19" {
		return []models.Room{}, errors.New("some error")
	}
	return []models.Room{{RoomName: "general's quarter", ID: 1}}, nil
}

// GetRoomById takes only one argument ID and returns the relevant room's data
func (repo *testDBRepo) GetRoomById(id int) (models.Room, error) {
	var room models.Room
	if id > 2 {
		return room, errors.New("some error")
	}
	return room, nil
}

// GetUserById gets user from DB by id
func (repo *testDBRepo) GetUserById(id int) (models.User, error) {
	return models.User{}, nil
}

// UpdateUser updates a user
func (repo *testDBRepo) UpdateUser(u models.User) error {
	return nil
}

// Authenticate authenticates a user
func (repo *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	if email == "jack@nimble.com" {
		return 0, "", errors.New("some error")
	}
	return 0, "", nil
}

func (repo *testDBRepo) AllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation
	return reservations, nil
}

func (repo *testDBRepo) AllNewReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation
	return reservations, nil
}

func (repo *testDBRepo) GetReservationById(id int) (models.Reservation, error) {
	var reservation models.Reservation
	return reservation, nil
}

func (repo *testDBRepo) UpdateReservation(r models.Reservation) error {
	return nil
}

func (repo *testDBRepo) DeleteReservation(id int) error {
	return nil
}

func (repo *testDBRepo) UpdateProcessedForReservation(id, processed int) error {

	return nil
}

// AllRooms return the rooms from the DB
func (repo *testDBRepo) AllRooms() ([]models.Room, error) {
	var rooms []models.Room
	return rooms, nil
}

// GetRestrictionForRoomByDate returns if a room for given date is available or not
func (repo *testDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {

	var restrictions []models.RoomRestriction
	return restrictions, nil
}

// InsertBlockForRoom inserts a new block for a given room in DB
func (repo *testDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {
	return nil
}

// RemoveBlockForRoom removes the block for given room restriction
func (repo *testDBRepo) RemoveBlockForRoom(id int) error {
	return nil
}
