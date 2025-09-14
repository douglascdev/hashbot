package backend

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RunServer() error {
	e := echo.New()

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:       "./frontend/build",
		Browse:     true,
		IgnoreBase: true,
	}))

	// Custom 404 handler
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		if code == http.StatusNotFound {
			c.Redirect(http.StatusFound, "/404.html")
		} else {
			c.String(code, http.StatusText(code))
		}
	}
	// e.Static("/static", "static")
	//	auth.InitAuth(e)
	//	routes.Router(e)

	// try to start the server on port 1323 and if it fails show Error
	e.Start(":1323")
	return nil
}
