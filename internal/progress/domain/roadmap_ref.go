package domain

// RoadmapBlockRef is a read-only snapshot for progress UI.
type RoadmapBlockRef struct {
	BlockID   BlockID
	SortOrder int
	Title     string
}
