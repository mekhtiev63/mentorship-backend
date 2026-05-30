package domain

// AccountStatus reflects the users.status column.
type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusDisabled AccountStatus = "disabled"
)

// AccountState is the effective login eligibility of an account.
type AccountState struct {
	Status  AccountStatus
	Deleted bool
}

// CanLogin reports whether the account may authenticate.
func (s AccountState) CanLogin() bool {
	return !s.Deleted && s.Status == AccountStatusActive
}
