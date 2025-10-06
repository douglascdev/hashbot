package command

import (
	"database/sql"
	"errors"
	"fmt"
	"hashbot/config"
	"hashbot/database"
	"hashbot/types"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/rs/zerolog/log"
)

var Commands = []Command{
	ping,
	join,
	part,
	setLevel,
	buttsbot,
	butt,
	help,
	explore,
	enable,
	disable,
	optout,
	optin,
	remove,
	editor,
	add,
	yoink,
	language,
	buttword,
	location,
	cmdTime,
}

var UnknownCommandErr = errors.New("unknown command")

var (
	commandMap       map[string]Command
	commandsNoPrefix []Command
)

type Command struct {
	Name            string
	Aliases         []string
	Usage           string
	Description     string
	ChannelCooldown int
	UserCooldown    int
	NoPrefix        bool
	CanDisable      bool

	// `json:"-"` excludes these fields from being serialized into the command list json
	ValidUsage              func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) bool  `json:"-"`
	NoPrefixShouldRun       func(message *types.Message, sender types.MessageSender, args []string) bool            `json:"-"`
	Execute                 func(message *types.Message, sender types.MessageSender, args []string) error           `json:"-"`
	ExecuteParsed           func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) error `json:"-"`
	GetLocalizedDescription func(localizer *i18n.Localizer) string                                                  `json:"-"`
}

type SortByPrefixAndName []Command

func (a SortByPrefixAndName) Len() int      { return len(a) }
func (a SortByPrefixAndName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortByPrefixAndName) Less(i, j int) bool {
	if a[i].Name == a[j].Name {
		return a[i].NoPrefix && !a[j].NoPrefix
	}
	return a[i].Name < a[j].Name
}

func init() {
	commandMap = createCommandMap(Commands)

	for _, cmd := range Commands {
		if cmd.NoPrefix {
			commandsNoPrefix = append(commandsNoPrefix, cmd)
		}
	}
}

// Maps command names and aliases to Command structs
// If prefixedOnly is true, only commands with NoPrefix=false will be added
func createCommandMap(commands []Command) map[string]Command {
	cmdMap := make(map[string]Command)
	for _, cmd := range commands {
		if cmd.NoPrefix {
			continue
		}
		cmdMap[cmd.Name] = cmd
		for _, alias := range cmd.Aliases {
			cmdMap[alias] = cmd
		}
	}
	return cmdMap
}

type commandData struct {
	isCmdEnabled           bool
	isCmdOnChannelCoolDown bool
	isCmdOnUserCoolDown    bool
	isUserIgnored          bool
	isOptedOut             bool
}

func getCommandData(tx *sql.Tx, message *types.Message, cmd Command) (*commandData, error) {
	// TODO: turn selects into separate goroutines after migrating to postgres
	result := &commandData{
		isCmdEnabled:           false,
		isCmdOnChannelCoolDown: false,
		isCmdOnUserCoolDown:    false,
		isUserIgnored:          false,
		isOptedOut:             false,
	}

	var err error

	if !cmd.CanDisable {
		result.isCmdEnabled = true
	} else {
		result.isCmdEnabled, err = database.SelectIsUserCommandEnabled(tx, message.RoomID, cmd.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to select is_user_command_enabled: %w", err)
		}
	}

	result.isUserIgnored, err = database.SelectIsUserIgnored(tx, message.Chatter.ID)
	if err == sql.ErrNoRows {
		result.isUserIgnored = false
	} else if err != nil {
		return nil, fmt.Errorf("failed to select user's is_ignored: %w", err)
	}

	result.isCmdOnChannelCoolDown, err = database.SelectIsCommandOnChannelCooldown(tx, message.RoomID, cmd.Name, cmd.ChannelCooldown)
	if err != nil {
		return nil, fmt.Errorf("failed to select command cooldown: %w", err)
	}

	result.isCmdOnUserCoolDown, err = database.SelectIsCommandOnUserCooldown(tx, message.Chatter.ID, cmd.Name, cmd.UserCooldown)
	if err != nil {
		return nil, fmt.Errorf("failed to select user cooldown: %w", err)
	}

	result.isOptedOut, err = database.SelectIsCommandOptedOut(tx, message.Chatter.ID, cmd.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to select user's opted_out: %w", err)
	}

	return result, nil
}

func HandleCommands(message *types.Message, sender types.MessageSender, config *config.Config) error {
	var (
		cmdData *commandData
		args    []string
		tx      *sql.Tx
		err     error
	)

	tx, err = message.DB.Begin()
	if err != nil {
		log.Err(err).Msg("failed to start HandleCommands transaction")
		return err
	}
	defer tx.Rollback()

	hasPrefix := strings.HasPrefix(message.Message, config.Prefix)
	if hasPrefix {
		args = strings.Split(message.Message[len(config.Prefix):], " ")
	} else {
		args = strings.Split(message.Message, " ")

		// check if command is no prefix
		for _, noPrefixCmd := range commandsNoPrefix {
			if noPrefixCmd.NoPrefixShouldRun != nil && noPrefixCmd.NoPrefixShouldRun(message, sender, args) {
				// use the channel's user language as the defult for the new user
				channelUserLanguage, err := database.SelectUserLanguage(tx, message.RoomID)
				if err != nil {
					log.Err(err).Str("channel", message.Channel).Str("user", message.Chatter.Name).Msg("failed to select channel's language")
				}

				err = database.InsertUsers(tx, false, channelUserLanguage, struct{ ID, Name string }{message.Chatter.ID, message.Chatter.Name})
				if err != nil {
					return err
				}

				cmdData, err = getCommandData(tx, message, noPrefixCmd)
				if err != nil {
					return err
				}
				if !cmdData.isCmdEnabled {
					log.Debug().Str("command", noPrefixCmd.Name).Str("channel", message.Channel).Msg("ignored disabled no-prefix command")
					return nil
				}

				if cmdData.isUserIgnored {
					log.Debug().Str("user", message.Chatter.Name).Str("channel", message.Channel).Msg("ignored user")
					return nil
				}

				if cmdData.isCmdOnChannelCoolDown {
					log.Debug().Str("command", noPrefixCmd.Name).Str("channel", message.Channel).Msg("command ignored due to channel cooldown")
					return nil
				}

				if cmdData.isCmdOnUserCoolDown {
					log.Debug().Str("command", noPrefixCmd.Name).Str("channel", message.Channel).Msg("command ignored due to user command cooldown")
					return nil
				}

				if cmdData.isOptedOut {
					log.Debug().Str("command", noPrefixCmd.Name).Str("channel", message.Channel).Msg("command ignored due to opt out")
					return nil
				}

				err = database.UpdateUserCommandLastUsed(tx, message.RoomID, noPrefixCmd.Name, message.Chatter.ID)
				if err != nil {
					return fmt.Errorf("failed to update last_used for command %s: %w", noPrefixCmd.Name, err)
				}

				err = tx.Commit()
				if err != nil {
					return fmt.Errorf("failed to commit transaction to update last_used for command %s: %w", noPrefixCmd.Name, err)
				}

				if noPrefixCmd.ExecuteParsed == nil {
					err = noPrefixCmd.Execute(message, sender, args)
				} else {
					// fills the command variable since the first arg here isnt a command name
					parsedArgs := ParseArgs("noPrefix " + message.Message)
					err = noPrefixCmd.ExecuteParsed(message, sender, parsedArgs)
				}
				if err != nil {
					return err
				}

				break
			}
		}

		return nil
	}

	if cmd, ok := commandMap[args[0]]; ok {
		// use the channel's user language as the defult for the new user
		channelUserLanguage, err := database.SelectUserLanguage(tx, message.RoomID)
		if err != nil {
			log.Err(err).Str("channel", message.Channel).Str("user", message.Chatter.Name).Msg("failed to select channel's language")
		}

		err = database.InsertUsers(tx, false, channelUserLanguage, struct{ ID, Name string }{message.Chatter.ID, message.Chatter.Name})
		if err != nil {
			return err
		}

		cmdData, err = getCommandData(tx, message, cmd)
		if err != nil {
			return err
		}
		if !cmdData.isCmdEnabled {
			log.Debug().Str("command", cmd.Name).Str("channel", message.Channel).Msg("ignored disabled command")
			return nil
		}

		if cmdData.isUserIgnored {
			log.Debug().Str("user", message.Chatter.Name).Str("channel", message.Channel).Msg("ignored user")
			return nil
		}

		if cmdData.isCmdOnChannelCoolDown {
			log.Debug().Str("command", cmd.Name).Str("channel", message.Channel).Msg("command ignored due to channel cooldown")
			return nil
		}

		if cmdData.isCmdOnUserCoolDown {
			log.Debug().Str("command", cmd.Name).Str("channel", message.Channel).Msg("command ignored due to user command cooldown")
			return nil
		}

		if cmdData.isOptedOut {
			log.Debug().Str("command", cmd.Name).Str("channel", message.Channel).Msg("command ignored due to opt out")
			return nil
		}

		err = database.UpdateUserCommandLastUsed(tx, message.RoomID, cmd.Name, message.Chatter.ID)
		if err != nil {
			return fmt.Errorf("failed to update last_used for command %s: %w", cmd.Name, err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction to update last_used for command %s: %w", cmd.Name, err)
		}

		parsedArgs := ParseArgs(message.Message)

		if cmd.ValidUsage != nil && !cmd.ValidUsage(message, sender, parsedArgs) {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "Usage",
					Other: "❌Usage: {{.Usage}}",
				},
				TemplateData: map[string]string{
					"Usage": cmd.Usage,
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return nil
		}

		if cmd.Execute == nil {
			if err = cmd.ExecuteParsed(message, sender, parsedArgs); err != nil {
				return err
			}
		} else if err = cmd.Execute(message, sender, args); err != nil {
			return err
		}

	} else if hasPrefix {
		return fmt.Errorf("%w: '%s' called by '%s'", UnknownCommandErr, args, message.Chatter.Name)
	}

	return nil
}
