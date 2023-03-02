package api

import (
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/yukimochi/Activity-Relay/models"
	"net/http"
	"net/url"
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
		userTotal := len(RelayState.Subscribers)
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
				fallthrough
			case isToMyFollower(activity.To), isToMyFollower(activity.Cc):
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
				case "Accept":
					innerActivity, _ := activity.UnwrapInnerActivity()
					switch innerActivity.Type {
					case "Follow":
						finalizeMutuallyFollow(innerActivity, actor, activity.Type)
						writer.WriteHeader(202)
						writer.Write(nil)
					default:
						writer.WriteHeader(202)
						writer.Write(nil)
					}
				case "Reject":
					innerActivity, _ := activity.UnwrapInnerActivity()
					switch innerActivity.Type {
					case "Follow":
						finalizeMutuallyFollow(innerActivity, actor, activity.Type)
						writer.WriteHeader(202)
						writer.Write(nil)
					default:
						writer.WriteHeader(202)
						writer.Write(nil)
					}
				case "Announce":
					if !isActorSubscribersOrFollowers(actorID) {
						err = errors.New("to use the relay service, please follow in advance")
						writer.WriteHeader(401)
						writer.Write([]byte(err.Error()))

						return
					}
					switch innerObject := activity.Object.(type) {
					case string:
						origActivity, origActor, err := fetchOriginalActivityFromURL(innerObject)
						if err != nil {
							logrus.Debug("Failed Announce Activity : ", activity.Actor)
							writer.WriteHeader(400)
							writer.Write([]byte(err.Error()))

							return
						}
						executeAnnounceActivity(origActivity, origActor)
					default:
						logrus.Debug("Skipped Announce Activity : ", activity.Actor)
					}
					writer.WriteHeader(202)
					writer.Write(nil)
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
