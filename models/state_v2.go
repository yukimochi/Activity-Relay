package models

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"strings"
)

type AddOrDelOperation int
type SetOrUnSetOperation int
type RelayOptionType int

const (
	AddOperation AddOrDelOperation = iota + 1
	DelOperation

	SetOperation SetOrUnSetOperation = iota + 1
	UnSetOperation

	PersonOnlyV2 RelayOptionType = iota + 1
	ManuallyAcceptV2
)

type RelayOption struct {
	PersonOnly     bool `json:"blockService,omitempty"`
	ManuallyAccept bool `json:"manuallyAccept,omitempty"`
}

// Subscriber : Mastodon Traditional Style Relay Subscriber
type Subscriber struct {
	Domain     string `json:"domain,omitempty"`
	InboxURL   string `json:"inbox_url,omitempty"`
	ActivityID string `json:"activity_id,omitempty"`
	ActorID    string `json:"actor_id,omitempty"`
}

// Follower : LitePub Style Relay Follower
type Follower struct {
	Domain         string `json:"domain,omitempty"`
	InboxURL       string `json:"inbox_url,omitempty"`
	ActivityID     string `json:"activity_id,omitempty"`
	ActorID        string `json:"actor_id,omitempty"`
	MutuallyFollow bool   `json:"mutually_follow,omitempty"`
}

type RelayStateV2 struct {
	RelayOption             RelayOption  `json:"relayConfig,omitempty"`
	LimitedDomains          []string     `json:"limitedDomains,omitempty"`
	BlockedDomains          []string     `json:"blockedDomains,omitempty"`
	Subscribers             []Subscriber `json:"subscriptions,omitempty"`
	Followers               []Follower   `json:"followers,omitempty"`
	SubscribersAndFollowers []Subscriber `json:"-"`
}

type RelayStateV2BuilderOptions struct {
	Refreshable   bool
	RefreshResult chan<- bool
}

// NewStateV2 : Create new RelayState instance
func NewStateV2(ctx context.Context, config *RelayConfigV2, options RelayStateV2BuilderOptions) *RelayStateV2 {
	redisClient, err := config.NewRedisClient(ctx)
	if err != nil {
		panic(err)
	}

	state := RelayStateV2{}
	err = state.load(ctx, redisClient)
	if err != nil {
		panic(err)
	}

	if options.Refreshable {
		messageChannel := redisClient.Subscribe(ctx, "relay_refresh_v2").Channel()
		go func(ctx context.Context, result chan<- bool) {
			for range messageChannel {
				err := state.load(ctx, redisClient)
				if err != nil {
					logrus.Error("<- State refresh failed: " + err.Error())
				} else {
					logrus.Info("<- State refreshed successfully")
				}

				if result != nil {
					result <- true
				}
			}
		}(context.Background(), options.RefreshResult)
	}

	return &state
}

func (state *RelayStateV2) load(ctx context.Context, redisClient *redis.Client) error {
	var subscribers []Subscriber
	var followers []Follower
	var subscribersAndFollowers []Subscriber

	// load RelayOption
	personOnlyInt, err := redisClient.HGet(ctx, "relay:config", "block_service").Result()
	if err != nil {
		return err
	}
	personOnly := personOnlyInt == "1"
	manuallyAcceptInt, err := redisClient.HGet(ctx, "relay:config", "manually_accept").Result()
	if err != nil {
		return err
	}
	manuallyAccept := manuallyAcceptInt == "1"

	// load LimitedDomains
	limitedDomains, err := redisClient.HKeys(ctx, "relay:config:limitedDomain").Result()
	if err != nil {
		return err
	}

	// load BlockedDomains
	blockedDomains, err := redisClient.HKeys(ctx, "relay:config:blockedDomain").Result()
	if err != nil {
		return err
	}

	// load Subscribers
	subscriberIterator, err := redisClient.Keys(ctx, "relay:subscription:*").Result()
	if err != nil {
		return err
	}
	for _, key := range subscriberIterator {
		details, err := redisClient.HGetAll(ctx, key).Result()
		if err != nil {
			return err
		}

		inboxURL, ok := details["inbox_url"]
		if !ok {
			return errors.New("inbox_url is required: " + key)
		}
		activityId, ok := details["activity_id"]
		if !ok {
			activityId = ""
		}
		actorId, ok := details["actor_id"]
		if !ok {
			actorId = ""
		}

		subscriber := Subscriber{
			Domain:     strings.Replace(key, "relay:subscription:", "", 1),
			InboxURL:   inboxURL,
			ActivityID: activityId,
			ActorID:    actorId,
		}
		subscribers = append(subscribers, subscriber)
	}

	// load Followers
	followerIterator, err := redisClient.Keys(ctx, "relay:follower:*").Result()
	if err != nil {
		return err
	}
	for _, key := range followerIterator {
		details, err := redisClient.HGetAll(ctx, key).Result()
		if err != nil {
			return err
		}

		inboxURL, ok := details["inbox_url"]
		if !ok {
			return errors.New("inbox_url is required: " + key)
		}
		activityId, ok := details["activity_id"]
		if !ok {
			activityId = ""
		}
		actorId, ok := details["actor_id"]
		if !ok {
			actorId = ""
		}
		mutuallyFollow, ok := details["mutually_follow"]
		if !ok {
			mutuallyFollow = "0"
		}

		follower := Follower{
			Domain:         strings.Replace(key, "relay:follower:", "", 1),
			InboxURL:       inboxURL,
			ActivityID:     activityId,
			ActorID:        actorId,
			MutuallyFollow: mutuallyFollow == "1",
		}
		followers = append(followers, follower)
	}

	// Build SubscribersAndFollowers
	subscribersAndFollowers = append(subscribersAndFollowers, subscribers...)
	for _, follower := range followers {
		subscribersAndFollowers = append(subscribersAndFollowers, Subscriber{
			Domain:     follower.Domain,
			InboxURL:   follower.InboxURL,
			ActivityID: follower.ActivityID,
			ActorID:    follower.ActorID,
		})
	}

	// Update state
	state.RelayOption.PersonOnly = personOnly
	state.RelayOption.ManuallyAccept = manuallyAccept
	state.LimitedDomains = limitedDomains
	state.BlockedDomains = blockedDomains
	state.Subscribers = subscribers
	state.Followers = followers
	state.SubscribersAndFollowers = subscribersAndFollowers

	return nil
}

// SetOption : Set relay option
func (state *RelayStateV2) SetOption(ctx context.Context, redisClient *redis.Client, key RelayOptionType, operation SetOrUnSetOperation) error {
	var operationStr string

	switch operation {
	case SetOperation:
		operationStr = "1"
	case UnSetOperation:
		operationStr = "0"
	default:
		return errors.New("unknown operation")
	}

	switch key {
	case PersonOnlyV2:
		_, err := redisClient.HSet(ctx, "relay:config", "block_service", operationStr).Result()
		if err != nil {
			logrus.Error("Failed to set or unset PersonOnly")
			return err
		}
		return nil
	case ManuallyAcceptV2:
		_, err := redisClient.HSet(ctx, "relay:config", "manually_accept", operationStr).Result()
		if err != nil {
			logrus.Error("Failed to set or unset ManuallyAccept")
			return err
		}
		return nil
	default:
		return errors.New("unknown option")
	}
}

func (state *RelayStateV2) AddOrDelLimitedDomains(ctx context.Context, redisClient *redis.Client, domains []string, operation AddOrDelOperation) error {
	for _, domain := range domains {
		var changed int64
		var err error

		switch operation {
		case AddOperation:
			changed, err = redisClient.HSet(ctx, "relay:config:limitedDomain", domain, "1").Result()
			break
		case DelOperation:
			changed, err = redisClient.HDel(ctx, "relay:config:limitedDomain", domain).Result()
			break
		default:
			return errors.New("unknown operation")
		}
		if err != nil {
			logrus.Error("Failed to add or delete limited domain: " + domain)
			return err
		}
		if changed == 0 {
			logrus.Warn("No change for domain: " + domain)
		}
	}

	return nil
}

func (state *RelayStateV2) AddOrDelBlockedDomains(ctx context.Context, redisClient *redis.Client, domains []string, operation AddOrDelOperation) error {
	for _, domain := range domains {
		var changed int64
		var err error

		switch operation {
		case AddOperation:
			changed, err = redisClient.HSet(ctx, "relay:config:blockedDomain", domain, "1").Result()
			break
		case DelOperation:
			changed, err = redisClient.HDel(ctx, "relay:config:blockedDomain", domain).Result()
			break
		default:
			return errors.New("unknown operation")
		}
		if err != nil {
			logrus.Error("Failed to add or delete blocked domain: " + domain)
			return err
		}
		if changed == 0 {
			logrus.Warn("No change for domain: " + domain)
		}
	}

	return nil
}

// SelectSubscriber : Select instance from subscriber list
func (state *RelayStateV2) SelectSubscriber(domain string) *Subscriber {
	for _, subscriber := range state.Subscribers {
		if domain == subscriber.Domain {
			return &subscriber
		}
	}
	return nil
}

// AddSubscriber : Add instance to subscriber list
func (state *RelayStateV2) AddSubscriber(ctx context.Context, redisClient *redis.Client, subscriber Subscriber) error {
	_, err := redisClient.HMSet(ctx, "relay:subscription:"+subscriber.Domain, map[string]interface{}{
		"inbox_url":   subscriber.InboxURL,
		"activity_id": subscriber.ActivityID,
		"actor_id":    subscriber.ActorID,
	}).Result()
	if err != nil {
		logrus.Error("Failed to add subscriber: " + subscriber.Domain)
		return err
	}
	return nil
}

// DelSubscriber : Delete instance from subscriber list
func (state *RelayStateV2) DelSubscriber(ctx context.Context, redisClient *redis.Client, domain string) error {
	changed, err := redisClient.Del(ctx, "relay:subscription:"+domain).Result()
	if err != nil {
		logrus.Error("Failed to delete subscriber: " + domain)
		return err
	}
	if changed == 0 {
		logrus.Warn("No change for domain: " + domain)
	}
	return nil
}

// SelectFollower : Select instance from follower list
func (state *RelayStateV2) SelectFollower(domain string) *Follower {
	for _, follower := range state.Followers {
		if domain == follower.Domain {
			return &follower
		}
	}
	return nil
}

// AddFollower : Add instance to follower list
func (state *RelayStateV2) AddFollower(ctx context.Context, redisClient *redis.Client, follower Follower) error {
	_, err := redisClient.HMSet(ctx, "relay:follower:"+follower.Domain, map[string]interface{}{
		"inbox_url":       follower.InboxURL,
		"activity_id":     follower.ActivityID,
		"actor_id":        follower.ActorID,
		"mutually_follow": follower.MutuallyFollow,
	}).Result()
	if err != nil {
		logrus.Error("Failed to add follower: " + follower.Domain)
		return err
	}
	return nil
}

// DelFollower : Delete instance from follower list
func (state *RelayStateV2) DelFollower(ctx context.Context, redisClient *redis.Client, domain string) error {
	changed, err := redisClient.Del(ctx, "relay:follower:"+domain).Result()
	if err != nil {
		logrus.Error("Failed to delete follower: " + domain)
		return err
	}
	if changed == 0 {
		logrus.Warn("No change for domain: " + domain)
	}
	return nil
}

// UpdateFollowerStatus : Update MutuallyFollow Status
func (state *RelayStateV2) UpdateFollowerStatus(ctx context.Context, redisClient *redis.Client, domain string, mutuallyFollow bool) error {
	if mutuallyFollow {
		_, err := redisClient.HSet(ctx, "relay:follower:"+domain, "mutually_follow", "1").Result()
		if err != nil {
			logrus.Error("Failed to update mutually follow status: " + domain)
			return err
		}
	} else {
		_, err := redisClient.HSet(ctx, "relay:follower:"+domain, "mutually_follow", "0").Result()
		if err != nil {
			logrus.Error("Failed to update mutually follow status: " + domain)
			return err
		}
	}
	return nil
}

// PublishModify : Publish refresh message
func (state *RelayStateV2) PublishModify(ctx context.Context, redisClient *redis.Client) error {
	_, err := redisClient.Publish(ctx, "relay_refresh_v2", "refresh").Result()
	if err != nil {
		return err
	}

	return nil
}
