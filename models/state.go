package models

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Config : Enum for RelayConfig
type Config int

const (
	// PersonOnly : Limited for Person-Type Actor
	PersonOnly Config = iota
	// ManuallyAccept : Manually Accept Follow-Request
	ManuallyAccept
)

// RelayState : Store Subscribers, Followers And Relay Configurations
type RelayState struct {
	RedisClient *redis.Client `json:"-"`
	notifiable  bool
	mutex       sync.RWMutex

	RelayConfig             relayConfig  `json:"relayConfig,omitempty"`
	LimitedDomains          []string     `json:"limitedDomains,omitempty"`
	BlockedDomains          []string     `json:"blockedDomains,omitempty"`
	Subscribers             []Subscriber `json:"subscriptions,omitempty"`
	Followers               []Follower   `json:"followers,omitempty"`
	SubscribersAndFollowers []Subscriber `json:"-"`
}

// NewState : Create new RelayState instance with redis client
func NewState(redisClient *redis.Client, notifiable bool) *RelayState {
	config := new(RelayState)
	config.RedisClient = redisClient
	config.notifiable = notifiable

	config.Load()
	return config
}

func (config *RelayState) ListenNotify(c chan<- bool) {
	_, err := config.RedisClient.Subscribe(context.TODO(), "relay_refresh").Receive(context.TODO())
	if err != nil {
		panic(err)
	}
	ch := config.RedisClient.Subscribe(context.TODO(), "relay_refresh").Channel()

	cNotify := c != nil
	go func() {
		for range ch {
			logrus.Info("RelayState reloaded")
			config.Load()
			if cNotify {
				c <- true
			}
		}
	}()
}

// Load : Refresh content from redis.
// If any Redis command fails, the previous in-memory state is kept so
// that a transient error does not wipe the subscriber lists.
func (config *RelayState) Load() {
	newRelayConfig, err := loadRelayConfig(config.RedisClient)
	if err != nil {
		logrus.Error("Failed to reload RelayState, keep previous state : ", err)
		return
	}

	limitedDomains, err := config.RedisClient.HKeys(context.TODO(), "relay:config:limitedDomain").Result()
	if err != nil {
		logrus.Error("Failed to reload RelayState, keep previous state : ", err)
		return
	}
	blockedDomains, err := config.RedisClient.HKeys(context.TODO(), "relay:config:blockedDomain").Result()
	if err != nil {
		logrus.Error("Failed to reload RelayState, keep previous state : ", err)
		return
	}

	subscriptionKeys, err := RedisScanKeys(config.RedisClient, "relay:subscription:*")
	if err != nil {
		logrus.Error("Failed to reload RelayState, keep previous state : ", err)
		return
	}
	followerKeys, err := RedisScanKeys(config.RedisClient, "relay:follower:*")
	if err != nil {
		logrus.Error("Failed to reload RelayState, keep previous state : ", err)
		return
	}

	// Fetch all subscription/follower fields in a single pipeline so that
	// a reload does not issue O(N) sequential round-trips.
	pipe := config.RedisClient.Pipeline()
	subscriptionCmds := make([]*redis.SliceCmd, len(subscriptionKeys))
	for i, key := range subscriptionKeys {
		subscriptionCmds[i] = pipe.HMGet(context.TODO(), key, "inbox_url", "activity_id", "actor_id")
	}
	followerCmds := make([]*redis.SliceCmd, len(followerKeys))
	for i, key := range followerKeys {
		followerCmds[i] = pipe.HMGet(context.TODO(), key, "inbox_url", "activity_id", "actor_id", "mutually_follow")
	}
	if _, err := pipe.Exec(context.TODO()); err != nil {
		logrus.Error("Failed to reload RelayState, keep previous state : ", err)
		return
	}

	var subscribers []Subscriber
	var followers []Follower
	var subscribersAndFollowers []Subscriber
	for i, key := range subscriptionKeys {
		values, _ := subscriptionCmds[i].Result()
		subscriber := Subscriber{
			Domain:     strings.Replace(key, "relay:subscription:", "", 1),
			InboxURL:   sliceStringValue(values, 0),
			ActivityID: sliceStringValue(values, 1),
			ActorID:    sliceStringValue(values, 2),
		}
		subscribers = append(subscribers, subscriber)
		subscribersAndFollowers = append(subscribersAndFollowers, subscriber)
	}
	for i, key := range followerKeys {
		values, _ := followerCmds[i].Result()
		follower := Follower{
			Domain:         strings.Replace(key, "relay:follower:", "", 1),
			InboxURL:       sliceStringValue(values, 0),
			ActivityID:     sliceStringValue(values, 1),
			ActorID:       sliceStringValue(values, 2),
			MutuallyFollow: sliceStringValue(values, 3) == "1",
		}
		followers = append(followers, follower)
		subscribersAndFollowers = append(subscribersAndFollowers, Subscriber{
			Domain:     follower.Domain,
			InboxURL:   follower.InboxURL,
			ActivityID: follower.ActivityID,
			ActorID:    follower.ActorID,
		})
	}

	config.mutex.Lock()
	defer config.mutex.Unlock()

	config.RelayConfig = newRelayConfig
	config.LimitedDomains = limitedDomains
	config.BlockedDomains = blockedDomains
	config.Subscribers = subscribers
	config.Followers = followers
	config.SubscribersAndFollowers = subscribersAndFollowers
}

// MarshalJSON : Serialize relay state under the read lock so that a
// concurrent reload cannot race with an export.
func (config *RelayState) MarshalJSON() ([]byte, error) {
	config.mutex.RLock()
	defer config.mutex.RUnlock()

	type alias RelayState
	return json.Marshal((*alias)(config))
}

// SubscribersSnapshot : Return a copy of the subscriber list
func (config *RelayState) SubscribersSnapshot() []Subscriber {
	config.mutex.RLock()
	defer config.mutex.RUnlock()

	return append([]Subscriber(nil), config.Subscribers...)
}

// FollowersSnapshot : Return a copy of the follower list
func (config *RelayState) FollowersSnapshot() []Follower {
	config.mutex.RLock()
	defer config.mutex.RUnlock()

	return append([]Follower(nil), config.Followers...)
}

// SubscribersAndFollowersSnapshot : Return a copy of the subscriber and follower list
func (config *RelayState) SubscribersAndFollowersSnapshot() []Subscriber {
	config.mutex.RLock()
	defer config.mutex.RUnlock()

	return append([]Subscriber(nil), config.SubscribersAndFollowers...)
}

// LimitedDomainsSnapshot : Return a copy of the limited domain list
func (config *RelayState) LimitedDomainsSnapshot() []string {
	config.mutex.RLock()
	defer config.mutex.RUnlock()

	return append([]string(nil), config.LimitedDomains...)
}

// BlockedDomainsSnapshot : Return a copy of the blocked domain list
func (config *RelayState) BlockedDomainsSnapshot() []string {
	config.mutex.RLock()
	defer config.mutex.RUnlock()

	return append([]string(nil), config.BlockedDomains...)
}

// IsPersonOnly : Return whether Person-Type Actor limitation is enabled
func (config *RelayState) IsPersonOnly() bool {
	config.mutex.RLock()
	defer config.mutex.RUnlock()

	return config.RelayConfig.PersonOnly
}

// IsManuallyAccept : Return whether manual follow request acceptance is enabled
func (config *RelayState) IsManuallyAccept() bool {
	config.mutex.RLock()
	defer config.mutex.RUnlock()

	return config.RelayConfig.ManuallyAccept
}

// SetConfig : Set relay configuration
func (config *RelayState) SetConfig(key Config, value bool) {
	strValue := 0
	if value {
		strValue = 1
	}
	switch key {
	case PersonOnly:
		config.RedisClient.HSet(context.TODO(), "relay:config", "block_service", strValue).Result()
	case ManuallyAccept:
		config.RedisClient.HSet(context.TODO(), "relay:config", "manually_accept", strValue).Result()
	}

	config.refresh()
}

// AddSubscriber : Add new instance for subscriber list
func (config *RelayState) AddSubscriber(domain Subscriber) {
	config.RedisClient.HMSet(context.TODO(), "relay:subscription:"+domain.Domain, map[string]interface{}{
		"inbox_url":   domain.InboxURL,
		"activity_id": domain.ActivityID,
		"actor_id":    domain.ActorID,
	})

	config.refresh()
}

// DelSubscriber : Delete instance from subscriber list
func (config *RelayState) DelSubscriber(domain string) {
	config.RedisClient.Del(context.TODO(), "relay:subscription:"+domain).Result()
	config.RedisClient.Del(context.TODO(), "relay:pending:"+domain).Result()

	config.refresh()
}

// SelectSubscriber : Select instance from subscriber list
func (config *RelayState) SelectSubscriber(domain string) *Subscriber {
	config.mutex.RLock()
	defer config.mutex.RUnlock()

	for _, subscriber := range config.Subscribers {
		if domain == subscriber.Domain {
			return &subscriber
		}
	}
	return nil
}

// AddFollower : Add new instance for follower list
func (config *RelayState) AddFollower(domain Follower) {
	config.RedisClient.HMSet(context.TODO(), "relay:follower:"+domain.Domain, map[string]interface{}{
		"inbox_url":       domain.InboxURL,
		"activity_id":     domain.ActivityID,
		"actor_id":        domain.ActorID,
		"mutually_follow": domain.MutuallyFollow,
	})

	config.refresh()
}

// UpdateFollowerStatus : Update MutuallyFollow Status
func (config *RelayState) UpdateFollowerStatus(domain string, mutuallyFollow bool) {
	if mutuallyFollow {
		config.RedisClient.HSet(context.TODO(), "relay:follower:"+domain, "mutually_follow", "1")
	} else {
		config.RedisClient.HSet(context.TODO(), "relay:follower:"+domain, "mutually_follow", "0")
	}

	config.refresh()
}

// DelFollower : Delete instance from follower list
func (config *RelayState) DelFollower(domain string) {
	config.RedisClient.Del(context.TODO(), "relay:follower:"+domain).Result()
	config.RedisClient.Del(context.TODO(), "relay:pending:"+domain).Result()

	config.refresh()
}

// SelectFollower : Select instance from follower list
func (config *RelayState) SelectFollower(domain string) *Follower {
	config.mutex.RLock()
	defer config.mutex.RUnlock()

	for _, follower := range config.Followers {
		if domain == follower.Domain {
			return &follower
		}
	}
	return nil
}

// SetBlockedDomain : Set/Unset instance for blocked domain
func (config *RelayState) SetBlockedDomain(domain string, value bool) {
	if value {
		config.RedisClient.HSet(context.TODO(), "relay:config:blockedDomain", domain, "1").Result()
	} else {
		config.RedisClient.HDel(context.TODO(), "relay:config:blockedDomain", domain).Result()
	}

	config.refresh()
}

// SetLimitedDomain : Set/Unset instance for limited domain
func (config *RelayState) SetLimitedDomain(domain string, value bool) {
	if value {
		config.RedisClient.HSet(context.TODO(), "relay:config:limitedDomain", domain, "1").Result()
	} else {
		config.RedisClient.HDel(context.TODO(), "relay:config:limitedDomain", domain).Result()
	}

	config.refresh()
}

func (config *RelayState) refresh() {
	if config.notifiable {
		config.RedisClient.Publish(context.TODO(), "relay_refresh", nil)
	} else {
		config.Load()
	}
}

// Subscriber : Manage for Mastodon Traditional Style Relay Subscriber
type Subscriber struct {
	Domain     string `json:"domain,omitempty"`
	InboxURL   string `json:"inbox_url,omitempty"`
	ActivityID string `json:"activity_id,omitempty"`
	ActorID    string `json:"actor_id,omitempty"`
}

// Follower : Manage for LitePub Style Relay Follower
type Follower struct {
	Domain         string `json:"domain,omitempty"`
	InboxURL       string `json:"inbox_url,omitempty"`
	ActivityID     string `json:"activity_id,omitempty"`
	ActorID        string `json:"actor_id,omitempty"`
	MutuallyFollow bool   `json:"mutually_follow,omitempty"`
}

type relayConfig struct {
	PersonOnly     bool `json:"blockService,omitempty"`
	ManuallyAccept bool `json:"manuallyAccept,omitempty"`
}

func loadRelayConfig(redisClient *redis.Client) (relayConfig, error) {
	personOnly, err := redisClient.HGet(context.TODO(), "relay:config", "block_service").Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return relayConfig{}, err
	}
	manuallyAccept, err := redisClient.HGet(context.TODO(), "relay:config", "manually_accept").Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return relayConfig{}, err
	}
	return relayConfig{
		PersonOnly:     personOnly == "1",
		ManuallyAccept: manuallyAccept == "1",
	}, nil
}
