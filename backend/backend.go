package backend

import (
	"hashbot/config"
	"hashbot/twitchapi"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type TwitchData struct {
	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri"`
}

func ConfigMiddleware(config config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Store the config in the context
			c.Set("config", config)
			return next(c)
		}
	}
}

func clientDataHandler(c echo.Context) error {
	data := new(TwitchData)
	if err := c.Bind(data); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	config := c.Get("config").(config.Config)
	data.ClientID = config.ClientID
	data.RedirectURI = config.RedirectURI
	return c.JSON(http.StatusOK, data)
}

func loginHandler(c echo.Context) error {
	stateCookie, err := c.Cookie("loginState")
	if err != nil {
		return c.String(http.StatusNotAcceptable, "State not found")
	}

	params := c.QueryParams()
	if state, found := params["state"]; !found || len(state) != 1 || state[0] != stateCookie.Value {
		return c.String(http.StatusUnauthorized, "Invalid state")
	}

	var (
		token []string
		found bool
	)
	if token, found = params["code"]; !found || len(token) != 1 {
		return c.String(http.StatusBadRequest, "Authorization code not provided")
	}

	config := c.Get("config").(config.Config)
	result, err := twitchapi.AuthorizationCode(config.ClientID, config.ClientSecret, token[0], config.RedirectURI)
	if err != nil {
		return c.String(http.StatusUnauthorized, "failed to obtain authorization code")
	}

	c.SetCookie(&http.Cookie{Name: "accessToken", Value: result.AccessToken})
	c.SetCookie(&http.Cookie{Name: "refreshToken", Value: result.RefreshToken})

	return c.Redirect(http.StatusFound, "/")
}

func RunServer(cfg *config.Config) error {
	e := echo.New()

	e.Use(ConfigMiddleware(*cfg))

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

	e.GET("/login", loginHandler)
	e.GET("/client_data", clientDataHandler)
	// e.Static("/static", "static")
	//	auth.InitAuth(e)
	//	routes.Router(e)

	// try to start the server on port 1323 and if it fails show Error
	e.Start(":8080")
	return nil
}
