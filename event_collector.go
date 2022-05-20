package paginator

import (
	"context"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// WaitForEvent waits for an event passing the filterFunc and then calls the actionFunc. You can cancel this function with the passed context.Context and the cancelFunc gets called then.
func WaitForEvent[E any](session *discordgo.Session, ctx context.Context, filterFunc func(s *discordgo.Session, e E) bool, actionFunc func(s *discordgo.Session, e E), cancelFunc func()) {
	ch, cancel := NewEventCollector(session, filterFunc)

	select {
	case <-ctx.Done():
		cancel()
		if cancelFunc != nil {
			cancelFunc()
		}
	case e := <-ch:
		cancel()
		if actionFunc != nil {
			actionFunc(session, e)
		}
	}
}

// NewEventCollector returns a channel in which the events of type T gets sent which pass the passed filter and a function which can be used to stop the event collector.
// The close function needs to be called to stop the event collector.
func NewEventCollector[E any](session *discordgo.Session, filterFunc func(s *discordgo.Session, e E) bool) (<-chan E, func()) {
	ch := make(chan E)
	var once sync.Once

	handler := func(s *discordgo.Session, e E) {
		if filterFunc != nil && !filterFunc(s, e) {
			return
		}
		ch <- e
	}
	removeFunc := session.AddHandler(handler)

	return ch, func() {
		once.Do(func() {
			removeFunc()
			close(ch)
		})
	}
}
