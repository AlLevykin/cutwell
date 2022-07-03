package pg

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"github.com/AlLevykin/cutwell/internal/api/handler"
	"github.com/AlLevykin/cutwell/internal/app/store"
	"github.com/AlLevykin/cutwell/internal/utils"
	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"net/url"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type LinkStore struct {
	db        *sql.DB
	KeyLength int
	BaseURL   string
}

func NewLinkStore(c store.Config, dsn string) *LinkStore {
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
		db:        db,
		KeyLength: c.KeyLength,
		BaseURL:   c.BaseURL,
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
	u, err := url.Parse(ls.BaseURL)
	if err != nil {
		return ls.BaseURL
	}
	return u.Host
}

func (ls *LinkStore) Create(ctx context.Context, lnk string, u string) (string, error) {
	key := utils.RandString(ls.KeyLength)
	_, err := ls.db.ExecContext(ctx,
		"INSERT INTO urls(id, lnk, usr) VALUES($1,$2,$3)",
		key, lnk, u)
	if err != nil {
		return "", err
	}
	return key, nil
}

func (ls *LinkStore) Get(ctx context.Context, key string) (string, error) {
	rows, err := ls.db.QueryContext(ctx, "SELECT lnk from urls where id=$1", key)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if rows.Next() {
		var link string
		err = rows.Scan(&link)
		if err != nil {
			return "", err
		}
		return link, nil
	}

	if err = rows.Err(); err != nil {
		return "", err
	}

	return "", sql.ErrNoRows
}

func (ls *LinkStore) GetURLList(ctx context.Context, u string) ([]handler.Item, error) {
	result := make([]handler.Item, 0)

	rows, err := ls.db.QueryContext(ctx, "SELECT id, lnk from urls where usr=$1", u)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var link string
		err = rows.Scan(&key, &link)
		if err != nil {
			return nil, err
		}

		shortURL := &url.URL{
			Scheme: "http",
			Host:   ls.Host(),
			Path:   key,
		}

		result = append(result,
			handler.Item{
				ShortURL: shortURL.String(),
				URL:      link,
			},
		)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, sql.ErrNoRows
	}

	return result, nil
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
