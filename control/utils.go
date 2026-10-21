package control

import "slices"

import "github.com/yukimochi/Activity-Relay/models"

func contains(entries any, key string) bool {
	switch entry := entries.(type) {
	case string:
		return entry == key
	case []string:
		return slices.Contains(entry, key)
	case []models.Subscriber:
		for i := range entry {
			if entry[i].Domain == key {
				return true
			}
		}
		return false
	case []models.Follower:
		for i := range entry {
			if entry[i].Domain == key {
				return true
			}
		}
		return false
	}
	return false
}
