package paginator

import (
	"github.com/bwmarrin/discordgo"
)

var DefaultConfig = Config{
	ButtonsConfig: ButtonsConfig{
		First: &ComponentOptions{
			Emoji: &discordgo.ComponentEmoji{
				Name: "⏮",
			},
			Style: discordgo.PrimaryButton,
		},
		Back: &ComponentOptions{
			Emoji: &discordgo.ComponentEmoji{
				Name: "◀",
			},
			Style: discordgo.PrimaryButton,
		},
		Stop: &ComponentOptions{
			Emoji: &discordgo.ComponentEmoji{
				Name: "🗑",
			},
			Style: discordgo.DangerButton,
		},
		Next: &ComponentOptions{
			Emoji: &discordgo.ComponentEmoji{
				Name: "▶",
			},
			Style: discordgo.PrimaryButton,
		},
		Last: &ComponentOptions{
			Emoji: &discordgo.ComponentEmoji{
				Name: "⏩",
			},
			Style: discordgo.PrimaryButton,
		},
	},
	NotYourPaginatorMessage: "You can't interact with this paginator because it's not yours.",
	CustomIDPrefix:          "paginator",
	EmbedColor:              0x4c50c1,
}

type Config struct {
	ButtonsConfig           ButtonsConfig
	NotYourPaginatorMessage string
	CustomIDPrefix          string
	EmbedColor              int
}

type ButtonsConfig struct {
	First *ComponentOptions
	Back  *ComponentOptions
	Stop  *ComponentOptions
	Next  *ComponentOptions
	Last  *ComponentOptions
}

type ComponentOptions struct {
	Emoji *discordgo.ComponentEmoji
	Label string
	Style discordgo.ButtonStyle
}

type ConfigOpt func(config *Config)

// Apply applies the given RequestOpt(s) to the RequestConfig & sets the context if none is set
func (c *Config) Apply(opts []ConfigOpt) {
	for _, opt := range opts {
		opt(c)
	}
}

func WithButtonsConfig(buttonsConfig ButtonsConfig) ConfigOpt {
	return func(config *Config) {
		config.ButtonsConfig = buttonsConfig
	}
}

func WithNotYourPaginatorMessage(message string) ConfigOpt {
	return func(config *Config) {
		config.NotYourPaginatorMessage = message
	}
}

func WithCustomIDPrefix(prefix string) ConfigOpt {
	return func(config *Config) {
		config.CustomIDPrefix = prefix
	}
}

func WithEmbedColor(color int) ConfigOpt {
	return func(config *Config) {
		config.EmbedColor = color
	}
}
