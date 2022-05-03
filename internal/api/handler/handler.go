package handler

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/url"
	"path"
)

type ContextKey string

type Links interface {
	Host() string
	Create(ctx context.Context, lnk string) (string, error)
	Get(ctx context.Context, key string) (string, error)
}

type Link struct {
	URL string `json:"url"`
}

type ShortenLink struct {
	Result string `json:"result"`
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
	r.With(r.ReadBody, r.GetShortLink).Post("/", r.SendPlainText)
	r.With(r.ReadBody, r.UnmarshalData, r.GetShortLink, r.MarshalData).Post("/api/shorten", r.SendJSON)
	return r
}

func (r *Router) ReadBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
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
		json, _ := json.Marshal(&lnk)
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
		key, err := r.ls.Create(req.Context(), str)
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
