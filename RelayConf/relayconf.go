package relayconf

import "github.com/go-redis/redis"

// RelayConfig : struct for relay configuration
type RelayConfig struct {
	BlockService   bool
	ManuallyAccept bool
}

// LoadConfig : Loader for relay configuration
func LoadConfig(redClient *redis.Client) RelayConfig {
	blockService, err := redClient.HGet("relay:config", "block_service").Result()
	if err != nil {
		redClient.HSet("relay:config", "block_service", 0)
		blockService = "0"
	}
	manuallyAccept, err := redClient.HGet("relay:config", "manually_accept").Result()
	if err != nil {
		redClient.HSet("relay:config", "manually_accept", 0)
		manuallyAccept = "0"
	}
	return RelayConfig{
		BlockService:   blockService == "1",
		ManuallyAccept: manuallyAccept == "1",
	}
}

func SetConfig(redClient *redis.Client, key string, value bool) {
	strValue := 0
	if value {
		strValue = 1
	}
	redClient.HSet("relay:config", key, strValue)
}
