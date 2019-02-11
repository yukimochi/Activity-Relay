package main

import state "github.com/yukimochi/Activity-Relay/State"

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
