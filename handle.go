package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unsafe"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/yukimochi/Activity-Relay/ActivityPub"
)

func handleWebfinger(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query()["resource"]
	if r.Method != "GET" || len(resource) == 0 {
		w.WriteHeader(400)
		w.Write(nil)
	} else {
		request := resource[0]
		if request == WebfingerResource.Subject {
			wfresource, err := json.Marshal(&WebfingerResource)
			if err != nil {
				panic(err)
			}
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write(wfresource)
		} else {
			w.WriteHeader(404)
			w.Write(nil)
		}
	}
}

func handleActor(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		actor, err := json.Marshal(&Actor)
		if err != nil {
			panic(err)
		}
		w.Header().Add("Content-Type", "application/activity+json")
		w.WriteHeader(200)
		w.Write(actor)
	} else {
		w.WriteHeader(400)
		w.Write(nil)
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
	}
	return false
}

func pushRelayJob(sourceInbox string, body []byte) {
	domains, _ := redClient.Keys("relay:subscription:*").Result()
	for _, domain := range domains {
		if sourceInbox != strings.Replace(domain, "relay:subscription:", "", 1) {
			inboxURL, _ := redClient.HGet(domain, "inbox_url").Result()
			job := &tasks.Signature{
				Name:       "relay",
				RetryCount: 0,
				Args: []tasks.Arg{
					{
						Name:  "inboxURL",
						Type:  "string",
						Value: inboxURL,
					},
					{
						Name:  "body",
						Type:  "string",
						Value: *(*string)(unsafe.Pointer(&body)),
					},
				},
			}
			_, err := macServer.SendTask(job)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func pushRegistorJob(inboxURL string, body []byte) {
	job := &tasks.Signature{
		Name:       "registor",
		RetryCount: 25,
		Args: []tasks.Arg{
			{
				Name:  "inboxURL",
				Type:  "string",
				Value: inboxURL,
			},
			{
				Name:  "body",
				Type:  "string",
				Value: *(*string)(unsafe.Pointer(&body)),
			},
		},
	}
	_, err := macServer.SendTask(job)
	if err != nil {
		fmt.Println(err)
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
	blocked, _ := redClient.HExists("relay:config:blockedDomain", domain.Host).Result()
	if blocked {
		return false
	}
	return true
}

func relayAcceptable(activity *activitypub.Activity, actor *activitypub.Actor) error {
	if !contains(activity.To, "https://www.w3.org/ns/activitystreams#Public") && !contains(activity.Cc, "https://www.w3.org/ns/activitystreams#Public") {
		return errors.New("Activity should contain https://www.w3.org/ns/activitystreams#Public as receiver")
	}
	domain, _ := url.Parse(activity.Actor)
	exist, _ := redClient.Exists("relay:subscription:" + domain.Host).Result()
	if exist == 0 {
		return errors.New("To use the relay service, Subscribe me in advance")
	}
	return nil
}

func suitableRelay(activity *activitypub.Activity, actor *activitypub.Actor) bool {
	domain, _ := url.Parse(activity.Actor)
	limited, _ := redClient.HExists("relay:config:limitedDomain", domain.Host).Result()
	if limited {
		return false
	}
	if relConfig.BlockService && actor.Type == "Service" {
		return false
	}
	return true
}

func handleInbox(w http.ResponseWriter, r *http.Request, activityDecoder func(*http.Request) (*activitypub.Activity, *activitypub.Actor, []byte, error)) {
	switch r.Method {
	case "POST":
		activity, actor, body, err := activityDecoder(r)
		if err != nil {
			w.WriteHeader(400)
			w.Write(nil)
		} else {
			domain, _ := url.Parse(activity.Actor)
			switch activity.Type {
			case "Follow":
				err = followAcceptable(activity, actor)
				if err != nil {
					resp := activitypub.GenerateActivityResponse(hostname, domain, "Reject", *activity)
					jsonData, _ := json.Marshal(&resp)
					go pushRegistorJob(actor.Inbox, jsonData)
					fmt.Println("Reject Follow Request : ", err.Error(), activity.Actor)

					w.WriteHeader(202)
					w.Write(nil)
				} else {
					if suitableFollow(activity, actor) {
						if relConfig.ManuallyAccept {
							redClient.HMSet("relay:pending:"+domain.Host, map[string]interface{}{
								"inbox_url":   actor.Endpoints.SharedInbox,
								"activity_id": activity.ID,
								"type":        "Follow",
								"actor":       actor.ID,
								"object":      activity.Object.(string),
							})
							fmt.Println("Pending Follow Request : ", activity.Actor)
						} else {
							resp := activitypub.GenerateActivityResponse(hostname, domain, "Accept", *activity)
							jsonData, _ := json.Marshal(&resp)
							go pushRegistorJob(actor.Inbox, jsonData)
							redClient.HSet("relay:subscription:"+domain.Host, "inbox_url", actor.Endpoints.SharedInbox)
							fmt.Println("Accept Follow Request : ", activity.Actor)
						}
					} else {
						resp := activitypub.GenerateActivityResponse(hostname, domain, "Reject", *activity)
						jsonData, _ := json.Marshal(&resp)
						go pushRegistorJob(actor.Inbox, jsonData)
						fmt.Println("Reject Follow Request : ", activity.Actor)
					}

					w.WriteHeader(202)
					w.Write(nil)
				}
			case "Undo":
				nestedActivity, _ := activitypub.DescribeNestedActivity(activity.Object)
				if nestedActivity.Type == "Follow" && nestedActivity.Actor == activity.Actor {
					err = unFollowAcceptable(nestedActivity, actor)
					if err != nil {
						fmt.Println("Reject Unfollow Request : ", err.Error())
						w.WriteHeader(400)
						w.Write([]byte(err.Error()))
					} else {
						redClient.Del("relay:subscription:" + domain.Host)
						fmt.Println("Accept Unfollow Request : ", activity.Actor)

						w.WriteHeader(202)
						w.Write(nil)
					}
				} else {
					err = relayAcceptable(activity, actor)
					if err != nil {
						w.WriteHeader(400)
						w.Write([]byte(err.Error()))
					} else {
						domain, _ := url.Parse(activity.Actor)
						go pushRelayJob(domain.Host, body)
						fmt.Println("Accept Relay Status : ", activity.Actor)

						w.WriteHeader(202)
						w.Write(nil)
					}
				}
			case "Create", "Update", "Delete", "Announce":
				err = relayAcceptable(activity, actor)
				if err != nil {
					w.WriteHeader(400)
					w.Write([]byte(err.Error()))
				} else {
					if suitableRelay(activity, actor) {
						if relConfig.CreateAsAnnounce && activity.Type == "Create" {
							nestedObject, err := activitypub.DescribeNestedActivity(activity.Object)
							if err != nil {
								fmt.Println("Fail Assert activity : activity.Actor")
							}
							switch nestedObject.Type {
							case "Note":
								resp := activitypub.GenerateActivityAnnounce(hostname, domain, nestedObject.ID)
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

					w.WriteHeader(202)
					w.Write(nil)
				}
			}
		}
	default:
		w.WriteHeader(404)
		w.Write(nil)
	}
}
