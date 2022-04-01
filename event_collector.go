package paginator

import "github.com/bwmarrin/discordgo"

func NewEventCollector[E any](s *discordgo.Session, filterFunc func(s *discordgo.Session, e E) bool) (<-chan E, func()) {
	ch := make(chan E)

	coll := &collector[E]{
		FilterFunc: filterFunc,
		Chan:       ch,
	}
	remove := s.AddHandler(coll.handler)

	return ch, func() {
		remove()
		close(ch)
	}
}

type collector[E any] struct {
	FilterFunc func(s *discordgo.Session, e E) bool
	Chan       chan<- E
}

func (c *collector[E]) handler(s *discordgo.Session, e any) {
	if event, ok := e.(E); ok {
		if !c.FilterFunc(s, event) {
			return
		}
		c.Chan <- event
	}
}
