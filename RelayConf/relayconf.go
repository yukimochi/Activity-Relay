package relayconf

import (
	"strings"

	"github.com/go-redis/redis"
)

type Config int

const (
	BlockService Config = iota
	ManuallyAccept
	CreateAsAnnounce
)

func NewConfig(r *redis.Client) ExportConfig {
	var config ExportConfig
	config.RedisClient = r

	config.Load()
	return config
}

type ExportConfig struct {
	RedisClient *redis.Client

	RelayConfig    relayConfig    `json:"relayConfig"`
	LimitedDomains []string       `json:"limitedDomains"`
	BlockedDomains []string       `json:"blockedDomains"`
	Subscriptions  []Subscription `json:"subscriptions"`
}

func (config *ExportConfig) Load() {
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

func (config *ExportConfig) SetConfig(key Config, value bool) {
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

func (config *ExportConfig) AddSubscription(domain Subscription) {
	config.RedisClient.HMSet("relay:subscription:"+domain.Domain, map[string]interface{}{
		"inbox_url":   domain.InboxURL,
		"activity_id": domain.ActivityID,
		"actor_id":    domain.ActorID,
	})

	config.Load()
}

func (config *ExportConfig) DelSubscription(domain string) {
	config.RedisClient.Del("relay:subscription:" + domain).Result()
	config.RedisClient.Del("relay:pending:" + domain).Result()

	config.Load()
}

func (config *ExportConfig) SetBlockedDomain(domain string, value bool) {
	if value {
		config.RedisClient.HSet("relay:config:blockedDomain", domain, "1").Result()
	} else {
		config.RedisClient.HDel("relay:config:blockedDomain", domain).Result()
	}

	config.Load()
}

func (config *ExportConfig) SetLimitedDomain(domain string, value bool) {
	if value {
		config.RedisClient.HSet("relay:config:limitedDomain", domain, "1").Result()
	} else {
		config.RedisClient.HDel("relay:config:limitedDomain", domain).Result()
	}

	config.Load()
}

type Subscription struct {
	Domain     string `json:"domain"`
	InboxURL   string `json:"inbox_url"`
	ActivityID string `json:"activity_id"`
	ActorID    string `json:"actor_id"`
}

// RelayConfig : struct for relay configuration
type relayConfig struct {
	BlockService     bool `json:"blockService"`
	ManuallyAccept   bool `json:"manuallyAccept"`
	CreateAsAnnounce bool `json:"createAsAnnounce"`
}

func (config *relayConfig) load(r *redis.Client) {
	blockService, err := r.HGet("relay:config", "block_service").Result()
	if err != nil {
		blockService = "0"
	}
	manuallyAccept, err := r.HGet("relay:config", "manually_accept").Result()
	if err != nil {
		manuallyAccept = "0"
	}
	createAsAnnounce, err := r.HGet("relay:config", "create_as_announce").Result()
	if err != nil {
		createAsAnnounce = "0"
	}
	config.BlockService = blockService == "1"
	config.ManuallyAccept = manuallyAccept == "1"
	config.CreateAsAnnounce = createAsAnnounce == "1"
}
