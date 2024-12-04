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

type Option interface {
	initializeChat(*chatOptions) error
}

type chatOptions struct {
	context                 context.Context
	publisherTopic          string
	publisher               message.Publisher
	subscriber              message.Subscriber
	historyDepth            int
	historyRetention        time.Duration
	historyCleanupFrequency time.Duration
}

type DefaultOptions struct{}

func (o DefaultOptions) initializeChat(c *chatOptions) error {
	if c.context == nil {
		c.context = context.Background()
	}
	if c.historyDepth == 0 {
		c.historyDepth = 100
	}
	if c.historyRetention < time.Second {
		c.historyRetention = time.Minute * 60
	}
	if c.historyCleanupFrequency == 0 {
		c.historyCleanupFrequency = time.Minute
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
