# Watermill Chat Technical Demo

[![https://pkg.go.dev/github.com/dkotik/watermillchat](https://pkg.go.dev/badge/github.com/dkotik/watermillchat.svg)](https://pkg.go.dev/github.com/dkotik/watermillchat)
[![https://github.com/dkotik/watermillchat/actions?query=workflow:test](https://github.com/dkotik/watermillchat/workflows/test/badge.svg?branch=main&event=push)](https://github.com/dkotik/watermillchat/actions?query=workflow:test)
[![https://coveralls.io/github/dkotik/watermillchat](https://coveralls.io/repos/github/dkotik/watermillchat/badge.svg?branch=main)](https://coveralls.io/github/dkotik/watermillchat)
[![https://goreportcard.com/report/github.com/dkotik/watermillchat](https://goreportcard.com/badge/github.com/dkotik/watermillchat)](https://goreportcard.com/report/github.com/dkotik/watermillchat)

This package is a portfolio proof of concept of a durable HTTP live chat atop [Watermill](https://watermill.io/) for the back-end event sourcing and [Data Star](https://data-star.dev/) for the immediate front-end rendering responsive to [Server Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events).

## Development Roadmap

If this package reaches **1.0** release, it will present a flexible and durable live chat system that can be easily integrated into projects that require chat modules with minimal dependencies and wide storage system support.

- [x] Add message history recall on boot.
- [x] Add random room generator.
- [x] Make room link sharable.
- [x] Add Olama integration plugin.
- [ ] Add user 0Auth authentication.

## Installation

```sh
# Install latest version for your project:
go get -u github.com/dkotik/watermillchat@latest

# Run live demonstration on local host:
go run github.com/dkotik/watermillchat/cmd/wmcserver@latest
```
