package api

import (
	"encoding/json"
	"errors"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/yukimochi/Activity-Relay/models"
	"github.com/yukimochi/machinery-v1/v1/tasks"
)

func handleWebfinger(writer http.ResponseWriter, request *http.Request) {
	queriedResource := request.URL.Query()["resource"]
	if request.Method != "GET" || len(queriedResource) == 0 {
		writer.WriteHeader(400)
		writer.Write(nil)
	} else {
		queriedSubject := queriedResource[0]
		for _, webfingerResource := range WebfingerResources {
			if queriedSubject == webfingerResource.Subject {
				webfinger, err := json.Marshal(&webfingerResource)
				if err != nil {
					logrus.Fatal("Failed to marshal webfinger resource : ", err.Error())
					writer.WriteHeader(500)
					writer.Write(nil)
					return
				}
				writer.Header().Add("Content-Type", "application/json")
				writer.WriteHeader(200)
				writer.Write(webfinger)
				return
			}
		}
		writer.WriteHeader(404)
		writer.Write(nil)
	}
}

func handleNodeinfoLink(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		writer.WriteHeader(400)
		writer.Write(nil)
	} else {
		nodeinfoLinks, err := json.Marshal(&Nodeinfo.NodeinfoLinks)
		if err != nil {
			logrus.Fatal("Failed to marshal nodeinfo links : ", err.Error())
			writer.WriteHeader(500)
			writer.Write(nil)
			return
		}
		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
		writer.Write(nodeinfoLinks)
	}
}

func handleNodeinfo(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		writer.WriteHeader(400)
		writer.Write(nil)
	} else {
		userTotal := len(RelayState.Subscriptions)
		Nodeinfo.Nodeinfo.Usage.Users.Total = userTotal
		Nodeinfo.Nodeinfo.Usage.Users.ActiveMonth = userTotal
		Nodeinfo.Nodeinfo.Usage.Users.ActiveHalfyear = userTotal
		nodeinfo, err := json.Marshal(&Nodeinfo.Nodeinfo)
		if err != nil {
			logrus.Fatal("Failed to marshal nodeinfo : ", err.Error())
			writer.WriteHeader(500)
			writer.Write(nil)
			return
		}
		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
		writer.Write(nodeinfo)
	}
}

func handleRelayActor(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		relayActor, err := json.Marshal(&RelayActor)
		if err != nil {
			logrus.Fatal("Failed to marshal relay actor : ", err.Error())
			writer.WriteHeader(500)
			writer.Write(nil)
			return
		}
		writer.Header().Add("Content-Type", "application/activity+json")
		writer.WriteHeader(200)
		writer.Write(relayActor)
	} else {
		writer.WriteHeader(400)
		writer.Write(nil)
	}
}

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
	case []models.Subscription:
		for i := 0; i < len(entry); i++ {
			if entry[i].Domain == key {
				return true
			}
		}
		return false
	}
	return false
}

func enqueueRelayActivity(sourceDomain string, body []byte) {
	activityID := uuid.NewV4()
	remainCount := len(RelayState.Subscriptions) - 1

	if remainCount < 1 {
		return
	}

	pushActivityScript := "redis.call('HSET',KEYS[1], 'body', ARGV[1], 'remain_count', ARGV[2]); redis.call('EXPIRE', KEYS[1], ARGV[3]);"
	RelayState.RedisClient.Eval(pushActivityScript, []string{"relay:activity:" + activityID.String()}, body, remainCount, 2*60).Result()

	for _, subscription := range RelayState.Subscriptions {
		if sourceDomain == subscription.Domain {
			continue
		}

		job := &tasks.Signature{
			Name:       "relay-v2",
			RetryCount: 0,
			Args: []tasks.Arg{
				{
					Name:  "inboxURL",
					Type:  "string",
					Value: subscription.InboxURL,
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

func isActorBlocked(actorID *url.URL) bool {
	if contains(RelayState.BlockedDomains, actorID.Host) {
		return true
	}
	return false
}

func isActorSubscribed(actorID *url.URL) bool {
	if contains(RelayState.Subscriptions, actorID.Host) {
		return true
	}
	return false
}

func isActivityAbleToRelay(activity *models.Activity, actor *models.Actor) bool {
	domain, _ := url.Parse(activity.Actor)
	if contains(RelayState.LimitedDomains, domain.Host) {
		return false
	}
	if RelayState.RelayConfig.BlockService && actor.Type != "Person" {
		return false
	}
	return true
}

func executeFollowing(activity *models.Activity, actor *models.Actor) error {
	actorID, _ := url.Parse(actor.ID)
	if isActorBlocked(actorID) {
		return errors.New(actorID.Host + " is blocked")
	}
	switch {
	case contains(activity.Object, "https://www.w3.org/ns/activitystreams#Public"):
		if RelayState.RelayConfig.ManuallyAccept {
			RelayState.RedisClient.HMSet("relay:pending:"+actorID.Host, map[string]interface{}{
				"inbox_url":   actor.Endpoints.SharedInbox,
				"activity_id": activity.ID,
				"type":        "Follow",
				"actor":       actor.ID,
				"object":      activity.Object.(string),
			})
			logrus.Info("Pending Follow Request : ", activity.Actor)
		} else {
			resp := activity.GenerateReply(RelayActor, activity, "Accept")
			jsonData, _ := json.Marshal(&resp)
			go enqueueRegisterActivity(actor.Inbox, jsonData)
			RelayState.AddSubscription(models.Subscription{
				Domain:     actorID.Host,
				InboxURL:   actor.Endpoints.SharedInbox,
				ActivityID: activity.ID,
				ActorID:    actor.ID,
			})
			logrus.Info("Accepted Follow Request : ", activity.Actor)
		}
	case contains(activity.Object, RelayActor.ID):
		fallthrough
	default:
		err := errors.New("only https://www.w3.org/ns/activitystreams#Public is allowed to unfollow")
		return err
	}
	return nil
}

func executeUnfollowing(activity *models.Activity, actor *models.Actor) error {
	actorID, _ := url.Parse(actor.ID)
	switch {
	case contains(activity.Object, "https://www.w3.org/ns/activitystreams#Public"):
		RelayState.DelSubscription(actorID.Host)
		logrus.Info("Accepted Unfollow Request : ", activity.Actor)
		return nil
	case contains(activity.Object, RelayActor.ID):
		fallthrough
	default:
		err := errors.New("only https://www.w3.org/ns/activitystreams#Public is allowed to unfollow")
		return err
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
	if !isActorSubscribed(actorID) {
		err := errors.New("to use the relay service, please follow in advance")
		return err
	}
	if isActivityAbleToRelay(activity, actor) {
		if RelayState.RelayConfig.CreateAsAnnounce && activity.Type == "Create" {
			nestedObject, err := activity.UnwrapInnerActivity()
			if err != nil {
				logrus.Error("Failed to decode inner activity : ", err.Error())
			} else {
				announce := models.NewActivityPubActivity(RelayActor, []string{RelayActor.Followers()}, nestedObject.ID, "Announce")
				jsonData, _ := json.Marshal(&announce)
				go enqueueRelayActivity(actorID.Host, jsonData)
				logrus.Debug("Accepted Announce ", nestedObject.Type, " : ", activity.Actor)
			}
		} else {
			go enqueueRelayActivity(actorID.Host, body)
			logrus.Debug("Accepted Relay Activity : ", activity.Actor)
		}
	} else {
		logrus.Debug("Skipped Relay Activity : ", activity.Actor)
	}
	return nil
}

func handleInbox(writer http.ResponseWriter, request *http.Request, activityDecoder func(*http.Request) (*models.Activity, *models.Actor, []byte, error)) {
	switch request.Method {
	case "POST":
		activity, actor, body, err := activityDecoder(request)
		if err != nil {
			writer.WriteHeader(400)
			writer.Write(nil)
		} else {
			actorID, _ := url.Parse(activity.Actor)
			switch {
			case contains(activity.To, "https://www.w3.org/ns/activitystreams#Public"), contains(activity.Cc, "https://www.w3.org/ns/activitystreams#Public"):
				// Mastodon Traditional Style (Activity Transfer)
				switch activity.Type {
				case "Create", "Update", "Delete", "Move":
					err = executeRelayActivity(activity, actor, body)
					if err != nil {
						writer.WriteHeader(401)
						writer.Write([]byte(err.Error()))

						return
					}
					writer.WriteHeader(202)
					writer.Write(nil)
				default:
					writer.WriteHeader(202)
					writer.Write(nil)
				}
			case contains(activity.To, RelayActor.ID), contains(activity.Cc, RelayActor.ID):
				// LitePub Relay Style
				switch activity.Type {
				case "Follow":
					err = executeFollowing(activity, actor)
					if err != nil {
						executeRejectRequest(activity, actor, err)
					}
					writer.WriteHeader(202)
					writer.Write(nil)
				case "Undo":
					innerActivity, _ := activity.UnwrapInnerActivity()
					switch innerActivity.Type {
					case "Follow":
						err = executeUnfollowing(activity, actor)
						if err != nil {
							executeRejectRequest(activity, actor, err)
						}
						writer.WriteHeader(202)
						writer.Write(nil)
					default:
						writer.WriteHeader(202)
						writer.Write(nil)
					}
				case "Announce":
					if !isActorSubscribed(actorID) {
						err = errors.New("to use the relay service, please follow in advance")
						writer.WriteHeader(401)
						writer.Write([]byte(err.Error()))

						return
					}
					err = errors.New("recipient must include https://www.w3.org/ns/activitystreams#Public")
					writer.WriteHeader(400)
					writer.Write([]byte(err.Error()))
				default:
					writer.WriteHeader(202)
					writer.Write(nil)
				}
			default:
				// Follow, Unfollow Only
				switch activity.Type {
				case "Follow":
					err = executeFollowing(activity, actor)
					if err != nil {
						executeRejectRequest(activity, actor, err)
					}
					writer.WriteHeader(202)
					writer.Write(nil)
				case "Undo":
					innerActivity, _ := activity.UnwrapInnerActivity()
					switch innerActivity.Type {
					case "Follow":
						err = executeUnfollowing(innerActivity, actor)
						if err != nil {
							executeRejectRequest(activity, actor, err)
						}
						writer.WriteHeader(202)
						writer.Write(nil)
					default:
						writer.WriteHeader(202)
						writer.Write(nil)
					}
				default:
					writer.WriteHeader(202)
					writer.Write(nil)
				}
			}
		}
	default:
		writer.WriteHeader(405)
		writer.Write(nil)
	}
}
