package store

import (
	"context"
	"database/sql"
	"github.com/AlLevykin/cutwell/internal/api/handler"
	"github.com/AlLevykin/cutwell/internal/utils"
	"net/url"
	"sync"
)

type Config struct {
	BaseURL   string
	KeyLength int
}

type LinkStore struct {
	sync.Mutex
	File      string
	Mem       map[string]string
	Users     map[string]string
	KeyLength int
	BaseURL   string
}

func NewLinkStore(c Config, fileName string) *LinkStore {
	return &LinkStore{
		File:      fileName,
		Mem:       FileToMap(fileName),
		Users:     FileToMap(fileName + ".users"),
		KeyLength: c.KeyLength,
		BaseURL:   c.BaseURL,
	}
}

func (ls *LinkStore) Host() string {
	u, err := url.Parse(ls.BaseURL)
	if err != nil {
		return ls.BaseURL
	}
	return u.Host
}

func (ls *LinkStore) Create(ctx context.Context, lnk string, u string) (string, error) {
	ls.Lock()
	defer ls.Unlock()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	key := utils.RandString(ls.KeyLength)
	ls.Mem[key] = lnk
	ls.Users[key] = u
	return key, nil
}

func (ls *LinkStore) Get(ctx context.Context, key string) (string, error) {
	ls.Lock()
	defer ls.Unlock()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	lnk, ok := ls.Mem[key]
	if ok {
		return lnk, nil
	}
	return "", sql.ErrNoRows
}

func (ls *LinkStore) GetUrlList(ctx context.Context, u string) ([]handler.Item, error) {
	ls.Lock()
	defer ls.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	result := make([]handler.Item, 0)

	for lnk, user := range ls.Users {
		if user == u {
			shortUrl := &url.URL{
				Scheme: "http",
				Host:   ls.Host(),
				Path:   lnk,
			}
			result = append(result,
				handler.Item{
					ShortURL: shortUrl.String(),
					URL:      ls.Mem[lnk],
				},
			)
		}
	}

	if len(result) == 0 {
		return nil, sql.ErrNoRows
	}

	return result, nil
}

func (ls *LinkStore) Save() error {
	if err := MapToFile(ls.Mem, ls.File); err != nil {
		return err
	}
	if err := MapToFile(ls.Users, ls.File+".users"); err != nil {
		return err
	}
	return nil
}
