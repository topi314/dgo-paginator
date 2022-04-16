package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
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

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || m.Content != "!test2" {
			return
		}
		msg, _ := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content: "press the button within 10s",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.Button{
							Label:    "click me",
							Style:    discordgo.PrimaryButton,
							CustomID: "click_me:" + m.Message.ID,
						},
					},
				},
			},
		})
		go func() {
			eventChannel, closeFunc := paginator.NewEventCollector(s, func(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
				if i.Type != discordgo.InteractionMessageComponent {
					return false
				}
				data := strings.Split(i.MessageComponentData().CustomID, ":")
				if data[0] != "click_me" {
					return false
				}
				return data[1] == m.Message.ID
			})
			defer closeFunc()

			timer := time.NewTimer(time.Second * 10)
			defer timer.Stop()
			select {
			case i := <-eventChannel:
				fmt.Println("someone pressed the button!")
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseUpdateMessage,
					Data: &discordgo.InteractionResponseData{
						Content:    fmt.Sprintf("<@%s> pressed me first!", i.Member.User.ID),
						Components: []discordgo.MessageComponent{},
					},
				})
			case <-timer.C:
				content := "too slow!"
				_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
					Content:    &content,
					Components: []discordgo.MessageComponent{},
					ID:         msg.ID,
					Channel:    msg.ChannelID,
				})
				if err != nil {
					fmt.Println(err)
				}
			}

		}()
	})

	if err = dg.Open(); err != nil {
		fmt.Println("error opening connection: ", err)
		return
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
