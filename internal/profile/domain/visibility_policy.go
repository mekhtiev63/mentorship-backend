package domain

// RelationshipReader checks buddy assignment for visibility.
type RelationshipReader interface {
	IsAssignedBuddy(buddyID, studentID UserID) bool
}

// ViewerContext describes who is requesting a profile.
type ViewerContext struct {
	ViewerID   UserID
	IsAdmin    bool
	Relationship RelationshipReader
}

// CanView reports whether viewer may see owner's profile.
func CanView(viewer ViewerContext, ownerID UserID, visibility Visibility) bool {
	if viewer.ViewerID == ownerID {
		return true
	}
	if viewer.IsAdmin {
		return true
	}
	switch visibility {
	case VisibilityPublic:
		return true
	case VisibilityPrivate:
		return false
	case VisibilityBuddiesOnly:
		if viewer.Relationship == nil {
			return false
		}
		return viewer.Relationship.IsAssignedBuddy(viewer.ViewerID, ownerID)
	default:
		return false
	}
}
