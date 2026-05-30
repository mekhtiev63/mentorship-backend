package domain

// SequentialBlockPolicy enforces linear roadmap completion.
type SequentialBlockPolicy struct{}

// CanSubmit reports whether the student may submit the target block.
func (SequentialBlockPolicy) CanSubmit(
	target BlockID,
	ordered []RoadmapBlockRef,
	progressByBlock map[BlockID]BlockProgress,
) error {
	var targetOrder int
	found := false
	for _, b := range ordered {
		if b.BlockID == target {
			targetOrder = b.SortOrder
			found = true
			break
		}
	}
	if !found {
		return ErrBlockNotVisible
	}
	for _, b := range ordered {
		if b.SortOrder >= targetOrder {
			break
		}
		p, ok := progressByBlock[b.BlockID]
		if !ok || p.Status != StatusApproved {
			return ErrSequentialBlock
		}
	}
	return nil
}
