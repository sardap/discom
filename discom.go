package discom

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

//CommandHandler CommandHandler
type CommandHandler func(*discordgo.Session, *discordgo.MessageCreate)

//Command Command
type Command struct {
	Re          *regexp.Regexp
	Handler     CommandHandler
	Description string
}

//CommandSet CommandSet
type CommandSet struct {
	CaseInsensitive bool
	PrefixRe        *regexp.Regexp
	commands        []Command
}

var (
	helpRe *regexp.Regexp
)

func init() {
	helpRe = regexp.MustCompile("help")
}

func (c *Command) complete() bool {
	return c.Re != nil && c.Handler != nil && c.Re.String() != ""
}

//CreateCommandSet CreateCommandSet
func CreateCommandSet(CaseInsensitive bool, prefixRe *regexp.Regexp) *CommandSet {
	return &CommandSet{CaseInsensitive, prefixRe, make([]Command, 0)}
}

//AddCommand AddCommand
func (cs *CommandSet) AddCommand(com Command) error {
	if !com.complete() {
		return errors.New("Must set all fields in command struct")
	}

	cs.commands = append(cs.commands, com)
	return nil
}

func cleanPattern(pattern string) string {
	return strings.ReplaceAll(pattern, "\\", "")
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
		fmt.Fprintf(
			&result, "\"%s %s\" %s\n\n",
			cleanPattern(cs.PrefixRe.String()), cleanPattern(com.Re.String()), desc,
		)
	}

	return result.String()
}

//Handler Handler
func (cs *CommandSet) Handler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	msg := m.Content
	if cs.CaseInsensitive {
		msg = strings.ToLower(msg)
	}

	if !cs.PrefixRe.Match([]byte(msg)) {
		return
	}

	for _, com := range cs.commands {
		if com.Re.Match([]byte(msg)) {
			com.Handler(s, m)
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

func (cs *CommandSet) replyMessage(m *discordgo.MessageCreate, response string) string {
	return fmt.Sprintf("<@%s> %s", m.Author.ID, response)
}
