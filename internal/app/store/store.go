package store

import (
	"context"
	"database/sql"
	"github.com/AlLevykin/cutwell/internal/utils"
	"io"
	"net/url"
	"sync"
)

type Config struct {
	BaseURL   string
	KeyLength int
}

type LinkStore struct {
	sync.Mutex
	Storage   io.ReadWriteCloser
	Mem       map[string]string
	KeyLength int
	BaseURL   string
}

func NewLinkStore(c Config, s io.ReadWriteCloser) *LinkStore {
	return &LinkStore{
		Storage:   s,
		Mem:       make(map[string]string),
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

func (ls *LinkStore) Create(ctx context.Context, lnk string) (string, error) {
	ls.Lock()
	defer ls.Unlock()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	key := utils.RandString(ls.KeyLength)
	ls.Mem[key] = lnk
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
