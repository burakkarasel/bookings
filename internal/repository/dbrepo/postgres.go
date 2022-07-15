package dbrepo

import (
	"context"
	"github.com/burakkarasel/bookings/internal/models"
	"time"
)

func (repo *postgresDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation into database
func (repo *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// here we created a context to cancel this func with a timeout of 3 seconds, because we don't want it to run
	// 5 minutes as we specified in our driver package
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newID int

	statement := `insert into reservations (first_name, last_name, email, phone, start_date, end_date, 
                          room_id, created_at, updated_at)
                          values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`
	err := repo.DB.QueryRowContext(ctx, statement,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// InsertRoomRestriction insert a new room restriction to DB
func (repo *postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	statement := `insert into room_restrictions (start_date, end_date, room_id, reservation_id,
                    created_at, updated_at, restriction_id)
					values($1, $2, $3, $4, $5, $6, $7)`

	_, err := repo.DB.ExecContext(ctx, statement,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		time.Now(),
		time.Now(),
		r.RestrictionID,
	)

	if err != nil {
		return err
	}

	return nil
}

// SearchAvailabilityByDatesByRoomID returns true if availability exist for roomID and false if no availability exist
func (repo *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var numRows int

	query := `
		select 
			count(id)
		from 
		    room_restrictions
		where 
		    room_id = $1 and
		    $2 < end_date and $3 > start_date;
		`
	row := repo.DB.QueryRowContext(ctx, query, roomID, start, end)

	err := row.Scan(&numRows)

	if err != nil {
		return false, err
	}

	if numRows == 0 {
		return true, nil
	}

	return false, nil
}

// SearchAvailabilityForAllRooms checks for all rooms restriction's in a given period of time and returns available rooms
func (repo *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var roomSlc []models.Room

	query := `
			select 
				r.id, r.room_name
			from 
				rooms r 
			where 
				r.id not in (
								select 
									rr.room_id 
								from 
									room_restrictions rr
								where 
								$1 > rr.start_date and $2 < rr.end_date 
							)
			`
	rows, err := repo.DB.QueryContext(ctx, query, start, end)

	if err != nil {
		return roomSlc, err
	}

	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID, &room.RoomName)
		if err != nil {
			return roomSlc, err
		}
		roomSlc = append(roomSlc, room)
	}

	if err != nil {
		return roomSlc, err
	}

	return roomSlc, nil
}

// GetRoomById takes only one argument ID and returns the relevant room data
func (repo *postgresDBRepo) GetRoomById(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `
			select id, room_name, created_at, updated_at
			from rooms
			where id = $1
			`
	row := repo.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(&room.ID, &room.RoomName, &room.CreatedAt, &room.UpdatedAt)

	if err != nil {
		return room, err
	}

	return room, nil
}
