package backend

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func login(c echo.Context) error {
	stateCookie, err := c.Cookie("loginState")
	if err != nil {

		return c.String(http.StatusNotAcceptable, "State not found")
	}
	params := c.QueryParams()
	if state, found := params["state"]; !found || len(state) != 1 || state[0] != stateCookie.Value {

		return c.String(http.StatusUnauthorized, "Invalid state")
	}
	if token, found := params["code"]; !found || len(token) != 1 {
		return c.String(http.StatusBadRequest, "Authorization code not provided")
	} else {
		loginCookie := new(http.Cookie)
		loginCookie.Name = "twitchToken"
		loginCookie.Value = token[0]
		c.SetCookie(loginCookie)
	}
	return c.String(http.StatusOK, "OK")
}

func RunServer() error {
	e := echo.New()

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root: "./frontend/build",
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

	e.GET("/login", login)
	// e.Static("/static", "static")
	//	auth.InitAuth(e)
	//	routes.Router(e)

	// try to start the server on port 1323 and if it fails show Error
	e.Start(":8080")
	return nil
}
