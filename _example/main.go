package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TopiSenpai/dgo-paginator"
	"github.com/bwmarrin/discordgo"
)

var (
	Token = os.Getenv("TOKEN")
)

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	manager := paginator.NewManager()
	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(manager.OnInteractionCreate)

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || m.Content != "!test" {
			return
		}
		pages := []string{
			"page1",
			"page2",
			"page3",
		}
		if err = manager.CreateMessage(s, m.ChannelID, &paginator.Paginator{
			PageFunc: func(page int, embed *discordgo.MessageEmbed) {
				embed.Description = pages[page]
			},
			MaxPages:        len(pages),
			Expiry:          time.Now(),
			ExpiryLastUsage: true,
		}); err != nil {
			fmt.Println(err)
		}
	})

	if err = dg.Open(); err != nil {
		fmt.Println("error opening connection: ", err)
		return
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
