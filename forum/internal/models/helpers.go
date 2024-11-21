package models

import (
	"database/sql"
)

// NewModels creates all models necessary for the application.
func NewModels(db *sql.DB) (*ThreadModel, *UserModel, *PostModel, error) {
	threadModel, err := NewThreadModel(db)
	if err != nil {
		return nil, nil, nil, err
	}
	userModel, err := NewUserModel(db)
	if err != nil {
		return nil, nil, nil, err
	}
	postModel, err := NewPostModel(db)
	if err != nil {
		return nil, nil, nil, err
	}
	return threadModel, userModel, postModel, nil
}
