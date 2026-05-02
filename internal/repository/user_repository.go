package repository

import (
	"database/sql"
	"errors"
	"partying/internal/model"

	"github.com/go-sql-driver/mysql"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	query := `INSERT INTO users (id, username, password, created_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, user.ID, user.Username, user.Password, user.CreatedAt)
	if err != nil {
		// MySQL错误码1062 = Duplicate entry（用户名重复）
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return errors.New("username already taken")
		}
		return err
	}
	return nil
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	query := `SELECT id, username, password, created_at FROM users WHERE username = ?`
	row := r.db.QueryRow(query, username)

	user := &model.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByID(id string) (*model.User, error) {
	query := `SELECT id, username, password, created_at FROM users WHERE id = ?`
	row := r.db.QueryRow(query, id)

	user := &model.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}
