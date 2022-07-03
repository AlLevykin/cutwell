package pg

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"github.com/AlLevykin/cutwell/internal/api/handler"
	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type LinkStore struct {
	db *sql.DB
}

func NewLinkStore(dsn string) *LinkStore {
	goose.SetBaseFS(embedMigrations)
	db, err := goose.OpenDBWithDriver("postgres", dsn)
	if err != nil {
		// log
		fmt.Println(err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		db.Close()
		db = nil
	}
	return &LinkStore{
		db,
	}
}

func (ls *LinkStore) Ping(ctx context.Context) error {
	if ls.db == nil {
		return pq.ErrNotSupported
	}
	if err := ls.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func (ls *LinkStore) Host() string {
	return ""
}

func (ls *LinkStore) Create(ctx context.Context, lnk string, u string) (string, error) {
	return "", nil
}

func (ls *LinkStore) Get(ctx context.Context, key string) (string, error) {
	return "", sql.ErrNoRows
}

func (ls *LinkStore) GetURLList(ctx context.Context, u string) ([]handler.Item, error) {
	return nil, nil
}

func (ls *LinkStore) Close() error {
	if err := ls.Ping(context.Background()); err != nil {
		return err
	}
	if err := ls.db.Close(); err != nil {
		return err
	}
	return nil
}
