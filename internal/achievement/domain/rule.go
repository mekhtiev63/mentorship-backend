package domain

import "encoding/json"

// RuleType identifies evaluator strategy.
type RuleType string

const (
	RuleTypeThreshold                  RuleType = "threshold"
	RuleTypeFirstEvent                 RuleType = "first_event"
	RuleTypeAllPublishedBlocksApproved RuleType = "all_published_blocks_approved"
)

// AchievementRule is parsed rule_json.
type AchievementRule struct {
	Type   RuleType `json:"type"`
	On     string   `json:"on"`
	Metric string   `json:"metric,omitempty"`
	Gte    *int     `json:"gte,omitempty"`
	Eq     *int     `json:"eq,omitempty"`
}

// ParseRule unmarshals rule JSON.
func ParseRule(raw json.RawMessage) (AchievementRule, error) {
	if len(raw) == 0 {
		return AchievementRule{}, ErrInvalidRule
	}
	var rule AchievementRule
	if err := json.Unmarshal(raw, &rule); err != nil {
		return AchievementRule{}, ErrInvalidRule
	}
	if rule.On == "" {
		return AchievementRule{}, ErrInvalidRule
	}
	return rule, nil
}

// MatchesEvent reports whether rule listens to event name.
func (r AchievementRule) MatchesEvent(eventName string) bool {
	return r.On == eventName
}
