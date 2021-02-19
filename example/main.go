package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/sardap/discom"
)

var (
	prefixPattern string = "discom\\$"
	discToken     string
	commandSet    *discom.CommandSet
)

func init() {
	discToken = strings.Replace(os.Getenv("DISCORD_AUTH"), "\"", "", -1)
	commandSet = discom.CreateCommandSet(regexp.MustCompile(prefixPattern), errorHandler)

	commandSet.AddCommand(discom.Command{
		Name: "say hi", Handler: hiCommandHandler,
		Description: "says hi",
	})

	commandSet.AddCommand(discom.Command{
		Name: "say_bye", Handler: hiCommandHandler,
		Description: "says bye",
	})
}

func errorHandler(s *discordgo.Session, m *discordgo.MessageCreate, err error) {
	s.ChannelMessageSend(
		m.ChannelID,
		fmt.Sprintf(
			"<@%s> invalid command:\"%s\" error:%s",
			m.Author.ID, m.Content, err,
		),
	)
}

func hiCommandHandler(s *discordgo.Session, m *discordgo.MessageCreate, args ...string) error {
	s.ChannelMessageSend(
		m.ChannelID,
		fmt.Sprintf("<@%s> Hi", m.Author.ID),
	)
	return nil
}

func main() {
	discord, err := discordgo.New("Bot " + discToken)
	if err != nil {
		log.Printf("unable to create new discord instance")
		log.Fatal(err)
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	discord.AddHandler(commandSet.Handler)

	// Open a websocket connection to Discord and begin listening.
	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	discord.UpdateStatus(1, "try discom$ help")

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}
