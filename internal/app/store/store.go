package store

import (
	"context"
	"database/sql"
	"github.com/AlLevykin/cutwell/internal/utils"
	"sync"
)

type LinkStore struct {
	sync.Mutex
	storage   map[string]string
	keyLength int
}

func NewLinkStore(kl int) *LinkStore {
	return &LinkStore{
		storage:   make(map[string]string),
		keyLength: kl,
	}
}

func (ls *LinkStore) Create(ctx context.Context, lnk string) (string, error) {
	ls.Lock()
	defer ls.Unlock()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	key := utils.RandString(ls.keyLength)
	ls.storage[key] = lnk
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
	lnk, ok := ls.storage[key]
	if ok {
		return lnk, nil
	}
	return "", sql.ErrNoRows
}
