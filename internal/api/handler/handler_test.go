package handler

import (
	"fmt"
	"github.com/AlLevykin/cutwell/internal/app/store"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouter_CreateShortLink(t *testing.T) {
	type args struct {
		lnk    string
		keyLen int
	}
	type want struct {
		code        int
		contentType string
		contentLen  int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"ok",
			args{
				"ya.ru",
				9,
			},
			want{
				code:        http.StatusCreated,
				contentType: "text/plain; charset=utf-8",
				contentLen:  9,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.args.lnk)
			req := httptest.NewRequest(http.MethodPost, "/", body)
			w := httptest.NewRecorder()
			ls := store.NewLinkStore(tt.args.keyLen)
			r := NewRouter(ls)
			r.CreateShortLink(w, req)
			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			key := string(resBody)
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}
			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}
			if len(key) != tt.want.contentLen {
				t.Errorf("Expected status code %d, got %s len = %d", tt.want.contentLen, key, len(key))
			}
		})
	}
}

func TestRouter_Redirect(t *testing.T) {
	type args struct {
		key string
	}
	type want struct {
		code int
		key  string
		lnk  string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"ok",
			args{
				key: "xvtWzBTea",
			},
			want{
				code: http.StatusTemporaryRedirect,
				key:  "xvtWzBTea",
				lnk:  "http://ctqplvcsifak.biz/jqepl7eormvew4",
			},
		},
		{
			"bad request",
			args{
				key: "xvtWzBTea",
			},
			want{
				code: http.StatusBadRequest,
				key:  "111111111",
				lnk:  "http://ctqplvcsifak.biz/jqepl7eormvew4",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", tt.args.key), nil)
			w := httptest.NewRecorder()
			ls := &store.LinkStore{
				Storage: map[string]string{
					tt.want.key: tt.want.lnk,
				},
				KeyLength: len(tt.want.key),
			}
			r := NewRouter(ls)
			r.Redirect(w, req)
			res := w.Result()
			defer res.Body.Close()
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}
		})
	}
}
