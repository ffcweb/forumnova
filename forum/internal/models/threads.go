package models

import (
	"database/sql"
	"fmt"
	"time"
)

// Thread holds data about a thread.
type Thread struct {
	ID          int
	Title       string
	Author      *User
	Created     time.Time
	Posts       []*Post
	FieldErrors map[string]string
}

// ThreadModel holds a database handle to manipulate a Thread.
type ThreadModel struct {
	DB *sql.DB
}

// NewThreadModel creates a Threads table and returns a ThreadModel.
func NewThreadModel(db *sql.DB) (*ThreadModel, error) {
	m := ThreadModel{db}
	err := m.createTable()
	if err != nil {
		return nil, fmt.Errorf("creating table: %w", err)
	}
	return &m, nil
}

// createTable creates a Threads table.
func (m *ThreadModel) createTable() error {
	stmt := `
		CREATE TABLE IF NOT EXISTS Threads (
		    id INTEGER PRIMARY KEY,
		    title TEXT NOT NULL,
		    author_id INTEGER NOT NULL REFERENCES Users,
		    created DATE NOT NULL
		)
	`
	_, err := m.DB.Exec(stmt)
	if err != nil {
		return fmt.Errorf("creating Threads table: %w", err)
	}
	return nil
}

// Insert inserts a new thread in the database.
func (m *ThreadModel) Insert(title string, authorId int) (int, error) {
	stmt := `
		INSERT INTO Threads (title, author_id, created)
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`
	result, err := m.DB.Exec(stmt, title, authorId)
	if err != nil {
		return 0, fmt.Errorf("inserting new thread in db: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last thread id: %w", err)
	}
	return int(id), nil
}

// Get retrieves the thread with the given id from the database.
func (m *ThreadModel) Get(id int) (*Thread, error) {
	stmt := `
		SELECT T.id, T.title, T.created, U.id, U.username, U.email
		FROM Threads T, Users U
		WHERE T.author_id = U.id AND T.id = ?
	`
	row := m.DB.QueryRow(stmt, id)
	t, err := m.newThread(row, "ASC")
	if err != nil {
		return nil, fmt.Errorf("creating new thread: %w", err)
	}
	return t, nil
}

// Latests retrieves the 10 latests threads from the database.
func (m *ThreadModel) Latests() ([]*Thread, error) {
	stmt := `
		SELECT T.id, T.title, T.created, U.id, U.username, U.email
		FROM Threads T, Users U
		WHERE T.author_id = U.id
		ORDER BY T.created DESC
		LIMIT 9
	`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, fmt.Errorf("getting latests threads: %w", err)
	}
	defer rows.Close()

	var threads []*Thread
	for rows.Next() {
		t, err := m.newThread(rows, "DESC")
		if err != nil {
			return nil, fmt.Errorf("creating thread: %w", err)
		}
		threads = append(threads, t)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating over rows for latests threads: %w", err)
	}

	return threads, nil
}

// scanner implements the Scan function.
type scanner interface {
	Scan(dest ...any) error
}

// newThread creates a new Thread. It also creates a User to represent
// the Thread's author, and the Posts associated with that Thread.
func (m *ThreadModel) newThread(s scanner, postOrder string) (*Thread, error) {
	var (
		t Thread
		u User
	)
	err := s.Scan(
		&t.ID, &t.Title, &t.Created,
		&u.ID, &u.Username, &u.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning row: %w", err)
	}
	t.Author = &u
	t.Posts, err = m.getPosts(t.ID, postOrder)
	if err != nil {
		return nil, fmt.Errorf("getting posts with thread id %v: %w", t.ID, err)
	}
	return &t, nil
}

// getPosts retrieves all Posts related to the Thread with the given threadID.
// The value of order must be "ASC" or "DESC".
func (m *ThreadModel) getPosts(threadID int, order string) ([]*Post, error) {
	stmt := fmt.Sprintf(
		`
			SELECT P.id, P.body, P.created, U.id, U.username, U.email
			FROM Posts P, Users U
			WHERE P.author_id = U.id AND P.thread_id = ?
			ORDER BY P.created %v
		`,
		order,
	)
	rows, err := m.DB.Query(stmt, threadID, order)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		var (
			p Post
			u User
		)
		err := rows.Scan(
			&p.ID, &p.Body, &p.Created,
			&u.ID, &u.Username, &u.Email,
		)
		if err != nil {
			return nil, err
		}
		p.Author = &u
		posts = append(posts, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
