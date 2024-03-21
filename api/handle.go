package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/yukimochi/Activity-Relay/models"
)

func handleWebfinger(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "HEAD", "GET":
		queriedResource := request.URL.Query()["resource"]
		if len(queriedResource) == 0 {
			http.Error(writer, "400 Bad Request", 400)
			return
		}
		var queriedSubject *models.WebfingerResource
		for _, webfingerResource := range WebfingerResources {
			if queriedResource[0] == webfingerResource.Subject {
				queriedSubject = &webfingerResource
				break
			}
		}

		if queriedSubject == nil {
			http.Error(writer, "404 Not Found", 404)
			return
		}

		response, err := json.Marshal(queriedSubject)
		if err != nil {
			logrus.Fatal("Failed to marshal webfinger resource : ", err.Error())
			http.Error(writer, "500 Internal Server Error", 500)
			return
		}

		switch request.Method {
		case "HEAD":
			writer.Header().Add("Content-Type", "application/json")
			writer.Header().Add("Content-Length", strconv.Itoa(len(response)))
			writer.WriteHeader(200)
			_, err := writer.Write(nil)
			if err != nil {
				logrus.Fatal("Failed to write response : ", err.Error())
			}
			break
		case "GET":
			writer.Header().Add("Content-Type", "application/json")
			writer.WriteHeader(200)
			_, err := writer.Write(response)
			if err != nil {
				logrus.Fatal("Failed to write response : ", err.Error())
			}
			break
		}
		break
	default:
		http.Error(writer, "400 Bad Request", 400)
		return
	}
}

func handleNodeinfoLink(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "HEAD", "GET":
		response, err := json.Marshal(&Nodeinfo.NodeinfoLinks)
		if err != nil {
			logrus.Fatal("Failed to marshal nodeinfo links : ", err.Error())
			http.Error(writer, "500 Internal Server Error", 500)
			return
		}

		switch request.Method {
		case "HEAD":
			writer.Header().Add("Content-Type", "application/json")
			writer.Header().Add("Content-Length", strconv.Itoa(len(response)))
			writer.WriteHeader(200)
			_, err := writer.Write(nil)
			if err != nil {
				logrus.Fatal("Failed to write response : ", err.Error())
			}
			break
		case "GET":
			writer.Header().Add("Content-Type", "application/json")
			writer.WriteHeader(200)
			_, err := writer.Write(response)
			if err != nil {
				logrus.Fatal("Failed to write response : ", err.Error())
			}
			break
		}
		break
	default:
		http.Error(writer, "400 Bad Request", 400)
		return
	}
}

func handleNodeinfo(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "HEAD", "GET":
		userTotal := len(RelayState.Subscribers)
		Nodeinfo.Nodeinfo.Usage.Users.Total = userTotal
		Nodeinfo.Nodeinfo.Usage.Users.ActiveMonth = userTotal
		Nodeinfo.Nodeinfo.Usage.Users.ActiveHalfyear = userTotal

		response, err := json.Marshal(&Nodeinfo.Nodeinfo)
		if err != nil {
			logrus.Fatal("Failed to marshal nodeinfo : ", err.Error())
			http.Error(writer, "500 Internal Server Error", 500)
			return
		}

		switch request.Method {
		case "HEAD":
			writer.Header().Add("Content-Type", "application/json")
			writer.Header().Add("Content-Length", strconv.Itoa(len(response)))
			writer.WriteHeader(200)
			_, err := writer.Write(nil)
			if err != nil {
				logrus.Fatal("Failed to write response : ", err.Error())
			}
			break
		case "GET":
			writer.Header().Add("Content-Type", "application/json")
			writer.WriteHeader(200)
			_, err := writer.Write(response)
			if err != nil {
				logrus.Fatal("Failed to write response : ", err.Error())
			}
			break
		}
		break
	default:
		http.Error(writer, "400 Bad Request", 400)
		return
	}
}

func handleRelayActor(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "HEAD", "GET":
		response, err := json.Marshal(&RelayActor)
		if err != nil {
			logrus.Fatal("Failed to marshal relay actor : ", err.Error())
			http.Error(writer, "500 Internal Server Error", 500)
			return
		}

		switch request.Method {
		case "HEAD":
			writer.Header().Add("Content-Type", "application/activity+json")
			writer.Header().Add("Content-Length", strconv.Itoa(len(response)))
			writer.WriteHeader(200)
			_, err := writer.Write(nil)
			if err != nil {
				logrus.Fatal("Failed to write response : ", err.Error())
			}
			break
		case "GET":
			writer.Header().Add("Content-Type", "application/activity+json")
			writer.WriteHeader(200)
			_, err := writer.Write(response)
			if err != nil {
				logrus.Fatal("Failed to write response : ", err.Error())
			}
			break
		}
		break
	default:
		http.Error(writer, "400 Bad Request", 400)
		return
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
					innerActivity, err := activity.UnwrapInnerActivity()
					if err != nil {
						writer.WriteHeader(202)
						writer.Write(nil)

						return
					}
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
					innerActivity, err := activity.UnwrapInnerActivity()
					if err != nil {
						writer.WriteHeader(202)
						writer.Write(nil)

						return
					}
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
					innerActivity, err := activity.UnwrapInnerActivity()
					if err != nil {
						writer.WriteHeader(202)
						writer.Write(nil)

						return
					}
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
					innerActivity, err := activity.UnwrapInnerActivity()
					if err != nil {
						writer.WriteHeader(202)
						writer.Write(nil)

						return
					}
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
