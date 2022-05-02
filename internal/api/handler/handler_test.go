package handler

import (
	"context"
	"fmt"
	"github.com/AlLevykin/cutwell/internal/app/store"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"
)

type mockReader struct {
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestRouter_SendPlainText(t *testing.T) {
	type args struct {
		data interface{}
	}
	type want struct {
		code        int
		contentType string
		data        string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"ok",
			args{
				"data",
			},
			want{
				code:        http.StatusCreated,
				contentType: "text/plain; charset=utf-8",
				data:        "data",
			},
		},
		{
			"nil",
			args{
				nil,
			},
			want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				data:        "can't get context data",
			},
		},
		{
			"wrong data type",
			args{
				100,
			},
			want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				data:        "can't get context data",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter(nil)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.args.data != nil {
				ctx := context.WithValue(req.Context(), ContextKey("DATA"), tt.args.data)
				r.SendPlainText(w, req.WithContext(ctx))
			} else {
				r.SendPlainText(w, req)
			}
			res := w.Result()
			defer res.Body.Close()
			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			txt := string(b)
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}
			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}
			if strings.TrimRight(txt, "\n") != tt.want.data {
				t.Errorf("Expected data %s, got %s", tt.want.data, txt)
			}
		})
	}
}

func TestRouter_SendJson(t *testing.T) {
	type args struct {
		data interface{}
	}
	type want struct {
		code        int
		contentType string
		data        string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"ok",
			args{
				"{\"result\":\"http://localhost:8080/BvMIYOqSF\"}",
			},
			want{
				code:        http.StatusCreated,
				contentType: "application/json",
				data:        "{\"result\":\"http://localhost:8080/BvMIYOqSF\"}",
			},
		},
		{
			"nil",
			args{
				nil,
			},
			want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				data:        "can't get context data",
			},
		},
		{
			"wrong data type",
			args{
				100,
			},
			want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				data:        "can't get context data",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter(nil)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.args.data != nil {
				ctx := context.WithValue(req.Context(), ContextKey("DATA"), tt.args.data)
				r.SendJSON(w, req.WithContext(ctx))
			} else {
				r.SendJSON(w, req)
			}
			res := w.Result()
			defer res.Body.Close()
			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			txt := string(b)
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}
			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}
			if strings.TrimRight(txt, "\n") != tt.want.data {
				t.Errorf("Expected data %s, got %s", tt.want.data, txt)
			}
		})
	}
}

func TestRouter_ReadBody(t *testing.T) {
	type args struct {
		data string
	}
	type want struct {
		data string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"ok",
			args{
				"data",
			},
			want{
				"data",
			},
		},
		{
			"fail",
			args{
				"",
			},
			want{
				"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if len(tt.args.data) != 0 {
				body = strings.NewReader(tt.args.data)
			} else {
				body = &mockReader{}
			}
			r := NewRouter(nil)
			w := httptest.NewRecorder()
			br := httptest.NewRequest(http.MethodPost, "/", body)
			wantHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				data := req.Context().Value(ContextKey("DATA"))
				if data == nil {
					t.Error("DATA not present")
				}
				str, ok := data.(string)
				if !ok {
					t.Error("not string")
				}
				if str != tt.want.data {
					t.Error("wrong DATA")
				}
			})
			r.ReadBody(wantHandler).ServeHTTP(w, br)
		})
	}
}

func TestRouter_UnmarshalJson(t *testing.T) {
	type args struct {
		data interface{}
	}
	type want struct {
		data string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"ok",
			args{
				"{\"url\":\"ya.ru\"}",
			},
			want{
				"ya.ru",
			},
		},
		{
			"wrong data type",
			args{
				100,
			},
			want{
				"",
			},
		},
		{
			"wrong json",
			args{
				"{\"url\",\"ya.ru\"}",
			},
			want{
				"",
			},
		},
		{
			"nil",
			args{
				nil,
			},
			want{
				"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				data := req.Context().Value(ContextKey("DATA"))
				if data == nil {
					t.Error("DATA not present")
				}
				str, ok := data.(string)
				if !ok {
					t.Error("not string")
				}
				if str != tt.want.data {
					t.Error("wrong DATA")
				}
			})
			r := NewRouter(nil)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.args.data != nil {
				ctx := context.WithValue(req.Context(), ContextKey("DATA"), tt.args.data)
				r.UnmarshalJSON(wantHandler).ServeHTTP(w, req.WithContext(ctx))
			} else {
				r.UnmarshalJSON(wantHandler).ServeHTTP(w, req)
			}
		})
	}
}

func TestRouter_MarshalJson(t *testing.T) {
	type args struct {
		data interface{}
	}
	type want struct {
		data string
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
			},
			want{
				"{\"result\":\"ya.ru\"}",
			},
		},
		{
			"wrong data type",
			args{
				100,
			},
			want{
				"",
			},
		},
		{
			"nil",
			args{
				nil,
			},
			want{
				"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				data := req.Context().Value(ContextKey("DATA"))
				if data == nil {
					t.Error("DATA not present")
				}
				str, ok := data.(string)
				if !ok {
					t.Error("not string")
				}
				if str != tt.want.data {
					t.Error("wrong DATA")
				}
			})
			r := NewRouter(nil)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.args.data != nil {
				ctx := context.WithValue(req.Context(), ContextKey("DATA"), tt.args.data)
				r.MarshalJSON(wantHandler).ServeHTTP(w, req.WithContext(ctx))
			} else {
				r.MarshalJSON(wantHandler).ServeHTTP(w, req)
			}
		})
	}
}

func TestRouter_GetShortLink(t *testing.T) {
	type args struct {
		data   interface{}
		keyLen int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"ok",
			args{
				"ya.ru",
				9,
			},
		},
		{
			"wrong data type",
			args{
				100,
				9,
			},
		},
		{
			"nil",
			args{
				nil,
				9,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				data := req.Context().Value(ContextKey("DATA"))
				if data == nil {
					t.Error("DATA not present")
				}
				str, ok := data.(string)
				if !ok {
					t.Error("not string")
				}
				u, err := url.Parse(str)
				if err != nil {
					t.Error("wrong url")
				}
				key := path.Base(u.Path)
				if len(key) != tt.args.keyLen {
					t.Error("wrong short link")
				}
			})
			ls := store.NewLinkStore(tt.args.keyLen)
			r := NewRouter(ls)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.args.data != nil {
				ctx := context.WithValue(req.Context(), ContextKey("DATA"), tt.args.data)
				r.GetShortLink(wantHandler).ServeHTTP(w, req.WithContext(ctx))
			} else {
				r.GetShortLink(wantHandler).ServeHTTP(w, req)
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
