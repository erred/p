package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"go.seankhliao.com/p/base20"

	_ "embed"
)

//go:embed index.html
var indexHTML []byte

func main() {
	a := &App{
		dir: "data",
	}
	a.init()

	mux := http.NewServeMux()
	mux.Handle("/index.html", http.RedirectHandler("/", http.StatusFound))
	mux.Handle("/", a)
	log.Fatal(http.ListenAndServe("127.0.0.1:28002", mux))
}

type App struct {
	dir string
}

func (a *App) init() {
	err := os.MkdirAll(a.dir, 0o755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		log.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(a.dir, "index.html"), indexHTML, 0o644)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("ACCESS %s %s %s %s", r.Method, r.URL.Path, r.UserAgent(), r.RemoteAddr)
	if r.Method == http.MethodPost && r.URL.Path == "/api/v0/form" {
		a.formHandler(w, r)
		return
	}
	http.FileServer(http.Dir(a.dir)).ServeHTTP(w, r)
}

func (a *App) formHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1 << 25) // 32M
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	mpf, mph, err := r.FormFile("upload")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if err == nil {
		defer mpf.Close()

		ext := path.Ext(mph.Filename)
		f, err := uniqDir(a.dir, ext)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer f.Close()
		_, err = io.Copy(f, mpf)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/"+filepath.Base(f.Name()), http.StatusFound)
		return
	}

	p := r.FormValue("paste")
	if p == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "one of %q or %q is required", "upload", "paste")
		return
	}
	f, err := uniqDir(a.dir, ".txt")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()
	_, err = f.WriteString(p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/"+filepath.Base(f.Name()), http.StatusFound)
}

func uniqDir(dir, ext string) (*os.File, error) {
	totalRetries := 3
	for retries := totalRetries; retries > 0; retries-- {
		fn := base20.Encode(rand.Int63()) + ext
		f, err := os.OpenFile(filepath.Join(dir, fn), os.O_RDWR|os.O_CREATE, 0o644)
		switch {
		case err == nil:
			return f, nil
		case errors.Is(err, os.ErrExist):
			continue
		default:
			return nil, fmt.Errorf("create file=%s: %w", fn, err)
		}
	}
	return nil, fmt.Errorf("exceeded retries=%d", totalRetries)
}
