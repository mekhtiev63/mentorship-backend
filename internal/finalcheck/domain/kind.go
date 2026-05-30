package domain

// CheckKind is tech or roast track.
type CheckKind string

const (
	CheckTech  CheckKind = "tech"
	CheckRoast CheckKind = "roast"
)

// ParseCheckKind parses track name.
func ParseCheckKind(raw string) (CheckKind, error) {
	switch CheckKind(raw) {
	case CheckTech, CheckRoast:
		return CheckKind(raw), nil
	default:
		return "", ErrInvalidTransition
	}
}
