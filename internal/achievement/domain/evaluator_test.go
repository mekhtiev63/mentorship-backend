package domain_test

import (
	"context"
	"testing"

	"github.com/go-mentorship-platform/backend/internal/achievement/domain"
)

type memStats struct {
	approved int
	views    int
}

func (m memStats) CountApprovedBlocks(context.Context, domain.UserID) (int, error) {
	return m.approved, nil
}
func (m memStats) CountMaterialViews(context.Context, domain.UserID) (int, error) {
	return m.views, nil
}

type memRoadmap struct{ published int }

func (m memRoadmap) CountPublishedBlocks(context.Context) (int, error) {
	return m.published, nil
}

func TestThresholdApprovedBlocks(t *testing.T) {
	reg := domain.NewRuleRegistry()
	def := domain.AchievementDefinition{
		Code: "blocks_approved_3",
		Rule: domain.AchievementRule{
			Type:   domain.RuleTypeThreshold,
			On:     domain.ProgressBlockApproved,
			Metric: "approved_blocks_count",
			Gte:    intPtr(3),
		},
	}
	ev, err := reg.EvaluatorFor(def.Rule)
	if err != nil {
		t.Fatal(err)
	}
	ok, err := ev.Evaluate(context.Background(), def, domain.EvaluationContext{
		StudentID: "00000000-0000-0000-0000-000000000001",
	}, memStats{approved: 3}, memRoadmap{})
	if err != nil || !ok {
		t.Fatalf("expected grant ok, got %v %v", ok, err)
	}
}

func TestProgramCompleted(t *testing.T) {
	reg := domain.NewRuleRegistry()
	def := domain.AchievementDefinition{
		Rule: domain.AchievementRule{Type: domain.RuleTypeAllPublishedBlocksApproved},
	}
	ev, _ := reg.EvaluatorFor(def.Rule)
	ok, err := ev.Evaluate(context.Background(), def, domain.EvaluationContext{
		StudentID: "00000000-0000-0000-0000-000000000001",
	}, memStats{approved: 5}, memRoadmap{published: 5})
	if err != nil || !ok {
		t.Fatalf("expected program complete, got %v %v", ok, err)
	}
}

func intPtr(n int) *int { return &n }
