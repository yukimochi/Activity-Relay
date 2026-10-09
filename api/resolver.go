package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/yukimochi/Activity-Relay/models"
	"github.com/yukimochi/machinery-v1/v1/tasks"
)

func contains(entries interface{}, key string) bool {
	switch entry := entries.(type) {
	case string:
		return entry == key
	case []string:
		for i := 0; i < len(entry); i++ {
			if entry[i] == key {
				return true
			}
		}
		return false
	case []models.Subscriber:
		for i := 0; i < len(entry); i++ {
			if entry[i].Domain == key {
				return true
			}
		}
		return false
	case []models.Follower:
		for i := 0; i < len(entry); i++ {
			if entry[i].Domain == key {
				return true
			}
		}
		return false
	}
	return false
}

func enqueueRegisterActivity(inboxURL string, body []byte) {
	job := &tasks.Signature{
		Name:       "register",
		RetryCount: 2,
		Args: []tasks.Arg{
			{
				Name:  "inboxURL",
				Type:  "string",
				Value: inboxURL,
			},
			{
				Name:  "body",
				Type:  "string",
				Value: string(body),
			},
		},
	}
	_, err := MachineryServer.SendTask(job)
	if err != nil {
		logrus.Error(err)
	}
}

func enqueueRelayActivity(inboxURL string, activityID string) {
	job := &tasks.Signature{
		Name:       "relay-v2",
		RetryCount: 0,
		Args: []tasks.Arg{
			{
				Name:  "inboxURL",
				Type:  "string",
				Value: inboxURL,
			},
			{
				Name:  "activityID",
				Type:  "string",
				Value: activityID,
			},
		},
	}
	_, err := MachineryServer.SendTask(job)
	if err != nil {
		logrus.Error(err)
	}
}

// pushActivityScript stores the activity body with its remaining delivery
// count and expiry.
const pushActivityScript = "redis.call('HSET',KEYS[1], 'body', ARGV[1], 'remain_count', ARGV[2]); redis.call('EXPIRE', KEYS[1], ARGV[3]);"

// pushActivity stores the activity body and enqueues a relay job for every
// subscription except the source domain.
func pushActivity(sourceDomain string, body []byte, subscriptions []models.Subscriber) {
	targets := make([]models.Subscriber, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		if subscription.Domain == sourceDomain {
			continue
		}
		targets = append(targets, subscription)
	}
	if len(targets) < 1 {
		return
	}

	activityID := uuid.New()
	_, err := RelayState.RedisClient.Eval(context.TODO(), pushActivityScript, []string{"relay:activity:" + activityID.String()}, body, len(targets), 2*60).Result()
	if err != nil {
		logrus.Error("Failed to store activity "+activityID.String()+" : ", err)
		return
	}

	for _, subscription := range targets {
		enqueueRelayActivity(subscription.InboxURL, activityID.String())
	}
}

func enqueueActivityForAll(sourceDomain string, body []byte) {
	pushActivity(sourceDomain, body, RelayState.SubscribersAndFollowersSnapshot())
}

func enqueueActivityForSubscriber(sourceDomain string, body []byte) {
	pushActivity(sourceDomain, body, RelayState.SubscribersSnapshot())
}

func enqueueActivityForFollower(sourceDomain string, body []byte) {
	followers := RelayState.FollowersSnapshot()
	subscriptions := make([]models.Subscriber, 0, len(followers))
	for _, follower := range followers {
		subscriptions = append(subscriptions, models.Subscriber{
			Domain:     follower.Domain,
			InboxURL:   follower.InboxURL,
			ActivityID: follower.ActivityID,
			ActorID:    follower.ActorID,
		})
	}
	pushActivity(sourceDomain, body, subscriptions)
}

func isActorLimited(actorID *url.URL) bool {
	if contains(RelayState.LimitedDomainsSnapshot(), actorID.Host) {
		return true
	}
	return false
}

func isActorBlocked(actorID *url.URL) bool {
	if contains(RelayState.BlockedDomainsSnapshot(), actorID.Host) {
		return true
	}
	return false
}

func isActorSubscribed(actorID *url.URL) bool {
	if contains(RelayState.SubscribersSnapshot(), actorID.Host) {
		return true
	}
	return false
}

func isActorFollowers(actorID *url.URL) bool {
	if contains(RelayState.FollowersSnapshot(), actorID.Host) {
		return true
	}
	return false
}

func isActorSubscribersOrFollowers(actorID *url.URL) bool {
	if contains(RelayState.SubscribersAndFollowersSnapshot(), actorID.Host) {
		return true
	}
	return false
}

var (
	endingWithRelayActor         = regexp.MustCompile(`/relay$`)
	endingWithFollowerURLPattern = regexp.MustCompile(`/followers$`)
)

func isActorAbleToBeFollower(actor *models.Actor) bool {
	if actor.Type == "Application" {
		return true
	}
	actorID, _ := url.Parse(actor.ID)
	return endingWithRelayActor.MatchString(actorID.Path)
}

func isActorAbleToRelay(actor *models.Actor) bool {
	domain, _ := url.Parse(actor.ID)
	if contains(RelayState.LimitedDomainsSnapshot(), domain.Host) {
		return false
	}
	if RelayState.IsPersonOnly() && actor.Type != "Person" {
		return false
	}
	return true
}

func isToMyFollower(entries []string) bool {
	followers := RelayState.FollowersSnapshot()
	for _, entry := range entries {
		if endingWithFollowerURLPattern.MatchString(entry) {
			for _, follower := range followers {
				if follower.ActorID+"/followers" == entry {
					return true
				}
			}
		}
	}
	return false
}

func executeFollowing(activity *models.Activity, actor *models.Actor) error {
	actorID, _ := url.Parse(actor.ID)
	if isActorBlocked(actorID) {
		return errors.New(actorID.Host + " is blocked")
	}
	switch {
	case contains(activity.Object, "https://www.w3.org/ns/activitystreams#Public"):
		if RelayState.IsManuallyAccept() {
			object, ok := activity.Object.(string)
			if !ok {
				return errors.New("activity object is not a string")
			}
			RelayState.RedisClient.HMSet(context.TODO(), "relay:pending:"+actorID.Host, map[string]interface{}{
				"inbox_url":   actor.Endpoints.SharedInbox,
				"activity_id": activity.ID,
				"type":        "Follow",
				"actor":       actor.ID,
				"object":      object,
			})
			logrus.Info("Pending Follow Request : ", activity.Actor)
		} else {
			resp := activity.GenerateReply(RelayActor, activity, "Accept")
			jsonData, _ := json.Marshal(&resp)
			go enqueueRegisterActivity(actor.Inbox, jsonData)
			RelayState.AddSubscriber(models.Subscriber{
				Domain:     actorID.Host,
				InboxURL:   actor.Endpoints.SharedInbox,
				ActivityID: activity.ID,
				ActorID:    actor.ID,
			})
			logrus.Info("Accepted Follow Request : ", activity.Actor)
		}
	case contains(activity.Object, RelayActor.ID):
		if isActorAbleToBeFollower(actor) {
			if RelayState.IsManuallyAccept() {
				object, ok := activity.Object.(string)
				if !ok {
					return errors.New("activity object is not a string")
				}
				RelayState.RedisClient.HMSet(context.TODO(), "relay:pending:"+actorID.Host, map[string]interface{}{
					"inbox_url":   actor.Inbox,
					"activity_id": activity.ID,
					"type":        "Follow",
					"actor":       actor.ID,
					"object":      object,
				})
				logrus.Info("Pending Follow Request : ", activity.Actor)
			} else {
				resp := activity.GenerateReply(RelayActor, activity, "Accept")
				jsonData, _ := json.Marshal(&resp)
				go enqueueRegisterActivity(actor.Inbox, jsonData)
				follower := models.Follower{
					Domain:         actorID.Host,
					InboxURL:       actor.Inbox,
					ActivityID:     activity.ID,
					ActorID:        actor.ID,
					MutuallyFollow: false,
				}
				RelayState.AddFollower(follower)
				logrus.Info("Accepted Follow Request : ", activity.Actor)

				executeMutuallyFollow(follower)
			}
			return nil
		}
		fallthrough
	default:
		err := errors.New("only https://www.w3.org/ns/activitystreams#Public is allowed to follow")
		return err
	}
	return nil
}

func executeUnfollowing(activity *models.Activity, actor *models.Actor) error {
	actorID, _ := url.Parse(actor.ID)
	switch {
	case contains(activity.Object, "https://www.w3.org/ns/activitystreams#Public"):
		RelayState.DelSubscriber(actorID.Host)
		logrus.Info("Accepted Unfollow Request : ", activity.Actor)
		return nil
	case contains(activity.Object, RelayActor.ID):
		if isActorAbleToBeFollower(actor) {
			RelayState.DelFollower(actorID.Host)
			logrus.Info("Accepted Unfollow Request : ", activity.Actor)
			return nil
		}
		fallthrough
	default:
		err := errors.New("only https://www.w3.org/ns/activitystreams#Public is allowed to unfollow")
		return err
	}
}

func executeMutuallyFollow(follower models.Follower) error {
	actorID, _ := url.Parse(follower.ActorID)
	if !isActorLimited(actorID) {
		followRequest := models.NewActivityPubActivity(RelayActor, []string{follower.ActorID}, follower.ActorID, "Follow")
		jsonData, _ := json.Marshal(&followRequest)
		go enqueueRegisterActivity(follower.InboxURL, jsonData)
		logrus.Info("Sent MutuallyFollow Request : ", follower.ActorID)
	}
	return nil
}

func finalizeMutuallyFollow(activity *models.Activity, actor *models.Actor, activityType string) {
	actorID, _ := url.Parse(actor.ID)
	if contains(activity.Actor, RelayActor.ID) && contains(activity.Object, actor.ID) && isActorFollowers(actorID) {
		RelayState.UpdateFollowerStatus(actorID.Host, activityType == "Accept")
		logrus.Info("Confirmed MutuallyFollow "+activityType+"ed : ", actor.ID)
	}
}

func executeRejectRequest(activity *models.Activity, actor *models.Actor, err error) {
	reject := activity.GenerateReply(RelayActor, activity, "Reject")
	jsonData, _ := json.Marshal(&reject)
	go enqueueRegisterActivity(actor.Inbox, jsonData)
	logrus.Error("Rejected Follow, Unfollow Request : ", activity.Actor, " ", err.Error())
}

func executeRelayActivity(activity *models.Activity, actor *models.Actor, body []byte) error {
	actorID, _ := url.Parse(actor.ID)
	if !isActorSubscribersOrFollowers(actorID) {
		err := errors.New("to use the relay service, please follow in advance")
		return err
	}
	if isActorAbleToRelay(actor) {
		go enqueueActivityForSubscriber(actorID.Host, body)

		var innnerObjectId, err = activity.UnwrapInnerObjectId()
		if err != nil {
			logrus.Debug("Accepted Relay Activity (Announce Failed) : ", activity.Actor)
		} else {
			announce := models.NewActivityPubActivity(RelayActor, []string{RelayActor.Followers()}, innnerObjectId, "Announce")
			jsonData, _ := json.Marshal(&announce)
			go enqueueActivityForFollower(actorID.Host, jsonData)
			logrus.Debug("Accepted Relay Activity : ", activity.Actor)
		}
	} else {
		logrus.Debug("Skipped Relay Activity : ", activity.Actor)
	}
	return nil
}

func executeAnnounceActivity(activity *models.Activity, actor *models.Actor) error {
	actorID, _ := url.Parse(actor.ID)
	if isActorAbleToRelay(actor) {
		announce := models.NewActivityPubActivity(RelayActor, []string{RelayActor.Followers()}, activity.ID, "Announce")
		jsonData, _ := json.Marshal(&announce)
		go enqueueActivityForAll(actorID.Host, jsonData)
		logrus.Debug("Accepted Announce Activity : ", activity.Actor)
	} else {
		logrus.Debug("Skipped Announce Activity : ", activity.Actor)
	}
	return nil
}
