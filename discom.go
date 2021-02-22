package discom

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/pkg/errors"

	"github.com/bwmarrin/discordgo"
)

//CommandHandler A callback function which is triggered when a command is ran
//Error should only return data your fine with the user seeing
type CommandHandler func(*discordgo.Session, *discordgo.MessageCreate, ...string) error

//ErrorHandler called if a command handler returns an error
type ErrorHandler func(*discordgo.Session, *discordgo.MessageCreate, error)

//Command Represents a Command to the discord bot.
type Command struct {
	//Name the name of the command commands should not have spaces
	Name string
	//CaseInsensitive will to lower the incomming message before checking if it matches
	CaseInsensitive bool
	//Handler The handler function which is called on a message matching the regex
	Handler     CommandHandler
	Description string
	Example     string
}

//CommandSet Use this to regsiter commands and get the handler to pass to discordgo.
// This should be created with CreateCommandSet.
type CommandSet struct {
	Prefix       string
	ErrorHandler ErrorHandler
	commands     []Command
}

func init() {
}

func isLower(s string) bool {
	for _, r := range s {
		if !unicode.IsLower(r) && unicode.IsLetter(r) {
			return false
		}
	}
	return true
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

	if c.CaseInsensitive && !isLower(c.Name) {
		return fmt.Errorf("invalid name conatins uppercase while CaseInsensitive is true")
	}

	if c.Handler == nil {
		return fmt.Errorf("invalid handler is nil")
	}

	return nil
}

//CreateCommandSet Creates a command set
func CreateCommandSet(prefix string, errorHandler ErrorHandler) (*CommandSet, error) {
	if strings.Contains(prefix, " ") {
		return nil, fmt.Errorf("invlaid prefix contains space")
	}

	return &CommandSet{
		Prefix:       prefix,
		ErrorHandler: errorHandler,
		commands:     []Command{},
	}, nil
}

func cleanPattern(pattern string) string {
	return strings.ReplaceAll(pattern, "\\", "")
}

//AddCommand Use this to add a command to a command set
func (cs *CommandSet) AddCommand(com Command) error {
	if err := com.valid(); err != nil {
		return errors.Wrap(err, "invlaid command")
	}

	cs.commands = append(cs.commands, com)
	return nil
}

//Handler Regsiter this with discordgo.AddHandler will be called every time a new message is sent on a guild.
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

	for _, com := range cs.commands {
		tmpMsg := args[0]
		if com.CaseInsensitive {
			tmpMsg = strings.ToLower(tmpMsg)
		}

		if tmpMsg == com.Name {
			if len(args) > 1 {
				args = args[1:]
			}

			if err := com.Handler(s, m, args...); err != nil {
				cs.ErrorHandler(s, m, err)
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
		if com.Example != "" {
			result.WriteString(" ")
			result.WriteString(com.Example)
		}
		result.WriteString("\"")

		result.WriteString(
			fmt.Sprintf(" Case Insensitive? %t,", com.CaseInsensitive),
		)
		result.WriteString(" ")
		result.WriteString(desc)
		result.WriteString("\n\n")
	}

	return result.String()
}

func (cs *CommandSet) replyMessage(m *discordgo.MessageCreate, response string) string {
	return fmt.Sprintf("<@%s> %s", m.Author.ID, response)
}
