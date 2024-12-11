/*
Package watermillchat provides live hypermedia chat
for <watermill.io> event caster. Messages are delivered
as server side events.
*/
package watermillchat

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

const (
	DefaultWatermillTopic             = "watermillchat"
	DefaultHistoryMostMessagesPerRoom = 1000
	DefaultHistoryDepth               = DefaultHistoryMostMessagesPerRoom // TODO: deprecate
	DefaultHistoryRetention           = time.Minute * 60 * 24             // 24 hours
	DefaultHistoryCleanupFrequency    = time.Minute * 15
	DefaultCleanupFrequency           = DefaultHistoryCleanupFrequency // TODO: deprecate
)

type WatermillConfiguration struct {
	// Topic where the messages are published to and read from.
	// Defaults to [DefaultWatermillTopic].
	Topic      string
	Publisher  message.Publisher
	Subscriber message.Subscriber
}

type HistoryConfiguration struct {
	// Defaults to [VoidHistoryRepository].
	Repository HistoryRepository

	// Retention constraints the life time of messages before their deletion.
	// Defaults to [DefaultHistoryRetention].
	Retention time.Duration

	// CleanUpFrequency is the pause between message purge
	// cycles based on retention. Defaults to [DefaultHistoryCleanupFrequency].
	CleanUpFrequency time.Duration

	// MostMessagesPerRoom constraints the maximum number of
	// returned messages per room when history is loaded using. More messages may still
	// be present in the database, if they retention duration
	// has not yet run out. Defaults to [DefaultHistoryMostMessagesPerRoom].
	MostMessagesPerRoom int
}

type Configuration struct {
	Watermill WatermillConfiguration
	History   HistoryConfiguration
	Logger    *slog.Logger
}

func (c Configuration) Validate() (err error) {
	if strings.TrimSpace(c.Watermill.Topic) == "" {
		err = errors.Join(err, errors.New("missing Watermill topic"))
	}
	if c.Watermill.Publisher == nil {
		err = errors.Join(err, errors.New("missing Watermill publisher"))
	}
	if c.Watermill.Subscriber == nil {
		err = errors.Join(err, errors.New("missing Watermill subscriber"))
	}
	if c.History.Repository == nil {
		err = errors.Join(err, errors.New("missing history repository"))
	}
	if c.History.Retention < time.Minute {
		err = errors.Join(err, errors.New("history retention is lower than one minute"))
	}
	if c.History.CleanUpFrequency < time.Minute {
		err = errors.Join(err, errors.New("history clean up frequency is less than one minute"))
	}
	if c.History.MostMessagesPerRoom < 1 {
		err = errors.Join(err, errors.New("retained messages per room is lower than one"))
	}
	if c.Logger == nil {
		err = errors.Join(err, errors.New("missing logger"))
	}
	return err
}

type Chat struct {
	publisherTopic   string
	publisher        message.Publisher
	history          HistoryRepository
	historyDepth     int
	historyRetention time.Duration
	logger           *slog.Logger

	rooms map[string]*Room
	mu    *sync.Mutex
}

func New(ctx context.Context, c Configuration) (chat *Chat, err error) {
	if ctx == nil {
		return nil, errors.New("chat running context is missing")
	}
	if c.Logger == nil {
		c.Logger = slog.Default()
	}

	if c.Watermill.Topic == "" {
		c.Watermill.Topic = DefaultWatermillTopic
	}
	if c.Watermill.Publisher == nil || c.Watermill.Subscriber == nil {
		if c.Watermill.Publisher != nil {
			return nil, errors.New("watermill publisher is provided without a matching subscriber")
		}
		if c.Watermill.Subscriber != nil {
			return nil, errors.New("watermill subscriber is provided without a matching publisher")
		}
		pubSub := gochannel.NewGoChannel(
			gochannel.Config{},
			watermill.NewSlogLogger(c.Logger),
		)
		c.Watermill.Publisher = pubSub
		c.Watermill.Subscriber = pubSub
		go func(ctx context.Context, logger *slog.Logger) {
			<-ctx.Done()
			if err := pubSub.Close(); err != nil {
				logger.Error("unable to close down default watermill publisher", slog.Any("error", err))
			}
		}(ctx, c.Logger)
	}

	if c.History.Repository == nil {
		c.History.Repository = VoidHistoryRepository{}
	}
	if c.History.MostMessagesPerRoom == 0 {
		c.History.MostMessagesPerRoom = DefaultHistoryMostMessagesPerRoom
	}
	if c.History.Retention == 0 {
		c.History.Retention = DefaultHistoryRetention
	}
	if c.History.CleanUpFrequency == 0 {
		c.History.CleanUpFrequency = DefaultHistoryCleanupFrequency
	}

	if err = c.Validate(); err != nil {
		return nil, fmt.Errorf("unable to initialize Watermill chat: %w", err)
	}

	incomingBroadcasts, err := c.Watermill.Subscriber.Subscribe(ctx, c.Watermill.Topic)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to a Watermill subscriber: %w", err)
	}
	incomingHistoryBroadcasts, err := c.Watermill.Subscriber.Subscribe(ctx, c.Watermill.Topic)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to a Watermill subscriber: %w", err)
	}

	chat = &Chat{
		publisherTopic:   c.Watermill.Topic,
		publisher:        c.Watermill.Publisher,
		history:          c.History.Repository,
		historyDepth:     c.History.MostMessagesPerRoom,
		historyRetention: c.History.Retention,
		logger:           c.Logger,

		rooms: make(map[string]*Room),
		mu:    &sync.Mutex{},
	}
	go c.History.Repository.Listen(incomingHistoryBroadcasts)
	go chat.Listen(incomingBroadcasts)
	go chat.cleanup(ctx, c.History.CleanUpFrequency)
	return chat, nil
}
