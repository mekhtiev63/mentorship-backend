package domain

// DiscountState is active discount summary for a user.
type DiscountState struct {
	ActiveDiscountPercent      int
	RemainingDiscountHeadroom  int
}

// ApplyConvertHeadroom checks whether added discount fits under 15%.
func ApplyConvertHeadroom(active int, addedPercent int) error {
	if addedPercent <= 0 {
		return ErrInvalidAmount
	}
	if active+addedPercent > MaxDiscountPercent {
		return ErrDiscountLimit
	}
	return nil
}

// NewDiscountState builds state from active percent.
func NewDiscountState(active int) DiscountState {
	if active > MaxDiscountPercent {
		active = MaxDiscountPercent
	}
	return DiscountState{
		ActiveDiscountPercent:     active,
		RemainingDiscountHeadroom: MaxDiscountPercent - active,
	}
}
