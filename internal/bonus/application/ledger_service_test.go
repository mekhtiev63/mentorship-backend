package application_test

import (
	"context"
	"testing"

	"github.com/go-mentorship-platform/backend/internal/bonus/application"
	"github.com/go-mentorship-platform/backend/internal/bonus/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

type memAccounts struct {
	balances map[string]int64
}

func (m *memAccounts) Get(_ context.Context, id domain.UserID) (domain.BonusAccount, error) {
	return domain.BonusAccount{UserID: id, Balance: domain.BonusAmount(m.balances[string(id)])}, nil
}
func (m *memAccounts) GetForUpdate(_ context.Context, id domain.UserID) (domain.BonusAccount, error) {
	return m.Get(context.Background(), id)
}
func (m *memAccounts) EnsureAccount(_ context.Context, id domain.UserID) error {
	if m.balances == nil {
		m.balances = map[string]int64{}
	}
	if _, ok := m.balances[string(id)]; !ok {
		m.balances[string(id)] = 0
	}
	return nil
}
func (m *memAccounts) UpdateBalance(_ context.Context, id domain.UserID, bal domain.BonusAmount) error {
	m.balances[string(id)] = int64(bal)
	return nil
}

type memLedger struct {
	refs map[string]struct{}
}

func (m *memLedger) Append(_ context.Context, tx domain.BonusTransaction) error {
	if m.refs == nil {
		m.refs = map[string]struct{}{}
	}
	k := string(tx.UserID) + "/" + tx.Reference
	if _, ok := m.refs[k]; ok && tx.Reference != "" {
		return domain.ErrDuplicateOperation
	}
	m.refs[k] = struct{}{}
	return nil
}
func (m *memLedger) FindByIdempotencyKey(context.Context, string) (domain.BonusTransaction, bool, error) {
	return domain.BonusTransaction{}, false, nil
}
func (m *memLedger) FindByReference(_ context.Context, uid domain.UserID, ref string) (domain.BonusTransaction, bool, error) {
	_, ok := m.refs[string(uid)+"/"+ref]
	return domain.BonusTransaction{}, ok, nil
}
func (m *memLedger) ListByUser(context.Context, domain.UserID, pagination.Params) ([]domain.BonusTransaction, int64, error) {
	return nil, 0, nil
}
func (m *memLedger) SumConvertDiscountPercent(context.Context, domain.UserID) (int, error) {
	return 0, nil
}

type memTx struct{}

func (memTx) WithinTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type memEvents struct{}

func (memEvents) Record(context.Context, string, map[string]any) error { return nil }

func TestCreditAchievementIdempotent(t *testing.T) {
	accounts := &memAccounts{balances: map[string]int64{}}
	ledger := &memLedger{}
	svc := application.NewBonusLedgerService(memTx{}, accounts, ledger, memEvents{})
	uid := "00000000-0000-0000-0000-000000000001"
	c1, err := svc.CreditAchievement(context.Background(), uid, "first_material_view", "evt-1")
	if err != nil || !c1 {
		t.Fatalf("first credit: created=%v err=%v", c1, err)
	}
	c2, err := svc.CreditAchievement(context.Background(), uid, "first_material_view", "evt-1")
	if err != nil || c2 {
		t.Fatalf("second credit: created=%v err=%v", c2, err)
	}
}
