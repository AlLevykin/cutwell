package store

import (
	"context"
	"database/sql"
	"github.com/AlLevykin/cutwell/internal/utils"
	"sync"
)

type Config struct {
	BaseURL   string
	KeyLength int
}

type LinkStore struct {
	sync.Mutex
	Storage   map[string]string
	KeyLength int
	BaseURL   string
}

func NewLinkStore(c Config) *LinkStore {
	return &LinkStore{
		Storage:   make(map[string]string),
		KeyLength: c.KeyLength,
		BaseURL:   c.BaseURL,
	}
}

func (ls *LinkStore) Host() string {
	return ls.BaseURL
}

func (ls *LinkStore) Create(ctx context.Context, lnk string) (string, error) {
	ls.Lock()
	defer ls.Unlock()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	key := utils.RandString(ls.KeyLength)
	ls.Storage[key] = lnk
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
	lnk, ok := ls.Storage[key]
	if ok {
		return lnk, nil
	}
	return "", sql.ErrNoRows
}
