package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sardap/discom"
)

var (
	prefix     string = "discom$"
	discToken  string
	commandSet *discom.CommandSet
)

func errorHandler(s *discordgo.Session, i discom.Interaction, err error) {
	i.Respond(s, discom.Response{
		Content: fmt.Sprintf(
			"invalid command:\"%s\" error:%s",
			i.GetMessage(), err,
		),
	})
}

func hiCommandHandler(s *discordgo.Session, i discom.Interaction) error {
	i.Respond(s, discom.Response{
		Content: "processing",
	})

	i.Respond(s, discom.Response{
		Content: fmt.Sprintf("<@%s> Hi", i.GetAuthor()),
	})

	return nil
}

func main() {
	discToken = strings.Replace(os.Getenv("DISCORD_AUTH"), "\"", "", -1)
	commandSet, _ = discom.CreateCommandSet(prefix, errorHandler)

	commandSet.AddCommand(discom.Command{
		Name:        "say_hi",
		Handler:     hiCommandHandler,
		Description: "says hi",
	})

	commandSet.AddCommand(discom.Command{
		Name: "say_bye",
		Handler: func(s *discordgo.Session, i discom.Interaction) error {
			i.Respond(s, discom.Response{
				Content: fmt.Sprintf("<@%s> Bye", i.GetAuthor()),
			})

			return nil
		},
		Description: "says bye",
	})

	commandSet.AddCommand(discom.Command{
		Name: "option",
		Handler: func(s *discordgo.Session, i discom.Interaction) error {
			i.Respond(s, discom.Response{
				Content: fmt.Sprintf("<@%s> %s", i.GetAuthor(), i.Options()[0].StringValue()),
			})

			return nil
		},
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "bean",
				Description: "cool",
				Required:    true,
				Type:        discordgo.ApplicationCommandOptionString,
			},
		},
		Description: "says bye",
	})

	s, err := discordgo.New("Bot " + discToken)
	if err != nil {
		log.Fatal("unable to create new discord instance ", err)
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		s.UpdateListeningStatus("try discom$ or slash commands help")
		log.Println("Bot is up!")
	})

	s.AddHandler(commandSet.Handler)
	s.AddHandler(commandSet.IntreactionHandler)

	// Open a websocket connection to Discord and begin listening.
	if err := s.Open(); err != nil {
		log.Fatal("error opening connection,", err)
	}
	defer s.Close()

	commandSet.SyncAppCommands(s)

	// Wait here until CTRL-C or other term signal is received.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Gracefully shutdowning")
}
