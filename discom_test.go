package discom

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestHelpMessage(t *testing.T) {
	cs := CreateCommandSet(
		regexp.MustCompile("test\\$"),
		func(*discordgo.Session, *discordgo.MessageCreate, error) {},
	)

	testHandler := func(*discordgo.Session, *discordgo.MessageCreate, ...string) error {
		return nil
	}

	err := cs.AddCommand(Command{
		Name:        "nice",
		Handler:     testHandler,
		Description: "nice a test handler",
		Example:     "test",
	})
	if err != nil {
		t.Error(err)
	}

	cs.AddCommand(Command{
		Name:            "very_nice!",
		Handler:         testHandler,
		Description:     "very nice a test handler",
		CaseInsensitive: true,
	})
	if err != nil {
		t.Error(err)
	}

	helpMsg := cs.getHelpMessage()

	if expect := "\"test$ nice test\" Case Insensitive? false, nice a test handler"; !strings.Contains(helpMsg, expect) {
		t.Errorf("missmatch expected:%s to contain %s", helpMsg, expect)
	}

	if expect := "\"test$ very_nice!\" Case Insensitive? true, very nice a test handler"; !strings.Contains(helpMsg, expect) {
		t.Errorf("missmatch expected:%s to contain %s", helpMsg, expect)
	}
}

func TestCallingHandler(t *testing.T) {
	cs := CreateCommandSet(regexp.MustCompile("test\\$"), func(*discordgo.Session, *discordgo.MessageCreate, error) {})

	called := false
	testHandler := func(*discordgo.Session, *discordgo.MessageCreate, ...string) error {
		called = true
		return nil
	}

	err := cs.AddCommand(Command{
		Name:        "nice",
		Handler:     testHandler,
		Description: "nice a test handler",
	})
	if err != nil {
		t.Error(err)
	}

	cs.AddCommand(Command{
		Name:            "very_nice!",
		Handler:         testHandler,
		Description:     "very nice a test handler",
		CaseInsensitive: true,
	})

	testSession := &discordgo.Session{
		State: &discordgo.State{
			Ready: discordgo.Ready{
				User: &discordgo.User{
					ID: "botID",
				},
			},
		},
	}

	testMessage := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID: "messagerID",
			},
		},
	}

	msg := "test$ invalid"
	testMessage.Content = msg
	assert.Panics(
		t, func() { cs.Handler(testSession, testMessage) },
		"The code did not panic",
	)
	called = false

	msg = "test$ nice"
	testMessage.Content = msg
	cs.Handler(testSession, testMessage)

	if !called {
		t.Errorf("testHandler not called by %s", msg)
	}
	called = false

	msg = "test$ Nice"
	testMessage.Content = msg
	assert.Panics(
		t, func() {
			cs.Handler(testSession, testMessage)
		},
		"The code did not panic",
	)
	called = false

	msg = "test$ very_nIce!"
	testMessage.Content = msg
	cs.Handler(testSession, testMessage)

	if !called {
		t.Errorf("testHandler called by %s should call case insensitive", msg)
	}
	called = false
}

func TestErrorHandler(t *testing.T) {
	called := false
	cs := CreateCommandSet(regexp.MustCompile("test\\$"), func(*discordgo.Session, *discordgo.MessageCreate, error) {
		called = true
	})

	testHandler := func(s *discordgo.Session, m *discordgo.MessageCreate, args ...string) error {
		if args[0] == "bee" {
			return fmt.Errorf("hey cool")
		}
		return nil
	}

	err := cs.AddCommand(Command{
		Name:        "nice",
		Handler:     testHandler,
		Description: "nice a test handler",
	})
	if err != nil {
		t.Error(err)
	}

	testSession := &discordgo.Session{
		State: &discordgo.State{
			Ready: discordgo.Ready{
				User: &discordgo.User{
					ID: "botID",
				},
			},
		},
	}

	testMessage := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID: "messagerID",
			},
		},
	}

	testMessage.Content = "test$ nice bee moive"
	cs.Handler(testSession, testMessage)
	if !called {
		t.Errorf("error handler not called")
	}
	called = false

	testMessage.Content = "test$ nice moive"
	cs.Handler(testSession, testMessage)
	if called {
		t.Errorf("error handler called")
	}
	called = false
}
