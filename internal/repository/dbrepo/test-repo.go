package dbrepo

import (
	"errors"
	"github.com/burakkarasel/bookings/internal/models"
	"time"
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