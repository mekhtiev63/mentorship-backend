package domain

// AchievementCode is achievement identifier for bonus policy.
type AchievementCode string

// CreditForAchievement returns bonus points for achievement code.
func CreditForAchievement(code AchievementCode) (BonusAmount, bool) {
	switch code {
	case "first_material_view":
		return 10, true
	case "first_block_approved":
		return 50, true
	case "blocks_approved_3":
		return 100, true
	case "blocks_approved_5":
		return 200, true
	case "program_completed":
		return 500, true
	default:
		return 0, false
	}
}

// AchievementCreditReference builds idempotent reference key.
func AchievementCreditReference(code AchievementCode, sourceEventID string) string {
	return "achievement:" + string(code) + ":" + sourceEventID
}
