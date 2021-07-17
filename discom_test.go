package discom

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestHelpMessage(t *testing.T) {
	cs, _ := CreateCommandSet(
		"test$",
		func(*discordgo.Session, Interaction, error) {},
	)

	testHandler := func(*discordgo.Session, Interaction) error {
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
		Name:        "very_nice!",
		Handler:     testHandler,
		Description: "very nice a test handler",
	})
	if err != nil {
		t.Error(err)
	}

	helpMsg := cs.getHelpMessage()

	if expect := `"test$ nice" nice a test handler`; !strings.Contains(helpMsg, expect) {
		t.Errorf("missmatch expected:%s to contain %s", helpMsg, expect)
	}

	if expect := `"test$ very_nice!" very nice a test handler`; !strings.Contains(helpMsg, expect) {
		t.Errorf("missmatch expected:%s to contain %s", helpMsg, expect)
	}
}

func TestCallingHandler(t *testing.T) {
	cs, _ := CreateCommandSet("test$", func(*discordgo.Session, Interaction, error) {})

	called := false
	testHandler := func(*discordgo.Session, Interaction) error {
		called = true
		return nil
	}

	// Standard Call
	err := cs.AddCommand(Command{
		Name:        "nice",
		Handler:     testHandler,
		Description: "nice a test handler",
	})
	assert.NoError(t, err)

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

	// Test valid call
	msg := "test$ nice"
	testMessage.Content = msg
	cs.Handler(testSession, testMessage)
	assert.Truef(t, called, "testHandler not called by %s", msg)
	called = false

	// Test calling with no args
	msg = "test$"
	testMessage.Content = msg
	assert.Panics(
		t, func() {
			cs.Handler(testSession, testMessage)
		},
		"The code did not panic",
	)
	called = false
}

func TestErrorHandler(t *testing.T) {
	errCalled := false
	cs, _ := CreateCommandSet("test$", func(*discordgo.Session, Interaction, error) {
		errCalled = true
	})

	testHandler := func(s *discordgo.Session, i Interaction) error {
		if i.Options()[0].StringValue() != "bee" {
			return fmt.Errorf("hey cool")
		}
		return nil
	}

	err := cs.AddCommand(Command{
		Name:        "nice",
		Handler:     testHandler,
		Description: "nice a test handler",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "moive",
				Description: "bee moive",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	})
	assert.NoError(t, err)

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

	// No error
	testMessage.Content = "test$ nice -moive bee"
	cs.Handler(testSession, testMessage)
	assert.False(t, errCalled)
	errCalled = false

	// Error
	testMessage.Content = "test$ nice -moive bae"
	cs.Handler(testSession, testMessage)
	assert.True(t, errCalled)
	errCalled = false
}

func TestValid(t *testing.T) {
	cs, _ := CreateCommandSet("test$", func(*discordgo.Session, Interaction, error) {
	})

	testHandler := func(*discordgo.Session, Interaction) error {
		return nil
	}

	var err error
	//Fails since Name is empty
	err = cs.AddCommand(Command{
		Handler:     testHandler,
		Description: "nice a test handler",
	})
	if err == nil {
		t.Error("error was nil")
	}

	//Fails since name conatins space
	err = cs.AddCommand(Command{
		Name:        "nice one",
		Handler:     testHandler,
		Description: "nice a test handler",
	})
	if err == nil {
		t.Error("error was nil")
	}

	//Fails since name is help
	err = cs.AddCommand(Command{
		Name:        "help",
		Handler:     testHandler,
		Description: "nice a test handler",
	})
	if err == nil {
		t.Error("error was nil")
	}
}

// func TestArgs(t *testing.T) {
// 	cs, _ := CreateCommandSet("test$", func(*discordgo.Session, *discordgo.MessageCreate, error) {
// 	})

// 	argCount := 0
// 	testHandler := func(s *discordgo.Session, m *discordgo.MessageCreate, args ...string) error {
// 		argCount = len(args)
// 		return nil
// 	}

// 	cs.AddCommand(Command{
// 		Name:        "nice",
// 		Handler:     testHandler,
// 		Description: "nice a test handler",
// 	})

// 	testSession := &discordgo.Session{
// 		State: &discordgo.State{
// 			Ready: discordgo.Ready{
// 				User: &discordgo.User{
// 					ID: "botID",
// 				},
// 			},
// 		},
// 	}

// 	testMessage := &discordgo.MessageCreate{
// 		Message: &discordgo.Message{
// 			Author: &discordgo.User{
// 				ID: "messagerID",
// 			},
// 		},
// 	}

// 	//No args should be passed
// 	msg := "test$ nice"
// 	testMessage.Content = msg
// 	cs.Handler(testSession, testMessage)
// 	if argCount > 0 {
// 		t.Errorf("arg passed when none should be")
// 	}
// 	argCount = 0

// 	//One arg should be passed
// 	msg = "test$ nice very"
// 	testMessage.Content = msg
// 	cs.Handler(testSession, testMessage)
// 	if argCount != 1 {
// 		t.Errorf("1 arg should have been passed")
// 	}
// 	argCount = 0
// }
