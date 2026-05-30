package application

import "encoding/json"

type progressPayload struct {
	StudentID string `json:"studentId"`
}

func studentIDFromPayload(raw json.RawMessage) (string, error) {
	var p progressPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return "", err
	}
	return p.StudentID, nil
}
