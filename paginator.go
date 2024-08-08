package paginator

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

func NewManager(opts ...ConfigOpt) *Manager {
	config := &DefaultConfig
	config.Apply(opts)
	manager := &Manager{
		Config:     *config,
		paginators: map[string]*Paginator{},
	}
	manager.startCleanup()
	return manager
}

type Manager struct {
	Config Config

	mu         sync.Mutex
	paginators map[string]*Paginator
}

func (m *Manager) startCleanup() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			m.cleanup()
		}
	}()
}

func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for _, p := range m.paginators {
		if !p.Expiry.IsZero() && p.Expiry.After(now) {
			// TODO: remove components?
			delete(m.paginators, p.ID)
		}
	}
}

type Paginator struct {
	PageFunc        func(page int, embed *discordgo.MessageEmbed)
	MaxPages        int
	CurrentPage     int
	Creator         string
	Expiry          time.Time
	ExpiryLastUsage bool
	ID              string
}

func (m *Manager) CreateMessage(s *discordgo.Session, channelID string, paginator *Paginator) error {
	if paginator.ID == "" {
		paginator.ID = fmt.Sprintf("%s-%d", channelID, time.Now().UnixNano())
	}

	m.add(paginator)

	_, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{m.makeEmbed(paginator)},
		Components: []discordgo.MessageComponent{m.createComponents(paginator)},
	})
	return err
}

func (m *Manager) CreateInteraction(s *discordgo.Session, interaction *discordgo.Interaction, paginator *Paginator, acknowledged bool) error {
	if paginator.ID == "" {
		paginator.ID = interaction.ID
	}

	m.add(paginator)

	var err error
	if acknowledged {
		_, err = s.InteractionResponseEdit(interaction, m.makeMessageUpdate(paginator))
	} else {
		err = s.InteractionRespond(interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: m.makeInteractionResponseData(paginator),
		})
	}
	return err
}

func (m *Manager) add(paginator *Paginator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paginators[paginator.ID] = paginator
}

func (m *Manager) remove(paginatorID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.paginators, paginatorID)
}

func (m *Manager) OnInteractionCreate(s *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type != discordgo.InteractionMessageComponent {
		return
	}

	customID := interaction.MessageComponentData().CustomID
	if !strings.HasPrefix(customID, m.Config.CustomIDPrefix) {
		return
	}
	ids := strings.Split(customID, ":")
	paginatorID, action := ids[1], ids[2]
	paginator, ok := m.paginators[paginatorID]
	if !ok {
		return
	}

	if paginator.Creator != "" && paginator.Creator != interaction.Member.User.ID {
		_ = s.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: m.Config.NotYourPaginatorMessage,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	switch action {
	case "first":
		paginator.CurrentPage = 0

	case "back":
		paginator.CurrentPage--

	case "stop":
		_ = s.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Components: []discordgo.MessageComponent{},
			},
		})
		m.remove(paginatorID)
		return

	case "next":
		paginator.CurrentPage++

	case "last":
		paginator.CurrentPage = paginator.MaxPages - 1
	}

	paginator.Expiry = time.Now()

	if err := s.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: m.makeInteractionResponseData(paginator),
	}); err != nil {
		fmt.Printf("error editing interaction: %s\n", err)
	}

}

func (m *Manager) makeEmbed(paginator *Paginator) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color: m.Config.EmbedColor,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Page: %d/%d", paginator.CurrentPage+1, paginator.MaxPages),
		},
	}
	paginator.PageFunc(paginator.CurrentPage, embed)
	return embed
}

func (m *Manager) makeInteractionResponseData(paginator *Paginator) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Embeds:     []*discordgo.MessageEmbed{m.makeEmbed(paginator)},
		Components: []discordgo.MessageComponent{m.createComponents(paginator)},
	}
}

func (m *Manager) makeMessageUpdate(paginator *Paginator) *discordgo.WebhookEdit {
	return &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{m.makeEmbed(paginator)},
		Components: &[]discordgo.MessageComponent{m.createComponents(paginator)},
	}
}

func (m *Manager) formatCustomID(paginator *Paginator, action string) string {
	return m.Config.CustomIDPrefix + ":" + paginator.ID + ":" + action
}

func (m *Manager) createComponents(paginator *Paginator) discordgo.MessageComponent {
	cfg := m.Config.ButtonsConfig
	actionRow := discordgo.ActionsRow{}

	if cfg.First != nil {
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.First.Label,
			Style:    cfg.First.Style,
			Disabled: paginator.CurrentPage == 0,
			Emoji:    cfg.First.Emoji,
			CustomID: m.formatCustomID(paginator, "first"),
		})
	}
	if cfg.Back != nil {
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.Back.Label,
			Style:    cfg.Back.Style,
			Disabled: paginator.CurrentPage == 0,
			Emoji:    cfg.Back.Emoji,
			CustomID: m.formatCustomID(paginator, "back"),
		})
	}

	if cfg.Stop != nil {
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.Stop.Label,
			Style:    cfg.Stop.Style,
			Emoji:    cfg.Stop.Emoji,
			CustomID: m.formatCustomID(paginator, "stop"),
		})
	}

	if cfg.Next != nil {
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.Next.Label,
			Style:    cfg.Next.Style,
			Disabled: paginator.CurrentPage == paginator.MaxPages-1,
			Emoji:    cfg.Next.Emoji,
			CustomID: m.formatCustomID(paginator, "next"),
		})
	}
	if cfg.Last != nil {
		actionRow.Components = append(actionRow.Components, discordgo.Button{
			Label:    cfg.Last.Label,
			Style:    cfg.Last.Style,
			Disabled: paginator.CurrentPage == paginator.MaxPages-1,
			Emoji:    cfg.Last.Emoji,
			CustomID: m.formatCustomID(paginator, "last"),
		})
	}

	return actionRow
}
