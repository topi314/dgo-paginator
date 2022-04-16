package paginator

import "github.com/bwmarrin/discordgo"

func NewEventCollector[E any](session *discordgo.Session, filterFunc func(s *discordgo.Session, e E) bool) (<-chan E, func()) {
	ch := make(chan E)

	handler := func(s *discordgo.Session, e E) {
		if filterFunc != nil && !filterFunc(s, e) {
			return
		}
		ch <- e
	}
	removeFunc := session.AddHandler(handler)

	return ch, func() {
		removeFunc()
		close(ch)
	}
}
