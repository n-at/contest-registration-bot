package web

//from https://github.com/stnc/pongo4echo

import (
	"errors"
	"github.com/flosch/pongo2/v4"
	"github.com/labstack/echo/v4"
	"io"
)

type Pongo2Renderer struct {
	Debug bool
}

var (
	ctx pongo2.Context
	t   *pongo2.Template
	err error
)

// Render : Custom renderer
func (r Pongo2Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if data != nil {
		var ok bool
		ctx, ok = data.(pongo2.Context)

		if !ok {
			return errors.New("no pongo2.Context data was passed")
		}
	}

	if r.Debug {
		t, err = pongo2.FromFile(name)
	} else {
		t, err = pongo2.FromCache(name)
	}
	// Add some static values
	ctx["version_number"] = "v3.0"
	if err != nil {
		return err
	}
	return t.ExecuteWriter(ctx, w)
}
