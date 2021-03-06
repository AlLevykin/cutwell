package store

import (
	"context"
	"reflect"
	"testing"
)

func TestLinkStore_Create(t *testing.T) {
	type fields struct {
		storage map[string]string
		users   map[string]string
		keyLen  int
	}
	type args struct {
		withContext bool
		lnk         string
		u           string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			"ok",
			fields{
				map[string]string{},
				map[string]string{},
				9,
			},
			args{false, "ya.ru", "000001"},
			9,
			false,
		},
		{
			"context done",
			fields{
				map[string]string{},
				map[string]string{},
				9,
			},
			args{true, "ya.ru", "000001"},
			0,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context
			if tt.args.withContext {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			} else {
				ctx = context.Background()
			}
			ls := &LinkStore{
				Mem:       tt.fields.storage,
				Users:     tt.fields.users,
				KeyLength: tt.fields.keyLen,
			}
			got, err := ls.Create(ctx, tt.args.lnk, tt.args.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("Create() got = %v, want len = %v", got, tt.fields.keyLen)
			}
		})
	}
}

func TestLinkStore_Get(t *testing.T) {
	type fields struct {
		storage map[string]string
	}
	type args struct {
		withContext bool
		key         string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			"ok",
			fields{map[string]string{"1": "one"}},
			args{false, "1"},
			"one",
			false,
		},
		{
			"context done",
			fields{map[string]string{"1": "one"}},
			args{true, "1"},
			"",
			true,
		},
		{
			"no rows error",
			fields{map[string]string{"1": "one"}},
			args{false, "2"},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context
			if tt.args.withContext {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			} else {
				ctx = context.Background()
			}
			ls := &LinkStore{
				Mem: tt.fields.storage,
			}
			got, err := ls.Get(ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLinkStore(t *testing.T) {
	tests := []struct {
		name      string
		keyLength int
		want      *LinkStore
	}{
		{
			"ok",
			9,
			&LinkStore{
				Mem:       make(map[string]string),
				Users:     make(map[string]string),
				KeyLength: 9,
				BaseURL:   "127.0.0.1:8080",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				KeyLength: tt.keyLength,
				BaseURL:   "127.0.0.1:8080",
			}
			if got := NewLinkStore(cfg, ""); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLinkStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLinkStore_Host(t *testing.T) {
	type fields struct {
		BaseURL string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"Host",
			fields{"127.0.0.1:8080"},
			"127.0.0.1:8080",
		},
		{
			"Host",
			fields{"http://127.0.0.1:8080"},
			"127.0.0.1:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := &LinkStore{
				BaseURL: tt.fields.BaseURL,
			}
			if got := ls.Host(); got != tt.want {
				t.Errorf("Host() = %v, want %v", got, tt.want)
			}
		})
	}
}
