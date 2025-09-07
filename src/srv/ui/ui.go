package ui

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

// pathExists returns whether or not the given file path exists as a file or
// directory
func pathExists(fsys fs.FS, path string) (bool, error) {
	_, err := fs.Stat(fsys, path)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, fs.ErrNotExist):
		return false, nil
	default:
		// only occurs when err != nil, but it's not a "does not exist error"
		// i.e. wtf?
		return false, err
	}
}

// vueFSWrap is a super basic wrapper around http.FileServer that sets the url
// request path to "/index.html" when the requested file does not exist within
// the wrapped filesystem. this way, we always return the vue root for urls that
// don't have an associated file and can let the vue router handle all routing
type vueFSWrap struct {
	fsys fs.FS
	fsrv http.Handler
}

func (v *vueFSWrap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if strings.HasPrefix(upath, "/") {
		upath = upath[1:]
	}
	if exists, err := pathExists(v.fsys, upath); err != nil || exists {
		// if an error occurs, default to the standard behavior
	} else {
		// but if the path doesn't exist, set the path to "/" to return the vue root
		r.URL.Path = "/"
	}
	v.fsrv.ServeHTTP(w, r)
}

//go:embed .webui/dist
var content embed.FS

const dir = ".webui/dist"

func Handler() (http.Handler, error) {
	const op = "ui.Handler"
	// Remove the root
	f, err := fs.Sub(content, dir)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to trim the embded filesystem holding UI elements: %w", op, err)
	}
	return &vueFSWrap{
		fsys: f,
		fsrv: http.FileServer(http.FS(f)),
	}, nil
}
