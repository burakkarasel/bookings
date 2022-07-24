package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/burakkarasel/bookings/internal/models"
	"golang.org/x/crypto/bcrypt"
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
		    $2 <= end_date and $3 >= start_date;
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
								$1 <= rr.end_date and $2 >= rr.start_date 
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

	if err = rows.Err(); err != nil {
		return roomSlc, err
	}

	return roomSlc, nil
}

// GetRoomById takes only one argument ID and returns the relevant room's data
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

// GetUserById returns a user from DB by id
func (repo *postgresDBRepo) GetUserById(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var u models.User

	query := `
			select id, first_name, last_name, email, password, access_level, created_at, updated_at
			from users
			where id = $1
	`

	row := repo.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Password, &u.AccessLevel, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		return u, err
	}

	return u, nil
}

// UpdateUser updates a user in DB
func (repo *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update users set first_name = $1, last_name = $2, email = $3, access_level = $4, updated_at = $5
		where id = $6
	`

	_, err := repo.DB.ExecContext(ctx, query, u.FirstName, u.LastName, u.Email, u.AccessLevel, time.Now(), u.ID)

	if err != nil {
		return err
	}

	return nil
}

// Authenticate authenticates a user
func (repo *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	// first we check if email is valid
	row := repo.DB.QueryRowContext(ctx, "select id, password from users where email = $1", email)
	err := row.Scan(&id, &hashedPassword)

	if err != nil {
		return id, "", err
	}

	// now we check if the password given valid for the email or not
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))

	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}

// AllReservations returns all of the reservations from DB as a slice
func (repo *postgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, r.processed, rm.id, rm.room_name
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		order by r.start_date asc
	`

	rows, err := repo.DB.QueryContext(ctx, query)

	if err != nil {
		return reservations, err
	}

	defer rows.Close()

	for rows.Next() {
		var reservation models.Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.FirstName,
			&reservation.LastName,
			&reservation.Email,
			&reservation.Phone,
			&reservation.StartDate,
			&reservation.EndDate,
			&reservation.RoomID,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
			&reservation.Processed,
			&reservation.Room.ID,
			&reservation.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, reservation)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// AllNewReservations returns all of the new reservations from DB as a slice
func (repo *postgresDBRepo) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, r.processed, rm.id, rm.room_name
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		where r.processed = 0
		order by r.start_date asc
	`

	rows, err := repo.DB.QueryContext(ctx, query)

	if err != nil {
		return reservations, err
	}

	defer rows.Close()

	for rows.Next() {
		var reservation models.Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.FirstName,
			&reservation.LastName,
			&reservation.Email,
			&reservation.Phone,
			&reservation.StartDate,
			&reservation.EndDate,
			&reservation.RoomID,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
			&reservation.Processed,
			&reservation.Room.ID,
			&reservation.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, reservation)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// GetReservationById returns a reservation according to id
func (repo *postgresDBRepo) GetReservationById(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservation models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
			rm.id, rm.room_name
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		where r.id = $1
	`

	row := repo.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&reservation.ID,
		&reservation.FirstName,
		&reservation.LastName,
		&reservation.Email,
		&reservation.Phone,
		&reservation.StartDate,
		&reservation.EndDate,
		&reservation.RoomID,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
		&reservation.Processed,
		&reservation.Room.ID,
		&reservation.Room.RoomName,
	)

	if err != nil {
		return reservation, err
	}

	return reservation, nil
}

// UpdateUser updates a user in DB
func (repo *postgresDBRepo) UpdateReservation(r models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update reservations set first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = $5
		where id = $6
	`

	_, err := repo.DB.ExecContext(ctx, query, r.FirstName, r.LastName, r.Email, r.Phone, time.Now(), r.ID)

	if err != nil {
		return err
	}

	return nil
}

// DeleteReservation deletes a reservation from database by id
func (repo *postgresDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		delete
		from reservations
		where id = $1
	`

	_, err := repo.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	return nil
}

// UpdateProcessedForReservation updates processed for a reservation by id
func (repo *postgresDBRepo) UpdateProcessedForReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update reservations set processed = $1
		where id = $2
	`

	_, err := repo.DB.ExecContext(ctx, query, processed, id)

	if err != nil {
		return err
	}

	return nil
}

// AllRooms return the rooms from the DB
func (repo *postgresDBRepo) AllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `select id, room_name, created_at, updated_at from rooms order by room_name`

	rows, err := repo.DB.QueryContext(ctx, query)

	if err != nil {
		return rooms, err
	}

	defer rows.Close()

	for rows.Next() {
		var room models.Room

		err := rows.Scan(
			&room.ID,
			&room.RoomName,
			&room.CreatedAt,
			&room.UpdatedAt,
		)

		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// GetRestrictionForRoomByDate returns if a room for given date is available or not
func (repo *postgresDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restrictions []models.RoomRestriction

	// here we used coalesce for null reservation ids if reservation id is null it returns 0
	query := `
		select id, coalesce(reservation_id, 0), restriction_id, room_id, start_date, end_date
		from room_restrictions
		where $1 < end_date and $2 >= start_date and room_id = $3
	`

	rows, err := repo.DB.QueryContext(ctx, query, start, end, roomID)

	if err != nil {
		return restrictions, err
	}

	defer rows.Close()

	for rows.Next() {
		var r models.RoomRestriction

		err := rows.Scan(
			&r.ID,
			&r.ReservationID,
			&r.RestrictionID,
			&r.RoomID,
			&r.StartDate,
			&r.EndDate,
		)

		if err != nil {
			return restrictions, err
		}

		restrictions = append(restrictions, r)
	}

	if err = rows.Err(); err != nil {
		return restrictions, err
	}

	return restrictions, nil
}

// InsertBlockForRoom inserts a new block for a given room in DB
func (repo *postgresDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		insert into room_restrictions (start_date, end_date, room_id, restriction_id, created_at, updated_at)
		values($1, $2, $3, $4, $5, $6)
	`

	_, err := repo.DB.ExecContext(ctx, query,
		startDate,
		startDate.AddDate(0, 0, 1),
		id,
		2,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

// RemoveBlockForRoom removes the block for given room restriction
func (repo *postgresDBRepo) RemoveBlockForRoom(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `delete from room_restrictions where id = $1`

	_, err := repo.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	return nil
}
