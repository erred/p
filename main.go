package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	"github.com/go-logr/logr"
	"go.seankhliao.com/p/base20"
	"go.seankhliao.com/w/v16/render"
	"go.seankhliao.com/w/v16/webserver"
	"k8s.io/klog/v2/klogr"
)

func main() {
	var data string
	var wo webserver.Options
	flag.StringVar(&data, "data", "data", "path to data dir")
	wo.InitFlags(flag.CommandLine)
	flag.Parse()

	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	l := klogr.New()

	a := &App{
		l:   l,
		dir: data,
	}
	err := a.init()
	if err != nil {
		l.Error(err, "app init")
		os.Exit(1)
	}

	wo.Logger = l
	wo.Handler = a

	webserver.New(ctx, &wo).Run(ctx)
}

type App struct {
	l   logr.Logger
	dir string
}

func (a *App) init() error {
	err := os.MkdirAll(a.dir, 0o755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	ro := &render.Options{
		MarkdownSkip: true,
		Data: render.PageData{
			Title:       `paste`,
			Description: `simple paste service`,
			H1:          `paste`,
			H2:          `upload`,
			Style:       style,
			Compact:     true,
		},
	}

	b := bytes.NewReader([]byte(body))

	fn := filepath.Join(a.dir, "index.html")
	fout, err := os.Create(fn)
	if err != nil {
		return err
	}

	err = render.Render(ro, fout, b)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.l.Info("access", "method", r.Method, "path", r.URL.Path, "user-agent", r.UserAgent(), "remote", r.RemoteAddr)
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

	http.Redirect(w, r, "/"+filepath.Base(f.Name()), http.StatusSeeOther)
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

var body = `
<form id="form" autocomplete="off" action="/api/v0/form" enctype="multipart/form-data" method="POST"></form>

<div id="text">
<textarea
  id="paste"
  name="paste"
  form="form"
  autofocus
  placeholder="paste something here..."
  rows="10"
  cols="40"
></textarea>
</div>

<div id="buttons">
  <label for="upload" id="uploadlabel"><em>Or upload:</em></label>
  <input type="file" id="upload" name="upload" form="form" />
  <input type="submit" id="submit" value="Send" form="form" />
</div>

<script>
  const file = document.querySelector("#upload");
  file.addEventListener("change", (e) => {
    // Get the selected file
    const [file] = e.target.files;
    document.querySelector("#uploadlabel").innerHTML = "<em>Or upload:</em> ${file.name}";
  });
</script>
`

var style = `
#text,
#buttons {
  grid-column: 1 / span 5;
  margin: 1vmin 2vmin 0;
}
textarea {
  width: 100%;
  height: 60vh;
  margin: 1vh 0 2vh 0;
  font-size: 16px;
  background-color: #303030;
  color: #eceff1;
  border: none;
}
textarea:focus {
  outline: 3px solid #404040;
}

#upload {
  opacity: 0;
  width: 0.1px;
  height: 0.1px;
  position: absolute;
}
#uploadlabel {
  float: left;
  display: block;
  position: relative;
  cursor: pointer;
  border: 2px solid #303030;
  padding: 0.25em 1em;
}
#uploadlabel:hover {
  background-color: #303030;
  transition: 0.4s;
}
input[type="submit"] {
  background-color: #000;
  float: right;
  font: 18px "Inconsolata", monospace;
  font-weight: 700;
  padding: 0.25em 1em;
  color: #a06be0;
  border: 2px solid #a06be0;
}
input[type="submit"]:focus,
input[type="submit"]:hover {
  outline: none;
  background-color: #303030;
  transition: 0.4s;
}
`
