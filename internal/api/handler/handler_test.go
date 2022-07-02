package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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
		data   interface{}
		status int
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
				http.StatusCreated,
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
				0,
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
				0,
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
			r := NewRouter(nil, nil)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.args.data != nil {
				ctx := context.WithValue(req.Context(), ContextKey("DATA"), tt.args.data)
				ctx = context.WithValue(ctx, ContextKey("STATUS"), tt.args.status)
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
		data   interface{}
		status int
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
				http.StatusCreated,
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
				0,
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
				0,
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
			r := NewRouter(nil, nil)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.args.data != nil {
				ctx := context.WithValue(req.Context(), ContextKey("DATA"), tt.args.data)
				ctx = context.WithValue(ctx, ContextKey("STATUS"), tt.args.status)
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
			r := NewRouter(nil, nil)
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

func TestRouter_UnmarshalData(t *testing.T) {
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
			r := NewRouter(nil, nil)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.args.data != nil {
				ctx := context.WithValue(req.Context(), ContextKey("DATA"), tt.args.data)
				r.UnmarshalData(wantHandler).ServeHTTP(w, req.WithContext(ctx))
			} else {
				r.UnmarshalData(wantHandler).ServeHTTP(w, req)
			}
		})
	}
}

func TestRouter_MarshalData(t *testing.T) {
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
			r := NewRouter(nil, nil)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.args.data != nil {
				ctx := context.WithValue(req.Context(), ContextKey("DATA"), tt.args.data)
				r.MarshalData(wantHandler).ServeHTTP(w, req.WithContext(ctx))
			} else {
				r.MarshalData(wantHandler).ServeHTTP(w, req)
			}
		})
	}
}
