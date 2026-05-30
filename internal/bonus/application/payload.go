package application

import "encoding/json"

type achievementGrantedPayload struct {
	UserID          string `json:"userId"`
	AchievementCode string `json:"achievementCode"`
	SourceEventID   string `json:"sourceEventId"`
}

func parseAchievementGranted(raw json.RawMessage) (achievementGrantedPayload, error) {
	var p achievementGrantedPayload
	err := json.Unmarshal(raw, &p)
	return p, err
}
