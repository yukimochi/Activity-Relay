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
	GlobalConfig *models.RelayConfig

	// RelayActor : Relay's Actor
	RelayActor models.Actor

	HttpClient      *http.Client
	MachineryServer *machinery.Server
	RedisClient     *redis.Client
)

func relayActivityV2(args ...string) error {
	inboxURL := args[0]
	activityID := args[1]
	body, err := RedisClient.HGet("relay:activity:"+activityID, "body").Result()
	if err != nil {
		return errors.New("activity ttl expired")
	}

	err = sendActivity(inboxURL, RelayActor.ID, []byte(body), GlobalConfig.ActorKey())
	if err != nil {
		domain, _ := url.Parse(inboxURL)
		pushErrorLogScript := "local change = redis.call('HSETNX', KEYS[1], 'last_error', ARGV[1]); if change == 1 then redis.call('EXPIRE', KEYS[1], ARGV[2]) end;"
		RedisClient.Eval(pushErrorLogScript, []string{"relay:statistics:" + domain.Host}, err.Error(), 60).Result()
	}
	reductionRemainCountScript := "local remain_count = redis.call('HINCRBY', KEYS[1], 'remain_count', -1); if remain_count < 1 then redis.call('DEL', KEYS[1]) end;"
	RedisClient.Eval(reductionRemainCountScript, []string{"relay:activity:" + activityID}).Result()
	return err
}

func registerActivity(args ...string) error {
	inboxURL := args[0]
	body := args[1]
	err := sendActivity(inboxURL, RelayActor.ID, []byte(body), GlobalConfig.ActorKey())
	return err
}

func nilTask(_ ...string) error {
	return nil
}

func Entrypoint(g *models.RelayConfig, v string) error {
	var err error

	version = v
	GlobalConfig = g

	err = initialize(GlobalConfig)
	if err != nil {
		return err
	}

	err = MachineryServer.RegisterTask("register", registerActivity)
	if err != nil {
		return err
	}
	err = MachineryServer.RegisterTask("relay-v2", relayActivityV2)
	if err != nil {
		return err
	}

	// From migration from <v1.2.0, remove remaining tasks
	err = MachineryServer.RegisterTask("relay-v2", nilTask)
	if err != nil {
		return err
	}

	workerID := uuid.NewV4()
	worker := MachineryServer.NewWorker(workerID.String(), GlobalConfig.JobConcurrency())
	err = worker.Launch()
	if err != nil {
		logrus.Error(err)
	}

	return nil
}

func initialize(globalConfig *models.RelayConfig) error {
	var err error

	RedisClient = globalConfig.RedisClient()

	MachineryServer, err = models.NewMachineryServer(globalConfig)
	if err != nil {
		return err
	}
	HttpClient = &http.Client{Timeout: time.Duration(5) * time.Second}

	RelayActor = models.NewActivityPubActorFromSelfKey(globalConfig)
	newNullLogger := NewNullLogger()
	log.DEBUG = newNullLogger

	return nil
}
