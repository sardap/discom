package discom

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

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
	PrefixRe     *regexp.Regexp
	ErrorHandler ErrorHandler
	commands     []Command
}

var (
	helpRe *regexp.Regexp
)

func init() {
	helpRe = regexp.MustCompile("help")
}

func (c *Command) complete() bool {
	return c.Name != "" && c.Handler != nil
}

//CreateCommandSet Creates a command set
func CreateCommandSet(prefixRe *regexp.Regexp, errorHandler ErrorHandler) *CommandSet {
	return &CommandSet{
		PrefixRe:     prefixRe,
		ErrorHandler: errorHandler,
		commands:     []Command{},
	}
}

func cleanPattern(pattern string) string {
	return strings.ReplaceAll(pattern, "\\", "")
}

//AddCommand Use this to add a command to a command set
func (cs *CommandSet) AddCommand(com Command) error {
	if !com.complete() {
		return errors.New("Must set all fields in command struct")
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

	if !cs.PrefixRe.Match([]byte(msg)) {
		return
	}

	msg = string(cs.PrefixRe.ReplaceAll([]byte(msg), []byte("")))

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
	if helpRe.Match([]byte(msg)) {
		res = cs.getHelpMessage()
	} else {
		res = fmt.Sprintf("unknown command try \"%s help\"", cs.PrefixRe.String())
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
		result.WriteString(cleanPattern(cs.PrefixRe.String()))
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
