package watermillchat

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

const (
	DefaultHistoryDepth     = 1000
	DefaultHistoryRetention = time.Minute * 60 * 24 // 24 hours
	DefaultCleanupFrequency = time.Minute * 15
)

type Option interface {
	initializeChat(*chatOptions) error
}

type chatOptions struct {
	context                 context.Context
	publisherTopic          string
	publisher               message.Publisher
	subscriber              message.Subscriber
	history                 HistoryRepository
	historyDepth            int
	historyRetention        time.Duration
	historyCleanupFrequency time.Duration
}

type DefaultOptions struct{}

func (o DefaultOptions) initializeChat(c *chatOptions) error {
	if c.context == nil {
		c.context = context.Background()
	}
	if c.history == nil {
		c.history = VoidHistoryRepository{}
	}
	if c.historyDepth == 0 {
		c.historyDepth = DefaultHistoryDepth
	}
	if c.historyRetention < time.Second {
		c.historyRetention = DefaultHistoryRetention
	}
	if c.historyCleanupFrequency == 0 {
		c.historyCleanupFrequency = DefaultCleanupFrequency
	}
	if c.publisherTopic == "" {
		c.publisherTopic = "watermillchat"
	}
	if c.publisher == nil || c.subscriber == nil {
		pubSub := gochannel.NewGoChannel(
			gochannel.Config{},
			watermill.NewSlogLogger(nil),
		)
		c.publisher = pubSub
		c.subscriber = pubSub
		go func() {
			<-c.context.Done()
			if err := pubSub.Close(); err != nil {
				slog.Error("unable to close down default watermill publisher", slog.Any("error", err))
			}
		}()
	}
	return nil
}

type contextOption struct {
	ctx context.Context
}

func (o contextOption) initializeChat(c *chatOptions) error {
	if o.ctx == nil {
		return errors.New("cannot use a <nil> context")
	}
	c.context = o.ctx
	return nil
}

func WithContext(ctx context.Context) Option {
	return contextOption{ctx: ctx}
}

type historyRepositoryOption struct {
	repository HistoryRepository
}

func (o historyRepositoryOption) initializeChat(c *chatOptions) error {
	if o.repository == nil {
		return errors.New("cannot use a <nil> history repository")
	}
	if c.history != nil {
		return errors.New("history repository is already set")
	}
	c.history = o.repository
	return nil
}

// WithHistoryRepository stores all messages into a repository.
// If present, messages are recovered before a room is loaded
// into memory. Defaults to [VoidHistoryRepository].
func WithHistoryRepository(h HistoryRepository) Option {
	return historyRepositoryOption{
		repository: h,
	}
}

type historyDepthOption int

func (o historyDepthOption) initializeChat(c *chatOptions) error {
	if o < 1 {
		return fmt.Errorf("invalid history depth: %d", o)
	}
	if c.historyDepth != 0 {
		return errors.New("history depth is already set")
	}
	c.historyDepth = int(o)
	return nil
}

// WithHistoryDepth constraints the maximum number of
// retained messages per room. Defaults to [DefaultHistoryDepth].
func WithHistoryDepth(value int) Option {
	return historyDepthOption(value)
}

type historyRetentionOption time.Duration

func (o historyRetentionOption) initializeChat(c *chatOptions) error {
	if time.Duration(o) < time.Second {
		return errors.New("history retention must be longer than one second")
	}
	if c.historyRetention != 0 {
		return errors.New("history retention is already set")
	}
	c.historyRetention = time.Duration(o)
	return nil
}

// WithHistoryRetention constraints the life time of messages before deletion. Defaults to [DefaultHistoryRetention].
func WithHistoryRetention(d time.Duration) Option {
	return historyRetentionOption(d)
}

type historyCleanupFrequencyOption time.Duration

func (o historyCleanupFrequencyOption) initializeChat(c *chatOptions) error {
	if time.Duration(o) < time.Second {
		return errors.New("history clean up frequency must be longer than one second")
	}
	if c.historyCleanupFrequency != 0 {
		return errors.New("history clean up frequency is already set")
	}
	c.historyCleanupFrequency = time.Duration(o)
	return nil
}

// WithHistoryCleanupFrequency is the pause between message purge. Defaults to [DefaultHistoryCleanupFrequency].
func WithHistoryCleanupFrequency(d time.Duration) Option {
	return historyCleanupFrequencyOption(d)
}

type watermillTopic string

func (o watermillTopic) initializeChat(c *chatOptions) error {
	if len(o) == 0 {
		return errors.New("cannot publish to an empty Watermill topic")
	}
	if c.publisherTopic != "" {
		return errors.New("Watermill publisher topic is already set")
	}
	c.publisherTopic = string(o)
	return nil
}

func WithWatermillTopic(topic string) Option {
	return watermillTopic(topic)
}

type watermillPubSub struct {
	publisher  message.Publisher
	subscriber message.Subscriber
}

func (o watermillPubSub) initializeChat(c *chatOptions) error {
	if o.publisher == nil {
		return errors.New("cannot use a <nil> Watermill publisher")
	}
	if c.publisherTopic == "" {
		return errors.New("provide a Watermill publisher topic first")
	}
	if o.subscriber == nil {
		return errors.New("cannot use a <nil> Watermill subscriber")
	}
	if o.publisher != nil || o.subscriber != nil {
		return errors.New("Watermill publisher or subscriber is already set")
	}
	c.publisher = o.publisher
	c.subscriber = o.subscriber
	return nil
}

func WithWatermillPubSub(publisher message.Publisher,
	subscriber message.Subscriber) Option {
	return watermillPubSub{
		publisher:  publisher,
		subscriber: subscriber,
	}
}
