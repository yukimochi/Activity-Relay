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
	if r.Method == "GET" && len(resource) == 0 {
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
	receivers, _ := RedClient.Keys("subscription:*").Result()
	for _, receiver := range receivers {
		if sourceInbox != strings.Replace(receiver, "subscription:", "", 1) {
			inboxURL, _ := RedClient.HGet(receiver, "inbox_url").Result()
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
			_, err := MacServer.SendTask(job)
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
	_, err := MacServer.SendTask(job)
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

func suitableFollow(activity *activitypub.Activity, actor *activitypub.Actor) bool {
	domain, _ := url.Parse(activity.Actor)
	receivers, _ := RedClient.Exists("blocked_domain:" + domain.Host).Result()
	if receivers != 0 {
		return false
	}
	return true
}

func relayAcceptable(activity *activitypub.Activity, actor *activitypub.Actor) error {
	if !contains(activity.To, "https://www.w3.org/ns/activitystreams#Public") && !contains(activity.Cc, "https://www.w3.org/ns/activitystreams#Public") {
		return errors.New("Activity should contain https://www.w3.org/ns/activitystreams#Public as receiver")
	}
	domain, _ := url.Parse(activity.Actor)
	receivers, _ := RedClient.Exists("subscription:" + domain.Host).Result()
	if receivers == 0 {
		return errors.New("To use the relay service, Subscribe me in advance")
	}
	return nil
}

func suitableRelay(activity *activitypub.Activity, actor *activitypub.Actor) bool {
	domain, _ := url.Parse(activity.Actor)
	receivers, _ := RedClient.Exists("limited_domain:" + domain.Host).Result()
	if receivers != 0 {
		return false
	}
	return true
}

func handleInbox(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		activity, actor, body, err := decodeActivity(r)
		if err != nil {
			w.WriteHeader(400)
			w.Write(nil)
		} else {
			switch activity.Type {
			case "Follow":
				err = followAcceptable(activity, actor)
				if err != nil {
					w.WriteHeader(400)
					w.Write([]byte(err.Error()))
				} else {
					domain, _ := url.Parse(activity.Actor)
					var responseType string
					if suitableFollow(activity, actor) {
						responseType = "Accept"
						RedClient.HSet("subscription:"+domain.Host, "inbox_url", actor.Endpoints.SharedInbox)
					} else {
						responseType = "Reject"
					}
					resp := activitypub.GenerateActivityResponse(Hostname, domain, responseType, *activity)
					jsonData, _ := json.Marshal(&resp)
					go pushRegistorJob(actor.Inbox, jsonData)

					fmt.Println(responseType+" Follow Request : ", activity.Actor)
					w.WriteHeader(202)
					w.Write(nil)
				}
			case "Undo":
				nestedActivity, _ := activitypub.DescribeNestedActivity(activity.Object)
				if nestedActivity.Type == "Follow" && nestedActivity.Actor == activity.Actor {
					domain, _ := url.Parse(activity.Actor)
					RedClient.Del("subscription:" + domain.Host)

					fmt.Println("Accept Unfollow Request : ", activity.Actor)
					w.WriteHeader(202)
					w.Write(nil)
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
						domain, _ := url.Parse(activity.Actor)
						go pushRelayJob(domain.Host, body)

						fmt.Println("Accept Relay Status : ", activity.Actor)
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
