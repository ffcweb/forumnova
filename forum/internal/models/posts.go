package models

import (
	"database/sql"
	"fmt"
	"time"
)

// Post holds data about a single post in a Thread.
type Post struct {
	ID      int
	Body    string
	Author  *User
	Created time.Time
}

// PostModel holds a database handle for manipulating posts.
type PostModel struct {
	DB *sql.DB
}

// NewPostModel creates a Posts table and returns a new PostModel.
func NewPostModel(db *sql.DB) (*PostModel, error) {
	m := PostModel{db}
	err := m.createTable()
	if err != nil {
		return nil, fmt.Errorf("creating table: %w", err)
	}
	return &m, nil
}

// createTable creates a Posts table.
func (m *PostModel) createTable() error {
	stmt := `
		CREATE TABLE IF NOT EXISTS Posts (
		    id INTEGER PRIMARY KEY,
		    body TEXT NOT NULL,
		    author_id INTEGER NOT NULL REFERENCES Users,
		    thread_id INTEGER NOT NULL REFERENCES Threads,
		    created DATE NOT NULL
		);
	`
	_, err := m.DB.Exec(stmt)
	if err != nil {
		return fmt.Errorf("creating Posts table: %w", err)
	}
	return nil
}

// Insert inserts a new post in the Posts table.
func (m *PostModel) Insert(body string, threadId, authorId int) (int, error) {
	stmt := `
		INSERT INTO Posts (body, thread_id, author_id, created)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`
	result, err := m.DB.Exec(stmt, body, threadId, authorId)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}
