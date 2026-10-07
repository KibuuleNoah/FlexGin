package server

import (
	"fmt"
	"log/slog"
	"net/http"

	_ "github.com/joho/godotenv/autoload"

	"flexgin/internal/database"
	"flexgin/internal/store"

	"flexgin/internal/config"
)

type Server struct {
	cfg config.Config
	log *slog.Logger
	db  *database.DB
}

func NewServer(cfg config.Config, log *slog.Logger, db *database.DB) *http.Server {
	q := store.New(db)

	s := &Server{
		cfg: cfg,
		log: log,
		db:  db,
	}

	return &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:           s.RegisterRoutes(q),
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTO,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
	}
}

// Shutdown drains the HTTP server, then closes the DB pool.
// func Shutdown(ctx context.Context, srv *http.Server, db *database.DB) error {
// 	err := srv.Shutdown(ctx)
// 	if cerr := db.Close(); cerr != nil {
// 		err = errors.Join(err, cerr)
// 	}
// 	return err
// }
