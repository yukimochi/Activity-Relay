package api

import (
	"encoding/json"
	"errors"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"net/url"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/sirupsen/logrus"
	"github.com/yukimochi/Activity-Relay/models"
)

func handleWebfinger(writer http.ResponseWriter, request *http.Request) {
	resource := request.URL.Query()["resource"]
	if request.Method != "GET" || len(resource) == 0 {
		writer.WriteHeader(400)
		writer.Write(nil)
	} else {
		request := resource[0]
		if request == WebfingerResource.Subject {
			webfingerResource, err := json.Marshal(&WebfingerResource)
			if err != nil {
				panic(err)
			}
			writer.Header().Add("Content-Type", "application/json")
			writer.WriteHeader(200)
			writer.Write(webfingerResource)
		} else {
			writer.WriteHeader(404)
			writer.Write(nil)
		}
	}
}

func handleNodeinfoLink(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		writer.WriteHeader(400)
		writer.Write(nil)
	} else {
		linksResource, err := json.Marshal(&Nodeinfo.NodeinfoLinks)
		if err != nil {
			panic(err)
		}
		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
		writer.Write(linksResource)
	}
}

func handleNodeinfo(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		writer.WriteHeader(400)
		writer.Write(nil)
	} else {
		userCount := len(relayState.Subscriptions)
		Nodeinfo.Nodeinfo.Usage.Users.Total = userCount
		Nodeinfo.Nodeinfo.Usage.Users.ActiveMonth = userCount
		Nodeinfo.Nodeinfo.Usage.Users.ActiveHalfyear = userCount
		linksResource, err := json.Marshal(&Nodeinfo.Nodeinfo)
		if err != nil {
			panic(err)
		}
		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
		writer.Write(linksResource)
	}
}

func handleActor(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		actor, err := json.Marshal(&Actor)
		if err != nil {
			panic(err)
		}
		writer.Header().Add("Content-Type", "application/activity+json")
		writer.WriteHeader(200)
		writer.Write(actor)
	} else {
		writer.WriteHeader(400)
		writer.Write(nil)
	}
}

func contains(entries interface{}, finder string) bool {
	switch entry := entries.(type) {
	case string:
		return entry == finder
	case []string:
		for i := 0; i < len(entry); i++ {
			if entry[i] == finder {
				return true
			}
		}
		return false
	case []models.Subscription:
		for i := 0; i < len(entry); i++ {
			if entry[i].Domain == finder {
				return true
			}
		}
		return false
	}
	return false
}

func pushRelayActivityJob(sourceDomain string, body []byte) {
	activityID := uuid.NewV4()
	remainCount := len(relayState.Subscriptions) - 1

	if remainCount < 1 {
		return
	}

	evalScript := "redis.call('HSET',KEYS[1], 'body', ARGV[1], 'remain_count', ARGV[2]); redis.call('EXPIRE', KEYS[1], ARGV[3]);"
	relayState.RedisClient.Eval(evalScript, []string{"relay:activity:" + activityID.String()}, body, remainCount, 2*60).Result()

	for _, domain := range relayState.Subscriptions {
		if sourceDomain != domain.Domain {
			job := &tasks.Signature{
				Name:       "relay-v2",
				RetryCount: 0,
				Args: []tasks.Arg{
					{
						Name:  "inboxURL",
						Type:  "string",
						Value: domain.InboxURL,
					},
					{
						Name:  "activityID",
						Type:  "string",
						Value: activityID,
					},
				},
			}
			_, err := machineryServer.SendTask(job)
			if err != nil {
				logrus.Error(err)
			}
		}
	}
}

func pushRegisterJob(inboxURL string, body []byte) {
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
	_, err := machineryServer.SendTask(job)
	if err != nil {
		logrus.Error(err)
	}
}

func followAcceptable(activity *models.Activity, actor *models.Actor) error {
	if contains(activity.Object, "https://www.w3.org/ns/activitystreams#Public") {
		return nil
	} else {
		return errors.New("Follow only allowed for https://www.w3.org/ns/activitystreams#Public")
	}
}

func unFollowAcceptable(activity *models.Activity, actor *models.Actor) error {
	if contains(activity.Object, "https://www.w3.org/ns/activitystreams#Public") {
		return nil
	} else {
		return errors.New("Unfollow only allowed for https://www.w3.org/ns/activitystreams#Public")
	}
}

func suitableFollow(activity *models.Activity, actor *models.Actor) bool {
	domain, _ := url.Parse(activity.Actor)
	if contains(relayState.BlockedDomains, domain.Host) {
		return false
	}
	return true
}

func relayAcceptable(activity *models.Activity, actor *models.Actor) error {
	if !contains(activity.To, "https://www.w3.org/ns/activitystreams#Public") && !contains(activity.Cc, "https://www.w3.org/ns/activitystreams#Public") {
		return errors.New("activity should contain https://www.w3.org/ns/activitystreams#Public as receiver")
	}
	domain, _ := url.Parse(activity.Actor)
	if contains(relayState.Subscriptions, domain.Host) {
		return nil
	}
	return errors.New("to use the relay service, Subscribe me in advance")
}

func suitableRelay(activity *models.Activity, actor *models.Actor) bool {
	domain, _ := url.Parse(activity.Actor)
	if contains(relayState.LimitedDomains, domain.Host) {
		return false
	}
	if relayState.RelayConfig.BlockService && actor.Type != "Person" {
		return false
	}
	return true
}

func handleInbox(writer http.ResponseWriter, request *http.Request, activityDecoder func(*http.Request) (*models.Activity, *models.Actor, []byte, error)) {
	switch request.Method {
	case "POST":
		activity, actor, body, err := activityDecoder(request)
		if err != nil {
			writer.WriteHeader(400)
			writer.Write(nil)
		} else {
			domain, _ := url.Parse(activity.Actor)
			switch activity.Type {
			case "Follow":
				err = followAcceptable(activity, actor)
				if err != nil {
					resp := activity.GenerateResponse(globalConfig.ServerHostname(), "Reject")
					jsonData, _ := json.Marshal(&resp)
					go pushRegisterJob(actor.Inbox, jsonData)
					logrus.Error("Reject Follow Request : ", err.Error(), activity.Actor)

					writer.WriteHeader(202)
					writer.Write(nil)
				} else {
					if suitableFollow(activity, actor) {
						if relayState.RelayConfig.ManuallyAccept {
							relayState.RedisClient.HMSet("relay:pending:"+domain.Host, map[string]interface{}{
								"inbox_url":   actor.Endpoints.SharedInbox,
								"activity_id": activity.ID,
								"type":        "Follow",
								"actor":       actor.ID,
								"object":      activity.Object.(string),
							})
							logrus.Info("Pending Follow Request : ", activity.Actor)
						} else {
							resp := activity.GenerateResponse(globalConfig.ServerHostname(), "Accept")
							jsonData, _ := json.Marshal(&resp)
							go pushRegisterJob(actor.Inbox, jsonData)
							relayState.AddSubscription(models.Subscription{
								Domain:     domain.Host,
								InboxURL:   actor.Endpoints.SharedInbox,
								ActivityID: activity.ID,
								ActorID:    actor.ID,
							})
							logrus.Info("Accept Follow Request : ", activity.Actor)
						}
					} else {
						resp := activity.GenerateResponse(globalConfig.ServerHostname(), "Reject")
						jsonData, _ := json.Marshal(&resp)
						go pushRegisterJob(actor.Inbox, jsonData)
						logrus.Info("Reject Follow Request : ", activity.Actor)
					}

					writer.WriteHeader(202)
					writer.Write(nil)
				}
			case "Undo":
				nestedActivity, _ := activity.NestedActivity()
				if nestedActivity.Type == "Follow" && nestedActivity.Actor == activity.Actor {
					err = unFollowAcceptable(nestedActivity, actor)
					if err != nil {
						logrus.Error("Reject Unfollow Request : ", err.Error())
						writer.WriteHeader(400)
						writer.Write([]byte(err.Error()))
					} else {
						relayState.DelSubscription(domain.Host)
						logrus.Info("Accept Unfollow Request : ", activity.Actor)

						writer.WriteHeader(202)
						writer.Write(nil)
					}
				} else {
					err = relayAcceptable(activity, actor)
					if err != nil {
						writer.WriteHeader(400)
						writer.Write([]byte(err.Error()))
					} else {
						domain, _ := url.Parse(activity.Actor)
						go pushRelayActivityJob(domain.Host, body)
						logrus.Debug("Accept Relay Status : ", activity.Actor)

						writer.WriteHeader(202)
						writer.Write(nil)
					}
				}
			case "Create", "Update", "Delete", "Announce", "Move":
				err = relayAcceptable(activity, actor)
				if err != nil {
					writer.WriteHeader(400)
					writer.Write([]byte(err.Error()))
				} else {
					if suitableRelay(activity, actor) {
						if relayState.RelayConfig.CreateAsAnnounce && activity.Type == "Create" {
							nestedObject, err := activity.NestedActivity()
							if err != nil {
								logrus.Error("Fail Decode Activity : ", err.Error())
							}
							switch nestedObject.Type {
							case "Note":
								resp := nestedObject.GenerateAnnounce(globalConfig.ServerHostname())
								jsonData, _ := json.Marshal(&resp)
								go pushRelayActivityJob(domain.Host, jsonData)
								logrus.Debug("Accept Announce Note : ", activity.Actor)
							default:
								logrus.Debug("Skipping Announce", nestedObject.Type, ": ", activity.Actor)
							}
						} else {
							go pushRelayActivityJob(domain.Host, body)
							logrus.Debug("Accept Relay Status : ", activity.Actor)
						}
					} else {
						logrus.Debug("Skipping Relay Status : ", activity.Actor)
					}

					writer.WriteHeader(202)
					writer.Write(nil)
				}
			}
		}
	default:
		writer.WriteHeader(404)
		writer.Write(nil)
	}
}
