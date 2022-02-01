package web

import (
	"github.com/flosch/pongo2/v4"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func NewServer() *echo.Echo {
	e := echo.New()

	e.HideBanner = true
	e.Renderer = Pongo2Renderer{Debug: false}
	e.HTTPErrorHandler = httpErrorHandler
	e.Static("/assets", "assets")

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "templates/index.twig", pongo2.Context{
			"hello": "world",
		})
	})

	return e
}

// Custom HTTP error handler
func httpErrorHandler(e error, c echo.Context) {
	code := http.StatusInternalServerError
	if httpError, ok := e.(*echo.HTTPError); ok {
		code = httpError.Code
	}

	log.Errorf("http error: %s, method=%s, url=%s", e, c.Request().Method, c.Request().URL)

	err := c.Render(code, "templates/error.twig", pongo2.Context{
		"error": e,
	})
	if err != nil {
		log.Errorf("error page render error: %s", err)
	}
}
