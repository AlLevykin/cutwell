package handler

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type ContextKey string

type Links interface {
	Host() string
	Create(ctx context.Context, lnk string, u string) (string, error)
	Get(ctx context.Context, key string) (string, error)
	GetUrlList(ctx context.Context, u string) ([]Item, error)
}

type Link struct {
	URL string `json:"url"`
}

type ShortenLink struct {
	Result string `json:"result"`
}

type Item struct {
	ShortURL string `json:"short_url"`
	URL      string `json:"original_url"`
}

type Router struct {
	*chi.Mux
	ls Links
}

func NewRouter(ls Links) *Router {
	r := &Router{
		Mux: chi.NewRouter(),
		ls:  ls,
	}
	r.Get("/{key}", r.Redirect)
	r.With(r.Auth, r.CheckSession, r.ReadBody, r.GetShortLink, r.Compress).Post("/", r.SendPlainText)
	r.With(r.Auth, r.CheckSession, r.ReadBody, r.UnmarshalData, r.GetShortLink, r.MarshalData, r.Compress).Post("/api/shorten", r.SendJSON)
	r.With(r.Auth, r.CheckSession, r.GetUrls, r.Compress).Get("/api/user/urls", r.SendJSON)
	return r
}

func (r *Router) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		uid := "00000001"
		ctx := context.WithValue(req.Context(), ContextKey("USERID"), uid)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (r *Router) CheckSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		uid, ok := req.Context().Value(ContextKey("USERID")).(string)
		if !ok || len(uid) == 0 {
			http.Error(w, "can't get user id", http.StatusBadRequest)
			return
		}
		if cookie, err := req.Cookie("cutwell-session"); err != nil {
			cookie = &http.Cookie{
				Name:  "cutwell-session",
				Value: uid,
			}
			http.SetCookie(w, cookie)
		}
		next.ServeHTTP(w, req)
	})
}

func (r *Router) ReadBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var reader io.Reader
		if req.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(req.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			reader = gz
			defer gz.Close()
		} else {
			reader = req.Body
		}
		buf, err := io.ReadAll(reader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), ContextKey("DATA"), string(buf))
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (r *Router) MarshalData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		data := req.Context().Value(ContextKey("DATA"))
		if data == nil {
			http.Error(w, "can't get context data", http.StatusBadRequest)
			return
		}
		res, ok := data.(string)
		if !ok {
			http.Error(w, "can't get context data", http.StatusBadRequest)
			return
		}
		lnk := ShortenLink{
			Result: res,
		}
		json, err := json.Marshal(&lnk)
		if err != nil {
			http.Error(w, "can't get context data", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), ContextKey("DATA"), string(json))
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (r *Router) UnmarshalData(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		data := req.Context().Value(ContextKey("DATA"))
		if data == nil {
			http.Error(w, "can't get context data", http.StatusBadRequest)
			return
		}
		str, ok := data.(string)
		if !ok {
			http.Error(w, "can't get context data", http.StatusBadRequest)
			return
		}
		lnk := Link{}
		err := json.Unmarshal([]byte(str), &lnk)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(req.Context(), ContextKey("DATA"), lnk.URL)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (r *Router) GetShortLink(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		uid, ok := req.Context().Value(ContextKey("USERID")).(string)
		if !ok || len(uid) == 0 {
			http.Error(w, "can't get user id", http.StatusBadRequest)
			return
		}
		data := req.Context().Value(ContextKey("DATA"))
		if data == nil {
			http.Error(w, "can't get context data", http.StatusBadRequest)
			return
		}
		str, ok := data.(string)
		if !ok {
			http.Error(w, "can't get context data", http.StatusBadRequest)
			return
		}
		key, err := r.ls.Create(req.Context(), str, uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		u := &url.URL{
			Scheme: "http",
			Host:   r.ls.Host(),
			Path:   key,
		}
		ctx := context.WithValue(req.Context(), ContextKey("DATA"), u.String())
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (r *Router) Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, req)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, req)
	})
}

func (r *Router) GetUrls(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		uid, ok := req.Context().Value(ContextKey("USERID")).(string)
		if !ok || len(uid) == 0 {
			http.Error(w, "can't get user id", http.StatusBadRequest)
			return
		}
		lnks, err := r.ls.GetUrlList(req.Context(), uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}
		json, err := json.Marshal(&lnks)
		if err != nil {
			http.Error(w, "can't get context data", http.StatusNoContent)
			return
		}
		ctx := context.WithValue(req.Context(), ContextKey("DATA"), string(json))
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (r *Router) SendPlainText(w http.ResponseWriter, req *http.Request) {
	data := req.Context().Value(ContextKey("DATA"))
	if data == nil {
		http.Error(w, "can't get context data", http.StatusBadRequest)
		return
	}
	str, ok := data.(string)
	if !ok {
		http.Error(w, "can't get context data", http.StatusBadRequest)
		return
	}
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(str))
}

func (r *Router) SendJSON(w http.ResponseWriter, req *http.Request) {
	data := req.Context().Value(ContextKey("DATA"))
	if data == nil {
		http.Error(w, "can't get context data", http.StatusBadRequest)
		return
	}
	str, ok := data.(string)
	if !ok {
		http.Error(w, "can't get context data", http.StatusBadRequest)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(str))
}

func (r *Router) Redirect(w http.ResponseWriter, req *http.Request) {
	key := path.Base(req.URL.Path)
	lnk, err := r.ls.Get(req.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", lnk)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
