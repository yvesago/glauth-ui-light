package helpers

import (
	"encoding/json"
	"time"
)

/* Session management */

// Status : session Object.
type Status struct {
	Lock     bool   `json:"lock"`
	LastSeen int64  `json:"lastseen"`
	Count    int    `json:"count"`
	User     string `json:"user"`
	UserID   string `json:"userid"`
	//	Confirm  bool   `json:"confirm"`
}

// ToJSONStr Status object to string.
func (s *Status) ToJSONStr() string {
	b, _ := json.Marshal(s)
	return string(b)
}

// StrToStatus un serialize to Status object.
func StrToStatus(str string) Status {
	var r Status
	json.Unmarshal([]byte(str), &r) //nolint:errcheck // return empty status
	return r
}

// FailLimiter : Lock for timeLimit seconds after 4 attempts.
func FailLimiter(s Status, timeLimit int64) Status {
	// var timeLimit int64 = 30
	now := time.Now().UTC().Unix()
	ret := s
	ret.LastSeen = now
	if s.Lock {
		if now-s.LastSeen > timeLimit {
			ret.Lock = false
			ret.Count = 1
		}
	} else {
		ret.Count++
		if now-s.LastSeen > timeLimit {
			ret.Count = 1
		}
		if ret.Count > 3 {
			ret.Lock = true
			ret.Count = 0
		}
	}
	return ret
}
