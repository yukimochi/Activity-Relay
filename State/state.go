package state

import (
	"strings"

	"github.com/go-redis/redis"
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

// NewState : Create new RelayState instance with redis client
func NewState(redisClient *redis.Client) RelayState {
	var config RelayState
	config.RedisClient = redisClient

	config.Load()
	return config
}

// RelayState : Store subscriptions and relay configrations
type RelayState struct {
	RedisClient *redis.Client

	RelayConfig    relayConfig    `json:"relayConfig"`
	LimitedDomains []string       `json:"limitedDomains"`
	BlockedDomains []string       `json:"blockedDomains"`
	Subscriptions  []Subscription `json:"subscriptions"`
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

// SetConfig : Set relay configration
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
	config.Load()
}

// AddSubscription : Add new instance for subscription list
func (config *RelayState) AddSubscription(domain Subscription) {
	config.RedisClient.HMSet("relay:subscription:"+domain.Domain, map[string]interface{}{
		"inbox_url":   domain.InboxURL,
		"activity_id": domain.ActivityID,
		"actor_id":    domain.ActorID,
	})

	config.Load()
}

// DelSubscription : Delete instance from subscription list
func (config *RelayState) DelSubscription(domain string) {
	config.RedisClient.Del("relay:subscription:" + domain).Result()
	config.RedisClient.Del("relay:pending:" + domain).Result()

	config.Load()
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

	config.Load()
}

// SetLimitedDomain : Set/Unset instance for limited domain
func (config *RelayState) SetLimitedDomain(domain string, value bool) {
	if value {
		config.RedisClient.HSet("relay:config:limitedDomain", domain, "1").Result()
	} else {
		config.RedisClient.HDel("relay:config:limitedDomain", domain).Result()
	}

	config.Load()
}

// Subscription : Instance subscription information
type Subscription struct {
	Domain     string `json:"domain"`
	InboxURL   string `json:"inbox_url"`
	ActivityID string `json:"activity_id"`
	ActorID    string `json:"actor_id"`
}

type relayConfig struct {
	BlockService     bool `json:"blockService"`
	ManuallyAccept   bool `json:"manuallyAccept"`
	CreateAsAnnounce bool `json:"createAsAnnounce"`
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
