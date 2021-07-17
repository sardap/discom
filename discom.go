package discom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrTooManyArgs = fmt.Errorf("more args given then options")
	ErrInvalidArg  = fmt.Errorf("invalid arg")
)

type Response struct {
	Content string
}

type InteractionPayload struct {
	Message  string
	AuthorId string
	GuildId  string
}

// Interaction any interfaction with the commands
type Interaction interface {
	Respond(*discordgo.Session, Response) error
	GetPayload() *InteractionPayload
	Option(name string) *discordgo.ApplicationCommandInteractionDataOption
}

// CommandHandler A callback function which is triggered when a command is ran
// Error should only return data your fine with the user seeing
type CommandHandler func(*discordgo.Session, Interaction) error

// ErrorHandler called if a command handler returns an error
type ErrorHandler func(*discordgo.Session, Interaction, error)

// Command Represents a Command to the discord bot.
type Command struct {
	// Name the name of the command commands should not have spaces
	Name string
	// Handler The handler function which is called on a message matching the regex
	Handler     CommandHandler
	Description string
	Version     string
	Options     []*discordgo.ApplicationCommandOption
}

func (c *Command) asDiscordAppCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name,
		Description: c.Description,
		Version:     c.Version,
		Options:     c.Options,
	}
}

func (c *Command) parseArgs(args []string) ([]*discordgo.ApplicationCommandInteractionDataOption, error) {
	if len(args) == 0 {
		return nil, nil
	}

	if len(args)%2 != 0 {
		return nil, errors.Wrapf(ErrInvalidArg, "argument missing value")
	}

	argMap := make(map[string]string)

	optionsMap := make(map[string]*discordgo.ApplicationCommandOption)
	requiredCount := 0
	for _, option := range c.Options {
		optionsMap[option.Name] = option
		if option.Required {
			requiredCount++
		}
	}

	for i := 0; i < len(args); i += 2 {
		cmd := args[i]
		if !strings.HasPrefix(cmd, "-") {
			return nil, errors.Wrapf(ErrInvalidArg, "%s must have prefix -", cmd)
		}

		cmd = strings.TrimPrefix(cmd, "-")

		if _, ok := optionsMap[cmd]; !ok {
			return nil, errors.Wrapf(ErrInvalidArg, "%s is an unknown argument", cmd)
		}

		if optionsMap[cmd].Required {
			requiredCount--
		}

		argMap[cmd] = args[i+1]
	}

	if requiredCount != 0 {
		return nil, errors.Wrapf(ErrInvalidArg, "missing required argument")
	}

	var result []*discordgo.ApplicationCommandInteractionDataOption

	for cmd, arg := range argMap {
		var value interface{}
		var err error
		switch optionsMap[cmd].Type {
		case discordgo.ApplicationCommandOptionString:
			value = arg
		case discordgo.ApplicationCommandOptionInteger:
			intVal, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("expected %s to be an int but was given %s", optionsMap[cmd].Name, arg)
			}
			value = float64(intVal)
		case discordgo.ApplicationCommandOptionBoolean:
			value, err = strconv.ParseBool(arg)
			if err != nil {
				return nil, fmt.Errorf("expected %s to be an bool but was given %s", optionsMap[cmd].Name, arg)
			}
		default:
			return nil, fmt.Errorf("not implemnted yet")
		}

		result = append(result, &discordgo.ApplicationCommandInteractionDataOption{
			Name:  optionsMap[cmd].Name,
			Value: value,
		})
	}

	return result, nil
}

func genOptionsMap(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	result := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, option := range options {
		result[option.Name] = option
	}

	return result
}

type discordInteraction struct {
	sent        bool
	interaction *discordgo.Interaction
	optionsMap  map[string]*discordgo.ApplicationCommandInteractionDataOption
}

func (d *discordInteraction) Option(name string) *discordgo.ApplicationCommandInteractionDataOption {
	if d.optionsMap == nil {
		d.optionsMap = genOptionsMap(d.interaction.ApplicationCommandData().Options)
	}

	return d.optionsMap[name]
}

func (d *discordInteraction) Options() []*discordgo.ApplicationCommandInteractionDataOption {
	return d.interaction.ApplicationCommandData().Options
}

func (d *discordInteraction) GetPayload() *InteractionPayload {
	return &InteractionPayload{
		Message:  d.interaction.Message.Content,
		AuthorId: d.interaction.Member.User.ID,
		GuildId:  d.interaction.GuildID,
	}
}

func (d *discordInteraction) Respond(s *discordgo.Session, res Response) error {
	if !d.sent {
		err := s.InteractionRespond(d.interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: res.Content,
			},
		})
		d.sent = err == nil
		return err
	}

	return s.InteractionResponseEdit(s.State.User.ID, d.interaction, &discordgo.WebhookEdit{
		Content: res.Content,
	})
}

type discordMessage struct {
	message    *discordgo.Message
	sentId     string
	options    []*discordgo.ApplicationCommandInteractionDataOption
	optionsMap map[string]*discordgo.ApplicationCommandInteractionDataOption
}

func (d *discordMessage) Option(name string) *discordgo.ApplicationCommandInteractionDataOption {
	if d.optionsMap == nil {
		d.optionsMap = genOptionsMap(d.options)
	}

	return d.optionsMap[name]
}

func (d *discordMessage) GetPayload() *InteractionPayload {
	return &InteractionPayload{
		Message:  d.message.Content,
		AuthorId: d.message.Author.ID,
		GuildId:  d.message.GuildID,
	}
}

func (d *discordMessage) Respond(s *discordgo.Session, res Response) error {
	if d.sentId == "" {
		msg, err := s.ChannelMessageSend(d.message.ChannelID, res.Content)
		if err == nil {
			d.sentId = msg.ID
		}
		return err
	}

	_, err := s.ChannelMessageEdit(d.message.ChannelID, d.sentId, res.Content)
	return err
}

// CommandSet Use this to regsiter commands and get the handler to pass to discordgo.
// This should be created with CreateCommandSet.
type CommandSet struct {
	Prefix       string
	ErrorHandler ErrorHandler
	commands     []Command
	handlers     map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func (c *Command) valid() error {
	if c.Name == "" {
		return fmt.Errorf("invalid name is empty")
	}

	if strings.Contains(c.Name, " ") {
		return fmt.Errorf("invalid name conatins space")
	}

	if strings.ToLower(c.Name) == "help" {
		return fmt.Errorf("invalid name cannot be help")
	}

	if c.Handler == nil {
		return fmt.Errorf("invalid handler is nil")
	}

	requiredCompleted := false
	for _, option := range c.Options {
		if option.Name == "" {
			return fmt.Errorf("all options must have a name")
		}

		if requiredCompleted {
			if option.Required {
				return fmt.Errorf("requires must be sequential")
			}
		} else if !option.Required {
			requiredCompleted = true
		}
	}

	return nil
}

// CreateCommandSet Creates a command set
func CreateCommandSet(prefix string, errorHandler ErrorHandler) (*CommandSet, error) {
	if strings.Contains(prefix, " ") {
		return nil, fmt.Errorf("invlaid prefix contains space")
	}

	return &CommandSet{
		Prefix:       prefix,
		ErrorHandler: errorHandler,
		commands:     []Command{},
		handlers:     make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)),
	}, nil
}

func cleanPattern(pattern string) string {
	return strings.ReplaceAll(pattern, "\\", "")
}

func commandsEqual(a, b *discordgo.ApplicationCommand) bool {
	aJson, _ := json.Marshal(a.Options)
	bJson, _ := json.Marshal(b.Options)
	return a.Name == b.Name && a.Description == b.Description && bytes.Equal(aJson, bJson)
}

func (cs *CommandSet) SyncAppCommands(s *discordgo.Session) error {

	commands := make(map[string]Command)

	for _, cmd := range cs.commands {
		commands[cmd.Name] = cmd
	}

	existingCmds, _ := s.ApplicationCommands(s.State.User.ID, "")
	// delete deleted commandss
	for _, v := range existingCmds {
		if _, ok := commands[v.Name]; !ok {
			err := s.ApplicationCommandDelete(v.ApplicationID, "", v.ID)
			if err != nil {
				return errors.Wrapf(err, "unable to delete out of date command")
			}
		}
	}

	// Edit updated commands
	for _, v := range existingCmds {
		cmd := commands[v.Name]
		if _, ok := commands[v.Name]; ok {
			if !commandsEqual(v, cmd.asDiscordAppCommand()) {
				_, err := s.ApplicationCommandEdit(v.ApplicationID, "", v.ID, cmd.asDiscordAppCommand())
				if err != nil {
					return errors.Wrapf(err, "Cannot edit '%v' command: %v", v.Name, err)
				}
			}
			delete(commands, v.Name)
		}
	}

	// Create new commands
	for _, cmd := range cs.commands {
		if _, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd.asDiscordAppCommand()); err != nil {
			log.Fatalf("Cannot create '%v' command: %v", cmd, err)
		}
	}

	return nil
}

// AddCommand Use this to add a command to a command set
func (cs *CommandSet) AddCommand(com Command) error {
	if err := com.valid(); err != nil {
		return errors.Wrap(err, "invlaid command")
	}

	cs.commands = append(cs.commands, com)
	cs.handlers[com.Name] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		com.Handler(s, &discordInteraction{
			sent:        false,
			interaction: i.Interaction,
		})
	}

	return nil
}

// Handler Regsiter this with discordgo.AddHandler will be called every time a new message is sent on a guild.
func (cs *CommandSet) Handler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	msg := m.Content
	if len(msg) < len(cs.Prefix) {
		return
	}

	if msg[0:len(cs.Prefix)] != cs.Prefix {
		return
	}

	//Remove prefix from message
	msg = msg[len(cs.Prefix):]
	if len(msg) < 1 {
		s.ChannelMessageSend(
			m.ChannelID,
			cs.replyMessage(m, "Missing command argument"),
		)
		return
	}

	args := strings.Split(msg[1:], " ")
	if len(args) < 1 {
		s.ChannelMessageSend(
			m.ChannelID,
			cs.replyMessage(m, "Missing command argument"),
		)
		return
	}

	for _, cmd := range cs.commands {
		tmpMsg := args[0]
		if tmpMsg == cmd.Name {

			if len(args) > 1 {
				//Remove command from args list
				args = args[1:]
			} else {
				//Clear args
				args = make([]string, 0)
			}

			options, err := cmd.parseArgs(args)
			if err != nil {
				cs.ErrorHandler(s, &discordMessage{message: m.Message}, err)
				return
			}

			inter := &discordMessage{
				message: m.Message,
				options: options,
			}

			if err := cmd.Handler(s, inter); err != nil {
				cs.ErrorHandler(s, inter, err)
			}
			return
		}
	}

	var res string
	if strings.ToLower(args[0]) == "help" {
		res = cs.getHelpMessage()
	} else {
		res = fmt.Sprintf("unknown command try \"%s help\"", cs.Prefix)
	}

	s.ChannelMessageSend(
		m.ChannelID,
		cs.replyMessage(m, res),
	)
}

func (cs *CommandSet) IntreactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if h, ok := cs.handlers[i.ApplicationCommandData().Name]; ok {
		h(s, i)
	}
}

func (cs *CommandSet) getHelpMessage() string {
	var result strings.Builder
	fmt.Fprintf(&result, "here are all the commands I know\n")
	for _, com := range cs.commands {
		var desc string
		if com.Description != "" {
			desc = com.Description
		} else {
			desc = "missing description"
		}

		result.WriteString("\"")
		result.WriteString(cleanPattern(cs.Prefix))
		result.WriteString(" ")
		result.WriteString(com.Name)
		result.WriteString("\"")
		result.WriteString(" ")
		result.WriteString(desc)
		if len(com.Options) > 0 {
			result.WriteString(" options\n")
		}

		for _, option := range com.Options {
			result.WriteString("\t")
			result.WriteString(option.Name)
			result.WriteString(" ")
			result.WriteString(option.Description)
			result.WriteString(" required ")
			result.WriteString(strconv.FormatBool(option.Required))
			result.WriteString(" type ")
			result.WriteString(ApplicationCommandOptionToString(option.Type))
			result.WriteString("\n")
		}

		result.WriteString("\n\n")
	}

	return result.String()
}

func (cs *CommandSet) replyMessage(m *discordgo.MessageCreate, response string) string {
	return fmt.Sprintf("<@%s> %s", m.Author.ID, response)
}

func ApplicationCommandOptionToString(option discordgo.ApplicationCommandOptionType) string {
	switch option {
	case discordgo.ApplicationCommandOptionSubCommand:
		return "Sub Command"
	case discordgo.ApplicationCommandOptionSubCommandGroup:
		return "Command Grouup"
	case discordgo.ApplicationCommandOptionString:
		return "String"
	case discordgo.ApplicationCommandOptionInteger:
		return "Integer"
	case discordgo.ApplicationCommandOptionBoolean:
		return "Boolean"
	case discordgo.ApplicationCommandOptionUser:
		return "User"
	case discordgo.ApplicationCommandOptionChannel:
		return "Channel"
	case discordgo.ApplicationCommandOptionRole:
		return "Role"
	case discordgo.ApplicationCommandOptionMentionable:
		return "Mentionable"
	}

	return ""
}
