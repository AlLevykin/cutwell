package store

import (
	"context"
	"reflect"
	"testing"
)

func TestLinkStore_Create(t *testing.T) {
	type fields struct {
		storage map[string]string
		keyLen  int
	}
	type args struct {
		withContext bool
		lnk         string
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
			fields{map[string]string{}, 9},
			args{false, "ya.ru"},
			9,
			false,
		},
		{
			"context done",
			fields{map[string]string{}, 9},
			args{true, "ya.ru"},
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
				Storage:   tt.fields.storage,
				KeyLength: tt.fields.keyLen,
			}
			got, err := ls.Create(ctx, tt.args.lnk)
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
				Storage: tt.fields.storage,
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
				Storage:   make(map[string]string),
				KeyLength: 9,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLinkStore(tt.keyLength); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLinkStore() = %v, want %v", got, tt.want)
			}
		})
	}
}
