package application_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-mentorship-platform/backend/internal/achievement/application"
	"github.com/go-mentorship-platform/backend/internal/achievement/domain"
)

type memDefs struct {
	byEvent []domain.AchievementDefinition
}

func (m *memDefs) ListActive(context.Context) ([]domain.AchievementDefinition, error) {
	return m.byEvent, nil
}
func (m *memDefs) ListActiveByEvent(_ context.Context, eventName string) ([]domain.AchievementDefinition, error) {
	var out []domain.AchievementDefinition
	for _, d := range m.byEvent {
		if d.Rule.On == eventName {
			out = append(out, d)
		}
	}
	return out, nil
}

type memGrants struct {
	keys map[string]struct{}
}

func (m *memGrants) Exists(_ context.Context, userID domain.UserID, code domain.AchievementCode) (bool, error) {
	_, ok := m.keys[string(userID)+"/"+string(code)]
	return ok, nil
}
func (m *memGrants) Grant(_ context.Context, a domain.UserAchievement) (bool, error) {
	k := string(a.UserID) + "/" + string(a.AchievementCode)
	if _, ok := m.keys[k]; ok {
		return false, nil
	}
	m.keys[k] = struct{}{}
	return true, nil
}
func (m *memGrants) ListByUser(context.Context, domain.UserID) ([]domain.UserAchievement, error) {
	return nil, nil
}

type memTx struct{}

func (memTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type memStats struct{ approved, views int }

func (m memStats) CountApprovedBlocks(context.Context, domain.UserID) (int, error) {
	return m.approved, nil
}
func (m memStats) CountMaterialViews(context.Context, domain.UserID) (int, error) {
	return m.views, nil
}

type memRoadmap struct{ n int }

func (m memRoadmap) CountPublishedBlocks(context.Context) (int, error) { return m.n, nil }

type memEvents struct{}

func (memEvents) Record(context.Context, string, map[string]any) error { return nil }

func TestGrantFirstMaterialView(t *testing.T) {
	gte := 1
	defs := &memDefs{byEvent: []domain.AchievementDefinition{{
		Code: "first_material_view",
		Rule: domain.AchievementRule{
			Type: domain.RuleTypeFirstEvent, On: domain.ProgressMaterialViewed,
			Metric: "material_views_count", Eq: &gte,
		},
	}}}
	grants := &memGrants{keys: map[string]struct{}{}}
	svc := application.NewAchievementGrantService(memTx{}, defs, grants, memStats{views: 1}, memRoadmap{}, memEvents{})
	payload, _ := json.Marshal(map[string]string{"studentId": "00000000-0000-0000-0000-000000000001"})
	err := svc.HandleProgressEvent(context.Background(), "00000000-0000-0000-0000-000000000099", domain.ProgressMaterialViewed, payload)
	if err != nil {
		t.Fatal(err)
	}
	if len(grants.keys) != 1 {
		t.Fatalf("expected 1 grant, got %d", len(grants.keys))
	}
}
