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
	"log"
	"net/url"
	"sync"
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

func (ls *LinkStore) Create(ctx context.Context, lnk string, user string) (string, error) {
	key := utils.RandString(ls.KeyLength)
	_, err := ls.db.ExecContext(ctx,
		"INSERT INTO urls(id, lnk, usr) VALUES($1,$2,$3)",
		key, lnk, user)

	if err != nil {
		return "", err
	}
	return key, nil
}

func (ls *LinkStore) Find(ctx context.Context, lnk string) (string, error) {
	rows, err := ls.db.QueryContext(ctx, "SELECT id from urls where lnk=$1", lnk)
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

func (ls *LinkStore) Get(ctx context.Context, key string) (string, error) {
	rows, err := ls.db.QueryContext(ctx, "SELECT lnk FROM urls WHERE id=$1 AND removed = false", key)
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

func (ls *LinkStore) Batch(ctx context.Context, batch []handler.BatchItem, user string) ([]handler.ResultItem, error) {
	tx, err := ls.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Fatalf("batch: unable to rollback: %v", err)
		}
	}()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO urls(id, lnk, usr) VALUES($1,$2,$3)")
	if err != nil {
		return nil, err
	}

	res := make([]handler.ResultItem, 0, len(batch))

	for _, i := range batch {
		key := utils.RandString(ls.KeyLength)
		if _, err = stmt.ExecContext(ctx, key, i.URL, user); err != nil {
			return nil, err
		}
		shortURL := &url.URL{
			Scheme: "http",
			Host:   ls.Host(),
			Path:   key,
		}
		res = append(res, handler.ResultItem{ID: i.ID, URL: shortURL.String()})
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ls *LinkStore) Delete(ctx context.Context, urls []string, user string) error {

	tx, err := ls.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Fatalf("batch: unable to rollback: %v", err)
		}
	}()

	stmt, err := tx.PrepareContext(ctx, "UPDATE urls SET removed = true WHERE id = '$1' AND usr = '$2'")
	if err != nil {
		return err
	}

	worker := func(url string) chan error {
		ch := make(chan error)
		go func() {
			_, err := stmt.ExecContext(ctx, url, user)
			if err != nil {
				log.Printf("delete: unable to update table: %v", err)
			}
			ch <- err
			close(ch)
		}()
		return ch
	}

	fanIn := func(chans []chan error) chan error {
		res := make(chan error)
		var wg sync.WaitGroup
		wg.Add(len(chans))

		for _, ch := range chans {
			go func(ch chan error) {
				defer wg.Done()
				for err := range ch {
					res <- err
				}
			}(ch)
		}

		go func() {
			wg.Wait()
			close(res)
		}()
		return res
	}

	var chans []chan error
	for _, url := range urls {
		ch := worker(url)
		chans = append(chans, ch)
	}

	ch := fanIn(chans)
	for err := range ch {
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
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
