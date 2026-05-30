package domain

// RelatedType links event to another entity.
type RelatedType string

const (
	RelatedOneOnOne RelatedType = "one_on_one"
	RelatedInterview RelatedType = "interview"
	RelatedOther    RelatedType = "other"
)

// ParseRelatedType parses related type.
func ParseRelatedType(raw string) (RelatedType, error) {
	switch RelatedType(raw) {
	case RelatedOneOnOne, RelatedInterview, RelatedOther:
		return RelatedType(raw), nil
	default:
		return RelatedOther, nil
	}
}
