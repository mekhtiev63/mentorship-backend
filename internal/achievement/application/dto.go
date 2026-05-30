package application

// DefinitionDTO is catalog item.
type DefinitionDTO struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UserAchievementDTO is granted achievement with catalog meta.
type UserAchievementDTO struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	GrantedAt   string `json:"grantedAt"`
}
