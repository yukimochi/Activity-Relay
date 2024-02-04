package models

import (
	"context"
	"strings"

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

	RelayConfig             relayConfig  `json:"relayConfig,omitempty"`
	LimitedDomains          []string     `json:"limitedDomains,omitempty"`
	BlockedDomains          []string     `json:"blockedDomains,omitempty"`
	Subscribers             []Subscriber `json:"subscriptions,omitempty"`
	Followers               []Follower   `json:"followers,omitempty"`
	SubscribersAndFollowers []Subscriber `json:"-"`
}

// NewState : Create new RelayState instance with redis client
func NewState(redisClient *redis.Client, notifiable bool) RelayState {
	var config RelayState
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

// Load : Refrash content from redis
func (config *RelayState) Load() {
	config.RelayConfig.load(config.RedisClient)
	var limitedDomains []string
	var blockedDomains []string
	var subscribers []Subscriber
	var followers []Follower
	var subscribersAndFollowers []Subscriber

	domains, _ := config.RedisClient.HKeys(context.TODO(), "relay:config:limitedDomain").Result()
	for _, domain := range domains {
		limitedDomains = append(limitedDomains, domain)
	}
	domains, _ = config.RedisClient.HKeys(context.TODO(), "relay:config:blockedDomain").Result()
	for _, domain := range domains {
		blockedDomains = append(blockedDomains, domain)
	}

	domains, _ = config.RedisClient.Keys(context.TODO(), "relay:subscription:*").Result()
	for _, domain := range domains {
		domainName := strings.Replace(domain, "relay:subscription:", "", 1)
		inboxURL, _ := config.RedisClient.HGet(context.TODO(), domain, "inbox_url").Result()
		activityID, err := config.RedisClient.HGet(context.TODO(), domain, "activity_id").Result()
		if err != nil {
			activityID = ""
		}
		actorID, err := config.RedisClient.HGet(context.TODO(), domain, "actor_id").Result()
		if err != nil {
			actorID = ""
		}
		subscribers = append(subscribers, Subscriber{domainName, inboxURL, activityID, actorID})
		subscribersAndFollowers = append(subscribersAndFollowers, Subscriber{domainName, inboxURL, activityID, actorID})
	}

	domains, _ = config.RedisClient.Keys(context.TODO(), "relay:follower:*").Result()
	for _, domain := range domains {
		domainName := strings.Replace(domain, "relay:follower:", "", 1)
		inboxURL, _ := config.RedisClient.HGet(context.TODO(), domain, "inbox_url").Result()
		activityID, err := config.RedisClient.HGet(context.TODO(), domain, "activity_id").Result()
		if err != nil {
			activityID = ""
		}
		actorID, err := config.RedisClient.HGet(context.TODO(), domain, "actor_id").Result()
		if err != nil {
			actorID = ""
		}
		mutuallyFollow, err := config.RedisClient.HGet(context.TODO(), domain, "mutually_follow").Result()
		if err != nil {
			mutuallyFollow = "0"
		}
		followers = append(followers, Follower{domainName, inboxURL, activityID, actorID, mutuallyFollow == "1"})
		subscribersAndFollowers = append(subscribersAndFollowers, Subscriber{domainName, inboxURL, activityID, actorID})
	}

	config.LimitedDomains = limitedDomains
	config.BlockedDomains = blockedDomains
	config.Subscribers = subscribers
	config.Followers = followers
	config.SubscribersAndFollowers = subscribersAndFollowers
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

func (config *relayConfig) load(redisClient *redis.Client) {
	personOnly, err := redisClient.HGet(context.TODO(), "relay:config", "block_service").Result()
	if err != nil {
		personOnly = "0"
	}
	manuallyAccept, err := redisClient.HGet(context.TODO(), "relay:config", "manually_accept").Result()
	if err != nil {
		manuallyAccept = "0"
	}
	config.PersonOnly = personOnly == "1"
	config.ManuallyAccept = manuallyAccept == "1"
}
