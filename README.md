[![Go Reference](https://pkg.go.dev/badge/github.com/TopiSenpai/dgo-paginator.svg)](https://pkg.go.dev/github.com/TopiSenpai/dgo-paginator)
[![Go Report](https://goreportcard.com/badge/github.com/TopiSenpai/dgo-paginator)](https://goreportcard.com/report/github.com/TopiSenpai/dgo-paginator)
[![Go Version](https://img.shields.io/github/go-mod/go-version/TopiSenpai/dgo-paginator)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/TopiSenpai/dgo-paginator/blob/master/LICENSE)
[![Version](https://img.shields.io/github/v/tag/TopiSenpai/dgo-paginator?label=release)](https://github.com/TopiSenpai/dgo-paginator/releases/latest)

# dgo-paginator

dgo-paginator is a paginator working with buttons. It can be used with interactions and normal messages.

## Getting Started

### Installing

```sh
go get github.com/TopiSenpai/dgo-paginator
```

### Paginator Usage

```go
// create new dgo session
dg, err := discordgo.New("Bot " + token)

// create a new pagination manager
manager := paginator.NewManager()

// register the pagination handler
dg.AddHandler(manager.OnInteractionCreate)

// add a message create handler to spawn a paginator
dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
  if m.Author.ID == s.State.User.ID || m.Content != "!test" {
    return
  }

  // your pages can be anything
  pages := []string{
    "page1",
    "page2",
    "page3",
  }
  
  err := manager.CreateMessage(s, m.ChannelID, &paginator.Paginator{
    // the PageFunc is called when a new page is requested put your data per page in here
    PageFunc: func(page int, embed *discordgo.MessageEmbed) {
      embed.Description = pages[page]
    },
    // the max pages this paginator has
    MaxPages:        len(pages),
    // expire after last usage or when created?
    ExpiryLastUsage: true,
  })
  if err != nil {
    fmt.Println(err)
  }
})

// open the session
if err = dg.Open(); err != nil {
  fmt.Println("error opening connection: ", err)
  return
}

// keep the session open
s := make(chan os.Signal, 1)
signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
<-s
```

### EventCollector Usage

Create a new collector by passing it the `discordgo.Session` & a filter function.
The filter function will be called for every event.
If the filter function returns true, the event will be collected and passed through the channel.
If the filter function returns false, the event will be ignored.
If you are done collecting don't forget to close the collector.

```go
eventChannel, stopCollector := event_collector.NewEventCollector(ssession, func(s *discordgo.Session, e *discordgo.MessageCreate) bool {
    return // filter your events here
})
```

## Examples

You can find examples under [_examples](https://github.com/TopiSenpai/dgo-paginator/blob/master/_examples)

## Contributing

Contributions are welcomed but for bigger changes please create a discussion.

## License

Distributed under the [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/TopiSenpai/dgo-paginator/blob/master/LICENSE)
. See LICENSE for more information.

