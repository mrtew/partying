package repository

import (
	"database/sql"
	"errors"
	"partying/internal/model"
)

type RoomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) Create(room *model.Room) error {
	query := `INSERT INTO rooms (id, name, host_id, max_capacity, created_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, room.ID, room.Name, room.HostID, room.MaxCapacity, room.CreatedAt)
	return err
}

func (r *RoomRepository) FindAll() ([]*model.Room, error) {
	query := `SELECT id, name, host_id, max_capacity, created_at FROM rooms`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*model.Room
	for rows.Next() {
		room := &model.Room{}
		if err := rows.Scan(&room.ID, &room.Name, &room.HostID, &room.MaxCapacity, &room.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *RoomRepository) FindByID(id string) (*model.Room, error) {
	query := `SELECT id, name, host_id, max_capacity, created_at FROM rooms WHERE id = ?`
	row := r.db.QueryRow(query, id)

	room := &model.Room{}
	err := row.Scan(&room.ID, &room.Name, &room.HostID, &room.MaxCapacity, &room.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("room not found")
	}
	if err != nil {
		return nil, err
	}
	return room, nil
}
