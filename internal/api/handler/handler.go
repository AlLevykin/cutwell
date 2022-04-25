package handler

import (
	"context"
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
	*http.ServeMux
	ls Links
}

func NewRouter(ls Links) *Router {
	r := &Router{
		ServeMux: http.NewServeMux(),
		ls:       ls,
	}
	r.HandleFunc("/", r.SwitchHandlers)
	return r
}

func (r *Router) SwitchHandlers(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.Redirect(w, req)
	case http.MethodPost:
		r.CreateShortLink(w, req)
	default:
		http.Error(w, "bad request", http.StatusBadRequest)
	}
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
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(u.String()))
}
