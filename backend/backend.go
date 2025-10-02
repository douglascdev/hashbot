package backend

import (
	"context"
	"hashbot/command"
	"hashbot/config"
	"hashbot/twitchapi"
	"hashbot/types"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
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

func Cmds(c echo.Context) error {
	acceptLanguage := c.Request().Header.Get("Accept-Language")

	lang := "en"
	// pt-BR;q=0.7,en;q=0.3 => pt-BR;q=0.7
	for s := range strings.SplitSeq(acceptLanguage, ",") {
		// pt-BR;q=0.7 => pt-BR
		before, _, _ := strings.Cut(s, ";")
		// pt-BR => pt
		language, _, _ := strings.Cut(before, "-")
		if slices.Contains(types.SupportedLanguages, language) {
			lang = language
			break
		}
	}

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile("active.pt.toml")

	localizer := i18n.NewLocalizer(bundle, lang)

	cfg := c.Get("config").(config.Config)
	commands := slices.Clone(command.Commands)
	sort.Sort(command.SortByPrefixAndName(commands))
	// show commands on the list with the prefix
	for i, cmd := range commands {
		if !cmd.NoPrefix {
			commands[i].Name = cfg.Prefix + cmd.Name
		}

		if cmd.GetLocalizedDescription != nil {
			commands[i].Description = cmd.GetLocalizedDescription(localizer)
		}
	}

	return c.JSONPretty(http.StatusOK, commands, "  ")
}

func RunServer(ctx context.Context, cfg *config.Config) {
	e := echo.New()

	e.Use(ConfigMiddleware(*cfg))
	e.Use(middleware.Recover())

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

	e.GET("/api/commands", Cmds)
	//e.GET("/login", loginHandler)
	//e.GET("/client_data", clientDataHandler)
	// e.Static("/static", "static")
	//	auth.InitAuth(e)
	//	routes.Router(e)

	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("shutting down the server")
		}
	}()

	go func() {
		<-ctx.Done() // Wait for cancellation
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := e.Shutdown(ctxShutdown); err != nil {
			log.Error().Err(err).Msg("server forced to shutdown")
		}
	}()

}
