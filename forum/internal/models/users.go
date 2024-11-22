package models

import (
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// User holds data about a user.
type User struct {
	ID             int
	Username       string
	Email          string
	HashedPassword []byte
}

// UserModel holds a database handle to manipulate a User.
type UserModel struct {
	DB *sql.DB
}

// NewUserModel creates a Users table and returns a new UserModel.
func NewUserModel(db *sql.DB) (*UserModel, error) {
	m := UserModel{db}
	err := m.createTable()
	if err != nil {
		return nil, fmt.Errorf("creating table: %w", err)
	}
	return &m, nil
}

// createTable creates a Users table if it doesn't already exist.
func (m *UserModel) createTable() error {
	stmt := `
		CREATE TABLE IF NOT EXISTS Users (
		    id INTEGER PRIMARY KEY,
		    username TEXT NOT NULL,
		    email TEXT NOT NULL UNIQUE,
		    hashed_password TEXT NOT NULL
		);
	`
	_, err := m.DB.Exec(stmt)
	if err != nil {
		return fmt.Errorf("creating Users table: %w", err)
	}
	return nil
}

// Insert adds a new record to the "Users" table.
func (m *UserModel) Insert(username, email, password string) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hashing password: %w", err)
	}

	stmt := `
		INSERT INTO Users (username, email, hashed_password)
		VALUES (?, ?, ?)
	`
	result, err := m.DB.Exec(stmt, username, email, string(hashedPassword))
	if err != nil {
		return 0, fmt.Errorf("inserting new user in db: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last user id: %w", err)
	}
	return int(id), nil
}

// Get retrieves a user by their ID.
func (m *UserModel) Get(id int) (*User, error) {
	var user User
	stmt := `SELECT id, username, email, hashed_password FROM Users WHERE id = ?`

	err := m.DB.QueryRow(stmt, id).Scan(&user.ID, &user.Username, &user.Email, &user.HashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, fmt.Errorf("querying user by ID: %w", err)
	}
	return &user, nil
}

// Exists checks if a user with the given email exists.
func (m *UserModel) Exists(email string) (bool, error) {
	stmt := `SELECT id FROM Users WHERE email = ? LIMIT 1`
	var id int
	err := m.DB.QueryRow(stmt, email).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("checking email existence: %w", err)
	}
	return true, nil
}

// Authenticate verifies a user's credentials.
func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte
	stmt := `SELECT id, hashed_password FROM Users WHERE email = ?`

	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}
		return 0, fmt.Errorf("querying user by email: %w", err)
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		}
		return 0, fmt.Errorf("verifying password: %w", err)
	}

	return id, nil
}
