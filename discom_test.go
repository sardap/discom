package discom

import (
	"fmt"
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

	assert.NoError(t, cs.AddCommand(Command{
		Name:        "nice",
		Handler:     testHandler,
		Description: "nice a test handler",
	}))

	assert.NoError(t, cs.AddCommand(Command{
		Name:        "very_nice!",
		Handler:     testHandler,
		Description: "very nice a test handler",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "required_flag",
				Description: "required flag",
				Required:    true,
				Type:        discordgo.ApplicationCommandOptionString,
			},
			{
				Name:        "optional_flag",
				Description: "optional flag",
				Required:    false,
				Type:        discordgo.ApplicationCommandOptionInteger,
			},
		},
	}))

	helpMsg := cs.getHelpMessage()

	assert.Contains(t, helpMsg, `"test$ nice" nice a test handler`)

	assert.Contains(t, helpMsg, `"test$ very_nice!" very nice a test handler options`)
	assert.Contains(t, helpMsg, `required_flag required flag required true type String`)
	assert.Contains(t, helpMsg, `optional_flag optional flag required false type Integer`)
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
		if i.Option("moive").StringValue() != "bee" {
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

func TestArgs(t *testing.T) {
	cs, _ := CreateCommandSet("test$", func(*discordgo.Session, Interaction, error) {
	})

	var b bool
	var s string
	var i int64
	var opt string
	testHandler := func(sess *discordgo.Session, inter Interaction) error {
		b = inter.Option("bool").BoolValue()
		s = inter.Option("string").StringValue()
		i = inter.Option("int").IntValue()
		if inter.Option("optional") != nil {
			opt = inter.Option("optional").StringValue()
		}
		return nil
	}

	cs.AddCommand(Command{
		Name:        "nice",
		Handler:     testHandler,
		Description: "nice a test handler",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "bool",
				Description: "bool",
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Required:    true,
			},
			{
				Name:        "string",
				Description: "string",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "int",
				Description: "int",
				Type:        discordgo.ApplicationCommandOptionInteger,
				Required:    true,
			},
			{
				Name:        "optional",
				Description: "optional",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    false,
			},
		},
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

	// test mandatory
	msg := "test$ nice -bool true -string paul -int 69"
	testMessage.Content = msg
	cs.Handler(testSession, testMessage)
	assert.Equal(t, true, b)
	assert.Equal(t, "paul", s)
	assert.Equal(t, int64(69), i)
	assert.Equal(t, "", opt)

	// One arg should be passed
	msg = "test$ nice -bool true -string paul -int 69 -optional cool"
	testMessage.Content = msg
	cs.Handler(testSession, testMessage)
	assert.Equal(t, "cool", opt)
}
