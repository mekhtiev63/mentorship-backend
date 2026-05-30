package domain

// BonusAmount is non-negative bonus points.
type BonusAmount int64

// MaxDiscountPercent is the discount ceiling.
const MaxDiscountPercent = 15

// PointsPerDiscountPercent is bonus points per 1% discount on convert.
const PointsPerDiscountPercent = 100

// ParseBonusAmount validates positive amount for operations.
func ParseBonusAmount(n int64) (BonusAmount, error) {
	if n <= 0 {
		return 0, ErrInvalidAmount
	}
	return BonusAmount(n), nil
}

// DiscountPercentFromPoints computes discount % from points (floor).
func DiscountPercentFromPoints(points BonusAmount) int {
	return int(points) / PointsPerDiscountPercent
}
