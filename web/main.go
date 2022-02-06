package web

import (
	"contest-registration-bot/bot"
	"github.com/flosch/pongo2/v4"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Configuration struct {
	DebugTemplates bool
	Listen         string
}

var (
	registrationBot *bot.Bot
)

///////////////////////////////////////////////////////////////////////////////

func NewServer(configuration Configuration, b *bot.Bot) *echo.Echo {
	registrationBot = b

	e := echo.New()

	e.HideBanner = true
	e.Renderer = Pongo2Renderer{Debug: configuration.DebugTemplates}
	e.HTTPErrorHandler = httpErrorHandler
	e.Static("/assets", "assets")

	e.GET("/", contestsGet)
	e.GET("/contest", contestNew)
	e.GET("/contest/:id", contestGet)
	e.POST("/contest", contestSave)
	e.POST("/contest/:id/hide", contestHide)
	e.POST("/contest/:id/show", contestShow)
	e.POST("/contest/:id/close", contestClose)
	e.POST("/contest/:id/open", contestOpen)

	e.GET("/contest/:id/participants", participantsList)
	e.GET("/contest/:id/participants/export", participantsExport)
	e.GET("/contest/:id/participant", participantNew)
	e.GET("/contest/:id/participant/:participant_id", participantEdit)
	e.POST("/contest/:id/participant", participantSave)
	e.POST("/contest/:id/participant/:participant_id/delete", participantDelete)

	e.GET("/contest/:id/notifications", contestNotifications)
	e.GET("/contest/:id/notification", contestNotificationNew)
	e.GET("/contest/:id/notification/:notification_id", contestNotificationEdit)
	e.POST("/contest/:id/notification", contestNotificationSave)
	e.POST("/contest/:id/notification/:notification_id/delete", contestNotificationDelete)

	return e
}

///////////////////////////////////////////////////////////////////////////////

// httpErrorHandler Custom HTTP error handler
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
