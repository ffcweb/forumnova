package main

import (
	"database/sql"
	"flag"
	"forum/internal/models"
	"log/slog"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/alexedwards/scs/v2"
	_ "github.com/mattn/go-sqlite3"
)

// application contains the server's dependencies.
type application struct {
	logger         *slog.Logger
	threads        *models.ThreadModel
	users          *models.UserModel
	posts          *models.PostModel
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
}

func main() {
	addr := flag.String("addr", ":5000", "HTTP network address")
	dbPath := flag.String("dbPath", "./db.sqlite", "Path to database file")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(*dbPath)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	threadModel, userModel, postModel, err := models.NewModels(db)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour

	app := &application{
		logger:         logger,
		threads:        threadModel,
		users:          userModel,
		posts:          postModel,
		templateCache:  templateCache,
		sessionManager: sessionManager,
	}

	logger.Info("Starting server", "addr", *addr)

	err = http.ListenAndServe(*addr, app.routes())
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB opens a connection to the databate and returns an handle.
func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
