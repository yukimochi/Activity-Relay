package control

import "github.com/yukimochi/Activity-Relay/models"

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
