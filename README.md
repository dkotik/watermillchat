# Watermill Chat Technical Demo

[![https://pkg.go.dev/github.com/dkotik/watermillchat](https://pkg.go.dev/badge/github.com/dkotik/watermillchat.svg)](https://pkg.go.dev/github.com/dkotik/watermillchat)
[![https://github.com/dkotik/watermillchat/actions?query=workflow:test](https://github.com/dkotik/watermillchat/workflows/test/badge.svg?branch=main&event=push)](https://github.com/dkotik/watermillchat/actions?query=workflow:test)
[![https://coveralls.io/github/dkotik/watermillchat](https://coveralls.io/repos/github/dkotik/watermillchat/badge.svg?branch=main)](https://coveralls.io/github/dkotik/watermillchat)
[![https://goreportcard.com/report/github.com/dkotik/watermillchat](https://goreportcard.com/badge/github.com/dkotik/watermillchat)](https://goreportcard.com/report/github.com/dkotik/watermillchat)

This package is a portfolio proof of concept of a durable HTTP live chat atop [Watermill](https://watermill.io/) for the back-end event sourcing and [Data Star](https://data-star.dev/) for the immediate front-end rendering responsive to [Server Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events). If this package reaches **1.0** release, it will provide a flexible and durable live chat system that can be integrated into projects that require chat functionality with minimal dependencies and wide storage system support.

## Development Roadmap

- [x] Add message history recall on boot.
- [x] Add random room generator.
- [x] Make room link sharable.
- [x] Add Olama integration plugin.
- [ ] Upgrade [Data Star](https://data-star.dev/) to version **1.0**, once it is released.
- [ ] Add 0Auth authentication.
- [ ] Add side-mounted integration into existing HTML pages.

## Installation

```sh
# Install latest version for your project:
go get -u github.com/dkotik/watermillchat@latest

# Run live demonstration on local host:
go run github.com/dkotik/watermillchat/cmd/wmcserver@latest
```

## Impressions

- [Data Star](https://data-star.dev/) was much easier to pick up and run with than HTMX.
  It was a breath of fresh air compared to Svelte, Vue, and React.
- Durable event sourcing solves most of the synchronization problems.
  The only concievable scenario where a message could be lost is if two clients join
  an expired room simultaneously and one sends a message immediately upon joining.
  The message could be lost if one of them received the memory snapshot, while the
  other loaded messages from history provider, before the memory snapshot was updated.
- Integrating Ollama generative Ai conversationalist was easy.
