package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/TopiSenpai/dgo-paginator"
	"github.com/bwmarrin/discordgo"
)

var token = os.Getenv("token")

func main() {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Println("error creating Discord session,", err)
		return
	}

	session.AddHandler(func(s *discordgo.Session, event *discordgo.MessageCreate) {
		if event.Author.ID == s.State.User.ID {
			return
		}

		if event.Content == "start" {
			go func() {
				ch, cls := paginator.NewEventCollector(s, func(s *discordgo.Session, e *discordgo.MessageCreate) bool {
					return event.Author.ID == e.Author.ID && event.ChannelID == e.ChannelID
				})
				defer cls()

				timer := time.NewTimer(time.Second * 5)
				defer timer.Stop()
				var messages []string
			loop:
				for {
					select {
					case <-timer.C:
						_, _ = s.ChannelMessageSend(event.ChannelID, "timed out")
						return
					case e := <-ch:
						messages = append(messages, e.Content)
						if len(messages) == 5 {
							break loop
						}
					}
				}
				_, _ = s.ChannelMessageSend(event.ChannelID, "messages collected: "+strings.Join(messages, ", "))
			}()
		}
	})

	session.Identify.Intents = discordgo.IntentsGuildMessages
	if err = session.Open(); err != nil {
		log.Println("error opening connection,", err)
		return
	}

	log.Println("example is now running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-s
}
