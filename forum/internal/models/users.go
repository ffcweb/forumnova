package models

import (
	"database/sql"
	"errors"
	"fmt"
)

// User holds data about a user.
type User struct {
	ID       int
	Username string
	Email    string
	Password string
}

// UserModel holds a database handle to manipulate a User.
type UserModel struct {
	DB *sql.DB
}

// NewUserModel creates a Users table and returns a new UserModel.
// create a new users database table and a database model to access it.
func NewUserModel(db *sql.DB) (*UserModel, error) {
	m := UserModel{db}
	err := m.createTable()
	if err != nil {
		return nil, fmt.Errorf("creating table: %w", err)
	}
	return &m, nil
}

// createTable creates a Users table.
func (m *UserModel) createTable() error {
	stmt := `
		CREATE TABLE IF NOT EXISTS Users (
		    id INTEGER PRIMARY KEY,
		    username TEXT NOT NULL,
		    email TEXT NOT NULL UNIQUE,
		    password TEXT NOT NULL
		);
	`
	_, err := m.DB.Exec(stmt)
	if err != nil {
		return fmt.Errorf("creating Users table: %w", err)
	}
	return nil
}

// Use the Insert method to add a new record to the "users" table.
func (m *UserModel) Insert(username, email, password string) (int, error) {
	stmt := `
		INSERT INTO Users (username, email, password)
		VALUES (?, ?, ?)
	`
	result, err := m.DB.Exec(stmt, username, email, password)
	if err != nil {
		return 0, fmt.Errorf("inserting new user in db: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last user id: %w", err)
	}
	return int(id), nil
}

// Get retrieves a user by their ID from the database and
// returns the user object or an error if not found.
func (m *UserModel) Get(id int) (*User, error) {
	var user User
	stmt := `SELECT id, username, email, password FROM Users WHERE id = ?`

	err := m.DB.QueryRow(stmt, id).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, fmt.Errorf("querying user by ID: %w", err)
	}
	return &user, nil
}

// Exists checks if a user with the given email address exists in the database
// and returns true if found or an error if not.
func (m *UserModel) Exists(emailAddress string) (bool, error) {
	stmt := `
		SELECT id
		FROM users
		WHERE email = ?
		LIMIT 1
	`
	var id int
	err := m.DB.QueryRow(stmt, emailAddress).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrNoRecord
		} else {
			return true, err
		}
	}
	return true, nil
}

// Authenticate verifies user credentials against the database and returns
// the user ID if valid or an error if invalid.
func (m *UserModel) Authenticate(email string, password string) (int, error) {
	var id int
	var realPassword string
	stmt := `
		SELECT id, password
		FROM users
		WHERE email = ?
	`
	err := m.DB.QueryRow(stmt, email).Scan(&id, &realPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	if realPassword == password {
		return id, nil
	}
	return 0, ErrInvalidCredentials
}
