package handler

import (
	"context"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/url"
	"path"
)

type Links interface {
	Create(ctx context.Context, lnk string) (string, error)
	Get(ctx context.Context, key string) (string, error)
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
	r.Post("/", r.CreateShortLink)
	return r
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

func (r *Router) CreateShortLink(w http.ResponseWriter, req *http.Request) {
	buf, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	link := string(buf)
	key, err := r.ls.Create(req.Context(), link)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	u := &url.URL{
		Scheme: "http",
		Host:   req.Host,
		Path:   key,
	}
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(u.String()))
}
