package discom

import (
	"regexp"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestHelpMessage(t *testing.T) {
	cs := CreateCommandSet(regexp.MustCompile("test\\$"))

	testHandler := func(*discordgo.Session, *discordgo.MessageCreate) {}

	err := cs.AddCommand(Command{
		Re:          regexp.MustCompile("nice"),
		Handler:     testHandler,
		Description: "nice a test handler",
	})
	if err != nil {
		t.Error(err)
	}

	cs.AddCommand(Command{
		Re:              regexp.MustCompile("very nice\\!"),
		Handler:         testHandler,
		Description:     "very nice a test handler",
		CaseInsensitive: true,
	})

	helpMsg := cs.getHelpMessage()

	if expect := "\"test$ nice\" Case Insensitive? false, nice a test handler"; !strings.Contains(helpMsg, expect) {
		t.Errorf("missmatch expected:%s to contain %s", helpMsg, expect)
	}

	if expect := "\"test$ very nice!\" Case Insensitive? true, very nice a test handler"; !strings.Contains(helpMsg, expect) {
		t.Errorf("missmatch expected:%s to contain %s", helpMsg, expect)
	}
}

func TestCallingHandler(t *testing.T) {
	cs := CreateCommandSet(regexp.MustCompile("test\\$"))

	called := false
	testHandler := func(*discordgo.Session, *discordgo.MessageCreate) {
		called = true
	}

	err := cs.AddCommand(Command{
		Re:          regexp.MustCompile("nice"),
		Handler:     testHandler,
		Description: "nice a test handler",
	})
	if err != nil {
		t.Error(err)
	}

	cs.AddCommand(Command{
		Re:              regexp.MustCompile("very nice\\!"),
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
	func() {
		defer func() { recover() }()

		testMessage.Content = msg
		cs.Handler(testSession, testMessage)

		t.Errorf("testHandler called by %s", msg)
	}()
	called = false

	msg = "test$ nice"
	testMessage.Content = msg
	cs.Handler(testSession, testMessage)

	if !called {
		t.Errorf("testHandler not called by %s", msg)
	}
	called = false

	msg = "test$ Nice"
	func() {
		defer func() { recover() }()

		testMessage.Content = msg
		cs.Handler(testSession, testMessage)

		t.Errorf("testHandler called by %s", msg)
	}()
	called = false

	msg = "test$ very nIce!"
	testMessage.Content = msg
	cs.Handler(testSession, testMessage)

	if !called {
		t.Errorf("testHandler called by %s should call case insensitive", msg)
	}
	called = false
}
