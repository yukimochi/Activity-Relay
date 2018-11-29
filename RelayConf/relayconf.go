package relayconf

import (
	"strings"

	"github.com/go-redis/redis"
)

type ExportConfig struct {
	RelayConfig    RelayConfig    `json:"relayConfig"`
	LimitedDomains []string       `json:"limitedDomains"`
	BlockedDomains []string       `json:"blockedDomains"`
	Subscriptions  []subscription `json:"subscriptions"`
}

type subscription struct {
	Domain     string `json:"domain"`
	InboxURL   string `json:"inbox_url"`
	ActivityID string `json:"activity_id"`
	ActorID    string `json:"actor_id"`
}

func (c *ExportConfig) Import(r *redis.Client) {
	c.RelayConfig.Load(r)
	var l []string
	var b []string
	var s []subscription
	domains, _ := r.HKeys("relay:config:limitedDomain").Result()
	for _, domain := range domains {
		l = append(l, domain)
	}
	domains, _ = r.HKeys("relay:config:blockedDomain").Result()
	for _, domain := range domains {
		b = append(b, domain)
	}
	domains, _ = r.Keys("relay:subscription:*").Result()
	for _, domain := range domains {
		d := strings.Replace(domain, "relay:subscription:", "", 1)
		i, _ := r.HGet(domain, "inbox_url").Result()
		s = append(s, subscription{d, i, "", ""})
	}
	c.LimitedDomains = l
	c.BlockedDomains = b
	c.Subscriptions = s
}

// RelayConfig : struct for relay configuration
type RelayConfig struct {
	BlockService     bool `json:"blockService"`
	ManuallyAccept   bool `json:"manuallyAccept"`
	CreateAsAnnounce bool `json:"createAsAnnounce"`
}

type Config int

const (
	BlockService Config = iota
	ManuallyAccept
	CreateAsAnnounce
)

func (c *RelayConfig) Load(r *redis.Client) {
	blockService, err := r.HGet("relay:config", "block_service").Result()
	if err != nil {
		c.Set(r, BlockService, false)
		blockService = "0"
	}
	manuallyAccept, err := r.HGet("relay:config", "manually_accept").Result()
	if err != nil {
		c.Set(r, ManuallyAccept, false)
		manuallyAccept = "0"
	}
	createAsAnnounce, err := r.HGet("relay:config", "create_as_announce").Result()
	if err != nil {
		c.Set(r, CreateAsAnnounce, false)
		createAsAnnounce = "0"
	}
	c.BlockService = blockService == "1"
	c.ManuallyAccept = manuallyAccept == "1"
	c.CreateAsAnnounce = createAsAnnounce == "1"
}

func (c *RelayConfig) Set(r *redis.Client, key Config, value bool) {
	strValue := 0
	if value {
		strValue = 1
	}
	switch key {
	case BlockService:
		c.BlockService = value
		r.HSet("relay:config", "block_service", strValue)
	case ManuallyAccept:
		c.ManuallyAccept = value
		r.HSet("relay:config", "manually_accept", strValue)
	case CreateAsAnnounce:
		c.CreateAsAnnounce = value
		r.HSet("relay:config", "create_as_announce", strValue)
	}
}
