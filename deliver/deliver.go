package deliver

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/log"
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
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

func relayActivityV2(args ...string) error {
	inboxURL := args[0]
	activityID := args[1]
	body, err := redisClient.HGet("relay:activity:"+activityID, "body").Result()
	if err != nil {
		return errors.New("this activity is expired")
	}

	err = sendActivity(inboxURL, Actor.ID, []byte(body), globalConfig.ActorKey())
	if err != nil {
		domain, _ := url.Parse(inboxURL)
		evalScript := "local change = redis.call('HSETNX', KEYS[1], 'last_error', ARGV[1]); if change == 1 then redis.call('EXPIRE', KEYS[1], ARGV[2]) end;"
		redisClient.Eval(evalScript, []string{"relay:statistics:" + domain.Host}, err.Error(), 60).Result()
	}
	evalScript := "local remain_count = redis.call('HINCRBY', KEYS[1], 'remain_count', -1); if remain_count < 1 then redis.call('DEL', KEYS[1]) end;"
	redisClient.Eval(evalScript, []string{"relay:activity:" + activityID}).Result()
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
	err = machineryServer.RegisterTask("relay-v2", relayActivityV2)
	if err != nil {
		return err
	}

	workerID := uuid.NewV4()
	worker := machineryServer.NewWorker(workerID.String(), globalConfig.JobConcurrency())
	err = worker.Launch()
	if err != nil {
		logrus.Error(err)
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
