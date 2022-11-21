package models

import (
	"strings"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

// Config : Enum for RelayConfig
type Config int

const (
	// BlockService : Blocking for service-type actor
	BlockService Config = iota
	// ManuallyAccept : Manually accept follow-request
	ManuallyAccept
	// CreateAsAnnounce : Announce activity instead of relay create activity
	CreateAsAnnounce
)

// RelayState : Store subscriptions and relay configurations
type RelayState struct {
	RedisClient *redis.Client
	notifiable  bool

	RelayConfig    relayConfig    `json:"relayConfig,omitempty"`
	LimitedDomains []string       `json:"limitedDomains,omitempty"`
	BlockedDomains []string       `json:"blockedDomains,omitempty"`
	Subscriptions  []Subscription `json:"subscriptions,omitempty"`
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
	_, err := config.RedisClient.Subscribe("relay_refresh").Receive()
	if err != nil {
		panic(err)
	}
	ch := config.RedisClient.Subscribe("relay_refresh").Channel()

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
	var subscriptions []Subscription
	domains, _ := config.RedisClient.HKeys("relay:config:limitedDomain").Result()
	for _, domain := range domains {
		limitedDomains = append(limitedDomains, domain)
	}
	domains, _ = config.RedisClient.HKeys("relay:config:blockedDomain").Result()
	for _, domain := range domains {
		blockedDomains = append(blockedDomains, domain)
	}
	domains, _ = config.RedisClient.Keys("relay:subscription:*").Result()
	for _, domain := range domains {
		domainName := strings.Replace(domain, "relay:subscription:", "", 1)
		inboxURL, _ := config.RedisClient.HGet(domain, "inbox_url").Result()
		activityID, err := config.RedisClient.HGet(domain, "activity_id").Result()
		if err != nil {
			activityID = ""
		}
		actorID, err := config.RedisClient.HGet(domain, "actor_id").Result()
		if err != nil {
			actorID = ""
		}
		subscriptions = append(subscriptions, Subscription{domainName, inboxURL, activityID, actorID})
	}
	config.LimitedDomains = limitedDomains
	config.BlockedDomains = blockedDomains
	config.Subscriptions = subscriptions
}

// SetConfig : Set relay configuration
func (config *RelayState) SetConfig(key Config, value bool) {
	strValue := 0
	if value {
		strValue = 1
	}
	switch key {
	case BlockService:
		config.RedisClient.HSet("relay:config", "block_service", strValue).Result()
	case ManuallyAccept:
		config.RedisClient.HSet("relay:config", "manually_accept", strValue).Result()
	case CreateAsAnnounce:
		config.RedisClient.HSet("relay:config", "create_as_announce", strValue).Result()
	}

	config.refresh()
}

// AddSubscription : Add new instance for subscription list
func (config *RelayState) AddSubscription(domain Subscription) {
	config.RedisClient.HMSet("relay:subscription:"+domain.Domain, map[string]interface{}{
		"inbox_url":   domain.InboxURL,
		"activity_id": domain.ActivityID,
		"actor_id":    domain.ActorID,
	})

	config.refresh()
}

// DelSubscription : Delete instance from subscription list
func (config *RelayState) DelSubscription(domain string) {
	config.RedisClient.Del("relay:subscription:" + domain).Result()
	config.RedisClient.Del("relay:pending:" + domain).Result()

	config.refresh()
}

// SelectSubscription : Select instance from string
func (config *RelayState) SelectSubscription(domain string) *Subscription {
	for _, subscription := range config.Subscriptions {
		if domain == subscription.Domain {
			return &subscription
		}
	}
	return nil
}

// SetBlockedDomain : Set/Unset instance for blocked domain
func (config *RelayState) SetBlockedDomain(domain string, value bool) {
	if value {
		config.RedisClient.HSet("relay:config:blockedDomain", domain, "1").Result()
	} else {
		config.RedisClient.HDel("relay:config:blockedDomain", domain).Result()
	}

	config.refresh()
}

// SetLimitedDomain : Set/Unset instance for limited domain
func (config *RelayState) SetLimitedDomain(domain string, value bool) {
	if value {
		config.RedisClient.HSet("relay:config:limitedDomain", domain, "1").Result()
	} else {
		config.RedisClient.HDel("relay:config:limitedDomain", domain).Result()
	}

	config.refresh()
}

func (config *RelayState) refresh() {
	if config.notifiable {
		config.RedisClient.Publish("relay_refresh", nil)
	} else {
		config.Load()
	}
}

// Subscription : Instance subscription information
type Subscription struct {
	Domain     string `json:"domain,omitempty"`
	InboxURL   string `json:"inbox_url,omitempty"`
	ActivityID string `json:"activity_id,omitempty"`
	ActorID    string `json:"actor_id,omitempty"`
}

type relayConfig struct {
	BlockService     bool `json:"blockService,omitempty"`
	ManuallyAccept   bool `json:"manuallyAccept,omitempty"`
	CreateAsAnnounce bool `json:"createAsAnnounce,omitempty"`
}

func (config *relayConfig) load(redisClient *redis.Client) {
	blockService, err := redisClient.HGet("relay:config", "block_service").Result()
	if err != nil {
		blockService = "0"
	}
	manuallyAccept, err := redisClient.HGet("relay:config", "manually_accept").Result()
	if err != nil {
		manuallyAccept = "0"
	}
	createAsAnnounce, err := redisClient.HGet("relay:config", "create_as_announce").Result()
	if err != nil {
		createAsAnnounce = "0"
	}
	config.BlockService = blockService == "1"
	config.ManuallyAccept = manuallyAccept == "1"
	config.CreateAsAnnounce = createAsAnnounce == "1"
}
