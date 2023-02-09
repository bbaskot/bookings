package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/atom91/bookings/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDbRepo) AllUsers() bool {
	return true
}
func (m *postgresDbRepo) InsertReservation(res models.Reservation) (int, error){
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	var newId int;
	stmt:= `INSERT INTO reservations
	(first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`
	err:= m.DB.QueryRowContext(ctx,stmt,
	res.FirstName,
	res.LastName,
	res.Email,
	res.Phone,
	res.StartDate,
	res.EndDate,
	res.RoomId,
	time.Now(),
	time.Now()).Scan(&newId)
	if err!=nil{
		return 0,err
	}

	return newId ,nil
}

func (m *postgresDbRepo)InsertRoomRestriction(r *models.RoomRestriction)error{
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	stmt:=`INSERT INTO room_restrictions
	(start_date, end_date, room_id, reservation_id, restriction_id, created_at, updated_at)
	values ($1, $2, $3, $4, $5, $6, $7)`
	_,err:=m.DB.ExecContext(ctx,stmt,
	r.StartDate,
	r.EndDate,
	r.RoomId,
	r.ReservationId,
	r.RestrictionId,
	time.Now(),
	time.Now(),
	)
	if err!=nil {
		return err
	}
	
	return nil
}
func (m *postgresDbRepo)SearchAvailabilityByDatesByRoomId(start, end time.Time,roomId int)(bool, error){
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	query:=`SELECT
		COUNT(id)
	FROM
		room_restrictions
	WHERE
		$1<end_date and $2>start_date and room_id=$3`
	var numRows int
	err:=m.DB.QueryRowContext(ctx,query,start,end,roomId).Scan(&numRows)
	if err!=nil{
		return false,err
	}
	if numRows==0{
		return true,nil
	}
	return false,nil
	
}
func (m *postgresDbRepo)SearchRoomById(id int)models.Room{
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	var room models.Room
	query:=`
			SELECT 
				r.id, r.room_name
			FROM
				rooms r
			WHERE
				r.id=$1
				`
	m.DB.QueryRowContext(ctx,query,id).Scan(&room.ID,&room.RoomName)
	return room
}
// returns a slice of available rooms for given date range
func (m *postgresDbRepo)SearchAvailabilityOfAllRooms(start, end time.Time)([]models.Room, error){
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	var rooms []models.Room
	query:=`
			SELECT 
				r.id, r.room_name
			FROM
				rooms r
			WHERE
				r.id NOT IN (SELECT rr.room_id FROM room_restrictions rr WHERE $1<rr.end_date AND $2>rr.start_date)
				`
	rows,err:=m.DB.QueryContext(ctx,query,start,end)
	if err!=nil{
		return rooms,err
	}
	for rows.Next(){
		var room models.Room
		err:=rows.Scan(
			&room.ID,
			&room.RoomName,
		)
		if err!=nil{
			return rooms, err
		}
		rooms=append(rooms,room)
	}
	return rooms,nil

}
func (m *postgresDbRepo)GetUserById(id int)(models.User, error){
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	query:= "SELECT id, first_name, last_name, email, password, access_level, created_at, updated_at FROM users WHERE id=$1"

	row:=m.DB.QueryRowContext(ctx,query,id)
	var u models.User
	err:= row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err!=nil{
		return u, err
	}
	return u, nil


}

func(m *postgresDbRepo) UpdateUser(u models.User)error{

	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	query:=`UPDATE users SET  first_name=$1, last_name=$2, email=$3, access_level=$4, updated_at=$5 
	WHERE 
	id=$6
	`
	_,err:=m.DB.ExecContext(ctx,query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.AccessLevel,
		time.Now(),
		u.ID,
	)
	if err!=nil{
		return err
	}
	return nil

}
//Authenticates a user
func (m *postgresDbRepo) Authenticate(email, testPassword string)(int, string, error){
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	var id int
	var hashedPassword string
	row :=m.DB.QueryRowContext(ctx,"SELECT id, password FROM users WHERE email= $1 ",email)
	err:=row.Scan(&id,&hashedPassword)
	if err!=nil {
		
		return id,"",err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword),[]byte(testPassword))
	if err==bcrypt.ErrMismatchedHashAndPassword{
		return 0,"",errors.New("incorrect password")
	}else if err!=nil{
		return 0,"",err
	}
	return id, hashedPassword,nil

}

func (m *postgresDbRepo) AllReservations() ([]models.Reservation, error){
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	var reservations []models.Reservation

	query:=`SELECT r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, r.processed, rm.id, rm.room_name
			FROM reservations r
			JOIN rooms rm ON (r.room_id=rm.id)
			ORDER BY r.start_date ASC
	`
	rows, err:=m.DB.QueryContext(ctx,query)
	if err!=nil{
		return reservations, err
	}
	for rows.Next(){
		var i models.Reservation
		err:=rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomId,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Processed,
			&i.Room.ID,
			&i.Room.RoomName,
		)
		if err!=nil{
			return reservations,err
		}
		reservations = append(reservations, i)
	}
	if err!=nil{
		return reservations, err
	}
	defer rows.Close()
	return reservations, nil
}

func (m *postgresDbRepo) NewReservations() ([]models.Reservation, error){
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	var reservations []models.Reservation

	query:=`SELECT r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, r.processed, rm.id, rm.room_name
			FROM reservations r
			JOIN rooms rm ON (r.room_id=rm.id)
			WHERE r.processed=0
			ORDER BY r.start_date ASC
	`
	rows, err:=m.DB.QueryContext(ctx,query)
	if err!=nil{
		return reservations, err
	}
	for rows.Next(){
		var i models.Reservation
		err:=rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomId,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Processed,
			&i.Room.ID,
			&i.Room.RoomName,
		)
		if err!=nil{
			return reservations,err
		}
		reservations = append(reservations, i)
	}
	if err!=nil{
		return reservations, err
	}
	defer rows.Close()
	return reservations, nil
}
func (m *postgresDbRepo)GetReservationById(id int)(models.Reservation, error){
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	var i models.Reservation
	query:=`SELECT r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, r.processed, rm.id, rm.room_name
			FROM reservations r
			JOIN rooms rm ON (r.room_id=rm.id)
			WHERE r.id=$1
	`
	row :=m.DB.QueryRowContext(ctx,query,id)
	err:=row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Email,
		&i.Phone,
		&i.StartDate,
		&i.EndDate,
		&i.RoomId,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Processed,
		&i.Room.ID,
		&i.Room.RoomName,
	)
	if err!=nil{
		return i, err
	}
	return i, nil
}

func(m *postgresDbRepo) UpdateReservation(u models.Reservation)error{

	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	query:=`UPDATE reservations SET  first_name=$1, last_name=$2, email=$3, phone=$4, updated_at=$5 
	WHERE 
	id=$6
	`
	_,err:=m.DB.ExecContext(ctx,query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.Phone,
		time.Now(),
		u.ID,
	)
	if err!=nil{
		return err
	}
	return nil

}
func (m *postgresDbRepo) DeleteReservation(id int)error{
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	query:=`DELETE FROM reservations WHERE id=$1`
	_,err:=m.DB.ExecContext(ctx,query,id)
	if err!=nil{
		return err
	}
	return nil

}
func (m *postgresDbRepo) UpdateProcessed(id,processed int)error{
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	query:=`UPDATE reservations SET processed=$1 WHERE id=$2`
	_,err:=m.DB.ExecContext(ctx,query,processed,id)
	if err!=nil{
		return err
	}
	return nil
}

func (m *postgresDbRepo) AllRooms() ([]models.Room, error){
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	var rooms []models.Room

	query:=`SELECT id, room_name, created_at, updated_at from rooms order by room_name
	`
	rows, err:=m.DB.QueryContext(ctx,query)
	if err!=nil{
		return rooms, err
	}
	for rows.Next(){
		var i models.Room
		err:=rows.Scan(
			&i.ID,
			&i.RoomName,
			&i.CreatedAt,
			&i.UpdatedAt,
		)
		if err!=nil{
			return rooms,err
		}
		rooms = append(rooms, i)
	}
	if err!=nil{
		return rooms, err
	}
	defer rows.Close()
	return rooms, nil
}
func (m *postgresDbRepo) GetRestrictionsForRoomByDate(roomId int, start, end time.Time) ([]models.RoomRestriction, error){
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	var restrictions []models.RoomRestriction

	query:=`SELECT id, COALESCE(reservation_id,0) , restriction_id, room_id, start_date, end_date
			FROM room_restrictions
			WHERE $1<=end_date AND $2>=start_date AND $3=room_id
			ORDER BY start_date ASC
	`
	rows, err:=m.DB.QueryContext(ctx,query,start,end,roomId)
	if err!=nil{
		return restrictions, err
	}
	for rows.Next(){
		var i models.RoomRestriction
		err:=rows.Scan(
			&i.ID,
			&i.ReservationId,
			&i.RestrictionId,
			&i.RoomId,
			&i.StartDate,
			&i.EndDate,
		)
		if err!=nil{
			return restrictions,err
		}
		restrictions = append(restrictions, i)
	}
	if err!=nil{
		return restrictions, err
	}
	defer rows.Close()
	return restrictions, nil
}

func (m *postgresDbRepo) DeleteBlocks(id int)error{
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	query:=`DELETE FROM room_restrictions WHERE id=$1`
	_,err:=m.DB.ExecContext(ctx,query,id)
	if err!=nil{
		return err
	}
	return nil

}
func (m *postgresDbRepo) InsertBlock(date time.Time, roomId int) (error){
	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	
	stmt:= `INSERT INTO room_restrictions
	(room_id, restriction_id, start_date, end_date, created_at, updated_at)
	values ($1, 2, $2, $3, $4, $5)`
	_,err:= m.DB.ExecContext(ctx,stmt,
	roomId,
	date,
	date,	
	time.Now(),
	time.Now())
	if err!=nil{
		return err
	}

	return nil
}