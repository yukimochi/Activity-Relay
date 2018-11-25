package relayconf

import "github.com/go-redis/redis"

// RelayConfig : struct for relay configuration
type RelayConfig struct {
	BlockService     bool
	ManuallyAccept   bool
	CreateAsAnnounce bool
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
