package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-mentorship-platform/backend/internal/interview/application"
	"github.com/go-mentorship-platform/backend/internal/interview/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

type memRepo struct {
	items map[domain.InterviewID]domain.Interview
}

func newMemRepo() *memRepo {
	return &memRepo{items: map[domain.InterviewID]domain.Interview{}}
}

func (m *memRepo) Insert(_ context.Context, i domain.Interview) error {
	m.items[i.ID] = i
	return nil
}

func (m *memRepo) GetByID(_ context.Context, id domain.InterviewID) (domain.Interview, error) {
	i, ok := m.items[id]
	if !ok {
		return domain.Interview{}, domain.ErrNotFound
	}
	return i, nil
}

func (m *memRepo) GetForUpdate(ctx context.Context, id domain.InterviewID) (domain.Interview, error) {
	return m.GetByID(ctx, id)
}

func (m *memRepo) Save(_ context.Context, i domain.Interview, expected domain.InterviewStatus) error {
	cur, ok := m.items[i.ID]
	if !ok || cur.Status != expected {
		return domain.ErrInvalidTransition
	}
	m.items[i.ID] = i
	return nil
}

func (m *memRepo) ListByStudent(context.Context, domain.StudentID, domain.InterviewKind, *domain.InterviewStatus, pagination.Params) ([]domain.Interview, int64, error) {
	return nil, 0, nil
}

func (m *memRepo) ListByInterviewer(context.Context, domain.UserID, domain.InterviewKind, *domain.InterviewStatus, pagination.Params) ([]domain.Interview, int64, error) {
	return nil, 0, nil
}

func (m *memRepo) ListCatalog(context.Context, pagination.Params, *string, *domain.InterviewOutcome) ([]domain.Interview, int64, error) {
	return nil, 0, nil
}

type memBuddy struct{ ok bool }

func (m memBuddy) IsActiveBuddyOf(context.Context, domain.UserID, domain.StudentID) (bool, error) {
	return m.ok, nil
}

type memTx struct{}

func (memTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type memEvents struct{}

func (memEvents) Record(context.Context, string, map[string]any) error { return nil }

func TestMockCreateRequiresAssignment(t *testing.T) {
	repo := newMemRepo()
	svc := application.NewMockInterviewService(repo, memBuddy{ok: false}, memEvents{})
	_, err := svc.Create(context.Background(), "33333333-3333-3333-3333-333333333333", application.MockCreateInput{
		StudentID: "22222222-2222-2222-2222-222222222222", ScheduledAt: time.Now().UTC(),
	})
	if err != domain.ErrForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestFeedbackCompleteMock(t *testing.T) {
	repo := newMemRepo()
	id := domain.InterviewID("11111111-1111-1111-1111-111111111111")
	buddy := domain.UserID("33333333-3333-3333-3333-333333333333")
	now := time.Now().UTC()
	mock, _ := domain.NewMockInterview(id, domain.StudentID("22222222-2222-2222-2222-222222222222"), buddy, now, "", now)
	repo.items[id] = mock

	fb := application.NewInterviewFeedbackService(repo, memBuddy{ok: true}, memTx{}, memEvents{})
	dto, err := fb.CompleteMock(context.Background(), string(buddy), string(id), application.FeedbackInput{
		Feedback: "well done", Outcome: domain.OutcomeNoResult,
	})
	if err != nil {
		t.Fatal(err)
	}
	if dto.Status != string(domain.StatusCompleted) || dto.Feedback == nil {
		t.Fatalf("unexpected dto %+v", dto)
	}
	if repo.items[id].CatalogPublished {
		t.Fatal("mock must not be in catalog")
	}
}

func TestRealCompletePublishesCatalog(t *testing.T) {
	repo := newMemRepo()
	id := domain.InterviewID("11111111-1111-1111-1111-111111111111")
	student := domain.StudentID("22222222-2222-2222-2222-222222222222")
	now := time.Now().UTC()
	real, _ := domain.NewRealInterview(id, student, "Co", "Dev", now, "", nil, now)
	repo.items[id] = real

	svc := application.NewRealInterviewService(repo, memTx{}, memEvents{})
	dto, err := svc.Complete(context.Background(), string(student), string(id), domain.OutcomeOffer)
	if err != nil {
		t.Fatal(err)
	}
	if !dto.CatalogPublished {
		t.Fatal("real should publish to catalog")
	}
}
