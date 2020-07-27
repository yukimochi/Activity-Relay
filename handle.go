package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/RichardKnop/machinery/v1/tasks"
	activitypub "github.com/yukimochi/Activity-Relay/ActivityPub"
	state "github.com/yukimochi/Activity-Relay/State"
)

func handleWebfinger(writer http.ResponseWriter, request *http.Request) {
	resource := request.URL.Query()["resource"]
	if request.Method != "GET" || len(resource) == 0 {
		writer.WriteHeader(400)
		writer.Write(nil)
	} else {
		request := resource[0]
		if request == WebfingerResource.Subject {
			wfresource, err := json.Marshal(&WebfingerResource)
			if err != nil {
				panic(err)
			}
			writer.Header().Add("Content-Type", "application/json")
			writer.WriteHeader(200)
			writer.Write(wfresource)
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
		linksresource, err := json.Marshal(&Nodeinfo.NodeinfoLinks)
		if err != nil {
			panic(err)
		}
		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
		writer.Write(linksresource)
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
		linksresource, err := json.Marshal(&Nodeinfo.Nodeinfo)
		if err != nil {
			panic(err)
		}
		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
		writer.Write(linksresource)
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
	case []state.Subscription:
		for i := 0; i < len(entry); i++ {
			if entry[i].Domain == finder {
				return true
			}
		}
		return false
	}
	return false
}

func pushRelayJob(sourceInbox string, body []byte) {
	for _, domain := range relayState.Subscriptions {
		if sourceInbox != domain.Domain {
			job := &tasks.Signature{
				Name:       "relay",
				RetryCount: 0,
				Args: []tasks.Arg{
					{
						Name:  "inboxURL",
						Type:  "string",
						Value: domain.InboxURL,
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
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
}

func pushRegistorJob(inboxURL string, body []byte) {
	job := &tasks.Signature{
		Name:       "registor",
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
		fmt.Fprintln(os.Stderr, err)
	}
}

func followAcceptable(activity *activitypub.Activity, actor *activitypub.Actor) error {
	if contains(activity.Object, "https://www.w3.org/ns/activitystreams#Public") {
		return nil
	} else {
		return errors.New("Follow only allowed for https://www.w3.org/ns/activitystreams#Public")
	}
}

func unFollowAcceptable(activity *activitypub.Activity, actor *activitypub.Actor) error {
	if contains(activity.Object, "https://www.w3.org/ns/activitystreams#Public") {
		return nil
	} else {
		return errors.New("Unfollow only allowed for https://www.w3.org/ns/activitystreams#Public")
	}
}

func suitableFollow(activity *activitypub.Activity, actor *activitypub.Actor) bool {
	domain, _ := url.Parse(activity.Actor)
	if contains(relayState.BlockedDomains, domain.Host) {
		return false
	}
	return true
}

func relayAcceptable(activity *activitypub.Activity, actor *activitypub.Actor) error {
	if !contains(activity.To, "https://www.w3.org/ns/activitystreams#Public") && !contains(activity.Cc, "https://www.w3.org/ns/activitystreams#Public") {
		return errors.New("Activity should contain https://www.w3.org/ns/activitystreams#Public as receiver")
	}
	domain, _ := url.Parse(activity.Actor)
	if contains(relayState.Subscriptions, domain.Host) {
		return nil
	}
	return errors.New("To use the relay service, Subscribe me in advance")
}

func suitableRelay(activity *activitypub.Activity, actor *activitypub.Actor) bool {
	domain, _ := url.Parse(activity.Actor)
	if contains(relayState.LimitedDomains, domain.Host) {
		return false
	}
	if relayState.RelayConfig.BlockService && actor.Type != "Person" {
		return false
	}
	return true
}

func handleInbox(writer http.ResponseWriter, request *http.Request, activityDecoder func(*http.Request) (*activitypub.Activity, *activitypub.Actor, []byte, error)) {
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
					resp := activity.GenerateResponse(hostURL, "Reject")
					jsonData, _ := json.Marshal(&resp)
					go pushRegistorJob(actor.Inbox, jsonData)
					fmt.Println("Reject Follow Request : ", err.Error(), activity.Actor)

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
							fmt.Println("Pending Follow Request : ", activity.Actor)
						} else {
							resp := activity.GenerateResponse(hostURL, "Accept")
							jsonData, _ := json.Marshal(&resp)
							go pushRegistorJob(actor.Inbox, jsonData)
							relayState.AddSubscription(state.Subscription{
								Domain:     domain.Host,
								InboxURL:   actor.Endpoints.SharedInbox,
								ActivityID: activity.ID,
								ActorID:    actor.ID,
							})
							fmt.Println("Accept Follow Request : ", activity.Actor)
						}
					} else {
						resp := activity.GenerateResponse(hostURL, "Reject")
						jsonData, _ := json.Marshal(&resp)
						go pushRegistorJob(actor.Inbox, jsonData)
						fmt.Println("Reject Follow Request : ", activity.Actor)
					}

					writer.WriteHeader(202)
					writer.Write(nil)
				}
			case "Undo":
				nestedActivity, _ := activity.NestedActivity()
				if nestedActivity.Type == "Follow" && nestedActivity.Actor == activity.Actor {
					err = unFollowAcceptable(nestedActivity, actor)
					if err != nil {
						fmt.Println("Reject Unfollow Request : ", err.Error())
						writer.WriteHeader(400)
						writer.Write([]byte(err.Error()))
					} else {
						relayState.DelSubscription(domain.Host)
						fmt.Println("Accept Unfollow Request : ", activity.Actor)

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
						go pushRelayJob(domain.Host, body)
						fmt.Println("Accept Relay Status : ", activity.Actor)

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
								fmt.Println("Fail Assert activity : activity.Actor")
							}
							switch nestedObject.Type {
							case "Note":
								resp := nestedObject.GenerateAnnounce(hostURL)
								jsonData, _ := json.Marshal(&resp)
								go pushRelayJob(domain.Host, jsonData)
								fmt.Println("Accept Announce Note : ", activity.Actor)
							default:
								fmt.Println("Skipping Announce", nestedObject.Type, ": ", activity.Actor)
							}
						} else {
							go pushRelayJob(domain.Host, body)
							fmt.Println("Accept Relay Status : ", activity.Actor)
						}
					} else {
						fmt.Println("Skipping Relay Status : ", activity.Actor)
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
