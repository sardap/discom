package discom

import (
	"regexp"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestHelpMessage(t *testing.T) {
	cs := CreateCommandSet(false, regexp.MustCompile("test\\$"))

	testHandler := func(s *discordgo.Session, m *discordgo.Message) {}

	err := cs.AddCommand(Command{
		regexp.MustCompile("nice"), testHandler, "nice a test handler",
	})
	if err != nil {
		t.Error(err)
	}

	cs.AddCommand(Command{
		regexp.MustCompile("very nice\\!"), testHandler, "very nice a test handler",
	})

	expected1 := "\"test$ nice\" nice a test handler"
	expected2 := "\"test$ nice\" very nice a test handler"
	if msg := cs.getHelpMessage(); strings.Contains(msg, expected1) && strings.Contains(msg, expected2) {
		t.Errorf("missmatch expected %s %s got %s", expected1, expected2, msg)
	}
}
