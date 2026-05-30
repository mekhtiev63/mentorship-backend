package application_test

import (
	"context"
	"testing"

	"github.com/go-mentorship-platform/backend/internal/oneonone/application"
	"github.com/go-mentorship-platform/backend/internal/oneonone/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

type memRepo struct {
	byID map[domain.RequestID]domain.OneOnOneRequest
}

func newMemRepo() *memRepo {
	return &memRepo{byID: map[domain.RequestID]domain.OneOnOneRequest{}}
}

func (m *memRepo) Insert(_ context.Context, req domain.OneOnOneRequest) error {
	m.byID[req.ID] = req
	return nil
}

func (m *memRepo) GetByID(_ context.Context, id domain.RequestID) (domain.OneOnOneRequest, error) {
	req, ok := m.byID[id]
	if !ok {
		return domain.OneOnOneRequest{}, domain.ErrNotFound
	}
	return req, nil
}

func (m *memRepo) GetForUpdate(ctx context.Context, id domain.RequestID) (domain.OneOnOneRequest, error) {
	return m.GetByID(ctx, id)
}

func (m *memRepo) Save(_ context.Context, req domain.OneOnOneRequest, expected domain.RequestStatus) error {
	cur, ok := m.byID[req.ID]
	if !ok || cur.Status != expected {
		return domain.ErrInvalidTransition
	}
	m.byID[req.ID] = req
	return nil
}

func (m *memRepo) ListByStudent(context.Context, domain.StudentID, pagination.Params) ([]domain.OneOnOneRequest, int64, error) {
	return nil, 0, nil
}

func (m *memRepo) ListAll(context.Context, pagination.Params, *domain.RequestStatus) ([]domain.OneOnOneRequest, int64, error) {
	return nil, 0, nil
}

type memBuddy struct{ buddy domain.BuddyID }

func (m memBuddy) GetActiveBuddyID(context.Context, domain.StudentID) (domain.BuddyID, error) {
	return m.buddy, nil
}

type memBonus struct {
	balance int64
	debited map[domain.RequestID]struct{}
}

func (m *memBonus) HasSufficientBalance(_ context.Context, _ domain.StudentID, amount int64) (bool, error) {
	return m.balance >= amount, nil
}

func (m *memBonus) DebitForRequest(_ context.Context, _ domain.StudentID, requestID domain.RequestID, amount int64) error {
	if m.balance < amount {
		return domain.ErrInsufficientBonus
	}
	if _, ok := m.debited[requestID]; ok {
		return nil
	}
	m.balance -= amount
	m.debited[requestID] = struct{}{}
	return nil
}

type memTx struct{}

func (memTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type memEvents struct{}

func (memEvents) Record(context.Context, string, map[string]any) error { return nil }

func TestAdminApproveDebitsOnce(t *testing.T) {
	repo := newMemRepo()
	reqID := domain.RequestID("11111111-1111-1111-1111-111111111111")
	student := domain.StudentID("22222222-2222-2222-2222-222222222222")
	repo.byID[reqID] = domain.OneOnOneRequest{
		ID: reqID, StudentID: student, BuddyID: "33333333-3333-3333-3333-333333333333",
		Status: domain.StatusPending,
	}
	bonus := &memBonus{balance: 2000, debited: map[domain.RequestID]struct{}{}}
	admin := application.NewOneOnOneAdminService(repo, bonus, memTx{}, memEvents{})
	adminID := "44444444-4444-4444-4444-444444444444"

	if err := admin.ApproveRequest(context.Background(), adminID, string(reqID)); err != nil {
		t.Fatal(err)
	}
	if err := admin.ApproveRequest(context.Background(), adminID, string(reqID)); err != nil {
		t.Fatal(err)
	}
	if bonus.balance != 1000 {
		t.Fatalf("expected balance 1000, got %d", bonus.balance)
	}
	if repo.byID[reqID].Status != domain.StatusAccepted {
		t.Fatalf("expected accepted, got %s", repo.byID[reqID].Status)
	}
}

func TestCreateRequiresBalance(t *testing.T) {
	repo := newMemRepo()
	bonus := &memBonus{balance: 500, debited: map[domain.RequestID]struct{}{}}
	svc := application.NewOneOnOneService(repo, memBuddy{buddy: "b"}, bonus, memEvents{})
	_, err := svc.CreateRequest(context.Background(), "22222222-2222-2222-2222-222222222222", application.CreateRequestInput{
		Message: "hello",
	})
	if err != domain.ErrInsufficientBonus {
		t.Fatalf("expected insufficient bonus, got %v", err)
	}
}
