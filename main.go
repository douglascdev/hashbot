package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"hashbot/backend"
	"hashbot/command"
	"hashbot/config"
	"hashbot/database"
	"hashbot/hashbot"
	"hashbot/seventvapi"
	"hashbot/twitchapi"
	"os"
	"sort"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// parse command-line arguments
	cfgPath := flag.String("cfg", "config.hb", "path to config file")
	debug := flag.Bool("debug", false, "sets log level to debug")
	cmdListPrefix := flag.String("cmd-list-prefix", "\\", "sets the bot's prefix used in the command list generation")
	generateCmdList := flag.String("cmd-list", "", "ignores all other args and generates command list json to the specified path")
	flag.Parse()

	// set up logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.DateTime,
		},
	)

	if *debug {
		log.Debug().Msg("debug mode on")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		// Enable caller info
		log.Logger = log.With().Caller().Logger()
	}

	// generate command list json
	if *generateCmdList != "" {
		log.Info().Str("path", *generateCmdList).Msg("generating command list json")
		var (
			commandListData []byte
			err             error
		)
		// show commands on the list with the prefix
		for i := range command.Commands {
			if !command.Commands[i].NoPrefix {
				command.Commands[i].Name = *cmdListPrefix + command.Commands[i].Name
			}
		}

		sort.Sort(command.SortByPrefixAndName(command.Commands))

		commandListData, err = json.MarshalIndent(command.Commands, "", "  ")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to generate command list json")
		}
		if err := os.WriteFile(*generateCmdList, commandListData, 0644); err != nil {
			log.Fatal().Str("path", *generateCmdList).Err(err).Msg("failed to write command list json")
		}
		log.Info().Str("path", *generateCmdList).Msg("command list json generated successfully")
		os.Exit(0)
	}

	_, err := os.Stat(*cfgPath)
	if os.IsNotExist(err) {
		log.Warn().Str("path", *cfgPath).Msg("config file does not exist, creating from template")

		var file *os.File
		file, err = os.Create(*cfgPath)
		if err != nil {
			log.Fatal().Str("path", *cfgPath).Err(err).Msg("failed to create temaplate")
		}
		defer file.Close()

		var templateJSONBytes []byte
		templateJSONBytes, err = config.ConfigTemplateJSON()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to generate template")
		}

		_, err = file.Write(templateJSONBytes)
		if err != nil {
			log.Fatal().Str("path", *cfgPath).Err(err).Msg("failed to create template file")
		}

		log.Info().Str("path", *cfgPath).Msgf("template created successfully, please edit the file and run the bot again")
		os.Exit(0)
	}

	if err != nil {
		log.Fatal().Err(err).Msg("failed to stat config file")
	}

	var (
		cfg  *config.Config
		data []byte
	)
	data, err = os.ReadFile(*cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read config file")
	}
	cfg, err = config.LoadConfig(data)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config file")
	}

	reader := new(bytes.Buffer)
	reader.Write(data)

	db, err := database.InitDB(cfg.DBConfig.Driver, cfg.DBConfig.DataSourceName, reader)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize database")
	}

	defer db.Close()

	var (
		mb                 *hashbot.HashBot
		invalidatedTokenCh = make(chan bool, 1)
	)
	mb, err = hashbot.NewHashBot(cfg, db, invalidatedTokenCh)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize hashbot")
	}

	appCtx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	backend.RunServer(appCtx, cfg)
	hashbot.RunTokenValidator(appCtx, cfg, invalidatedTokenCh, mb.ConnectClient, twitchapi.RefreshTwitchToken, twitchapi.ValidateToken)
	hashbot.RunSevenTVEditorReqAccepter(appCtx, cfg, invalidatedTokenCh, twitchapi.GetUserByName, seventvapi.GetUserByConnection, seventvapi.AcceptEditorRequest)

	connectWithBreaker := hashbot.Breaker(mb.ReconnectClient, 5)
	for {
		err = connectWithBreaker()
		if errors.Is(err, hashbot.CircuitBreakerErr) {
			time.Sleep(time.Second / 4)
			continue
		} else if errors.Is(err, twitch.ErrClientDisconnected) {
			log.Warn().Msg("client disconnected")
		} else if err != nil {
			log.Error().Err(err).Msg("failed to connect to Twitch")
		}
	}
}
