package handler

import (
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/AlLevykin/cutwell/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type ContextKey string

type Links interface {
	Host() string
	Create(ctx context.Context, lnk string, user string) (string, error)
	Get(ctx context.Context, key string) (string, error)
	GetURLList(ctx context.Context, user string) ([]Item, error)
	Ping(ctx context.Context) error
	Batch(ctx context.Context, batch []BatchItem, user string) ([]ResultItem, error)
	Find(ctx context.Context, lnk string) (string, error)
	Delete(ctx context.Context, urls []string, user string) error
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

type BatchItem struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

type ResultItem struct {
	ID  string `json:"correlation_id"`
	URL string `json:"short_url"`
}

type Router struct {
	*chi.Mux
	ls      Links
	decoder *utils.Decoder
}

func NewRouter(ls Links, d *utils.Decoder) *Router {
	r := &Router{
		Mux:     chi.NewRouter(),
		ls:      ls,
		decoder: d,
	}
	r.Get("/{key}", r.Redirect)
	r.With(r.CheckSession, r.ReadBody, r.GetShortLink, r.Compress).Post("/", r.SendPlainText)
	r.With(r.CheckSession, r.ReadBody, r.UnmarshalData, r.GetShortLink, r.MarshalData, r.Compress).Post("/api/shorten", r.SendJSON)
	r.With(r.CheckSession, r.GetUrls, r.Compress).Get("/api/user/urls", r.SendJSON)
	r.Get("/ping", r.Ping)
	r.With(r.CheckSession, r.ReadBody, r.Batch, r.Compress).Post("/api/shorten/batch", r.SendJSON)
	r.With(r.CheckSession, r.ReadBody).Delete("/api/user/urls", r.DeleteUrls)
	return r
}

func (r *Router) CheckSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		uid := utils.RandString(6)
		if cookie, err := req.Cookie("cutwell-session"); err != nil {
			cookie = &http.Cookie{
				Name:  "cutwell-session",
				Value: uid,
				Path:  "/",
			}
			http.SetCookie(w, cookie)
			req.AddCookie(cookie)
		} else {
			uid = cookie.Value
		}
		ctx := context.WithValue(req.Context(), ContextKey("USERID"), uid)
		next.ServeHTTP(w, req.WithContext(ctx))
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
		s := http.StatusCreated
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
			var pqerr *pq.Error
			if errors.As(err, &pqerr) && pqerr.Code == pgerrcode.UniqueViolation {
				key, err = r.ls.Find(req.Context(), str)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				s = http.StatusConflict
			} else {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		u := &url.URL{
			Scheme: "http",
			Host:   r.ls.Host(),
			Path:   key,
		}
		ctx := context.WithValue(req.Context(), ContextKey("DATA"), u.String())
		ctx = context.WithValue(ctx, ContextKey("STATUS"), s)
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
			if _, errWriteString := io.WriteString(w, err.Error()); errWriteString != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
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
		lnks, err := r.ls.GetURLList(req.Context(), uid)
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
		ctx = context.WithValue(ctx, ContextKey("STATUS"), http.StatusOK)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (r *Router) Batch(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		uid, ok := req.Context().Value(ContextKey("USERID")).(string)
		if !ok || len(uid) == 0 {
			http.Error(w, "can't get user id", http.StatusInternalServerError)
			return
		}
		data := req.Context().Value(ContextKey("DATA"))
		if data == nil {
			http.Error(w, "can't get context data", http.StatusInternalServerError)
			return
		}
		str, ok := data.(string)
		if !ok {
			http.Error(w, "can't get context data", http.StatusInternalServerError)
			return
		}
		var batch []BatchItem
		err := json.Unmarshal([]byte(str), &batch)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res, err := r.ls.Batch(req.Context(), batch, uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json, err := json.Marshal(&res)
		if err != nil {
			http.Error(w, "can't marshal data", http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(req.Context(), ContextKey("DATA"), string(json))
		ctx = context.WithValue(ctx, ContextKey("STATUS"), http.StatusCreated)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (r *Router) SendPlainText(w http.ResponseWriter, req *http.Request) {
	status := req.Context().Value(ContextKey("STATUS"))
	if status == nil {
		http.Error(w, "can't get context data", http.StatusInternalServerError)
		return
	}
	statusInt, ok := status.(int)
	if !ok {
		http.Error(w, "can't get context data", http.StatusInternalServerError)
		return
	}
	data := req.Context().Value(ContextKey("DATA"))
	if data == nil {
		http.Error(w, "can't get context data", http.StatusInternalServerError)
		return
	}
	str, ok := data.(string)
	if !ok {
		http.Error(w, "can't get context data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.WriteHeader(statusInt)
	_, err := w.Write([]byte(str))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (r *Router) SendJSON(w http.ResponseWriter, req *http.Request) {
	status := req.Context().Value(ContextKey("STATUS"))
	if status == nil {
		status = http.StatusOK
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
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status.(int))
	_, err := w.Write([]byte(str))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (r *Router) Redirect(w http.ResponseWriter, req *http.Request) {
	key := path.Base(req.URL.Path)
	lnk, err := r.ls.Get(req.Context(), key)
	if err == sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusGone)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", lnk)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (r *Router) Ping(w http.ResponseWriter, req *http.Request) {
	err := r.ls.Ping(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (r *Router) DeleteUrls(w http.ResponseWriter, req *http.Request) {
	uid, ok := req.Context().Value(ContextKey("USERID")).(string)
	if !ok || len(uid) == 0 {
		http.Error(w, "can't get user id", http.StatusInternalServerError)
		return
	}
	data := req.Context().Value(ContextKey("DATA"))
	if data == nil {
		http.Error(w, "can't get context data", http.StatusInternalServerError)
		return
	}
	str, ok := data.(string)
	if !ok {
		http.Error(w, "can't get context data", http.StatusInternalServerError)
		return
	}
	var urls []string
	err := json.Unmarshal([]byte(str), &urls)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = r.ls.Delete(req.Context(), urls, uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
