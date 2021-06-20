package deliver

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/log"
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
	"github.com/yukimochi/Activity-Relay/models"
)

var (
	version      string
	globalConfig *models.RelayConfig

	// Actor : Relay's Actor
	Actor models.Actor

	redisClient     *redis.Client
	machineryServer *machinery.Server
	httpClient      *http.Client
)

func relayActivity(args ...string) error {
	inboxURL := args[0]
	body := args[1]
	err := sendActivity(inboxURL, Actor.ID, []byte(body), globalConfig.ActorKey())
	if err != nil {
		domain, _ := url.Parse(inboxURL)
		evalScript := "local change = redis.call('HSETNX',KEYS[1], 'last_error', ARGV[1]); if change == 1 then redis.call('EXPIRE', KEYS[1], ARGV[2]) end;"
		redisClient.Eval(evalScript, []string{"relay:statistics:" + domain.Host}, err.Error(), 60).Result()
	}
	return err
}

func registerActivity(args ...string) error {
	inboxURL := args[0]
	body := args[1]
	err := sendActivity(inboxURL, Actor.ID, []byte(body), globalConfig.ActorKey())
	return err
}

func Entrypoint(g *models.RelayConfig, v string) error {
	var err error
	globalConfig = g
	version = v

	err = initialize(globalConfig)
	if err != nil {
		return err
	}

	err = machineryServer.RegisterTask("register", registerActivity)
	if err != nil {
		return err
	}
	err = machineryServer.RegisterTask("relay", relayActivity)
	if err != nil {
		return err
	}

	workerID := uuid.NewV4()
	worker := machineryServer.NewWorker(workerID.String(), globalConfig.JobConcurrency())
	err = worker.Launch()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	return nil
}

func initialize(globalConfig *models.RelayConfig) error {
	var err error

	redisClient = globalConfig.RedisClient()

	machineryServer, err = models.NewMachineryServer(globalConfig)
	if err != nil {
		return err
	}
	httpClient = &http.Client{Timeout: time.Duration(5) * time.Second}

	Actor = models.NewActivityPubActorFromSelfKey(globalConfig)
	newNullLogger := NewNullLogger()
	log.DEBUG = newNullLogger

	return nil
}
