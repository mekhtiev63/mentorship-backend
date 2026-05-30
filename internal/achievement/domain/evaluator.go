package domain

import "context"

// EvaluationContext carries data for rule evaluation.
type EvaluationContext struct {
	EventName string
	StudentID UserID
}

// RuleEvaluator decides whether to grant.
type RuleEvaluator interface {
	Evaluate(ctx context.Context, def AchievementDefinition, eval EvaluationContext, stats ProgressStatsPort, roadmap RoadmapStatsPort) (bool, error)
}

// RuleRegistry maps rule types to evaluators.
type RuleRegistry struct {
	evaluators map[RuleType]RuleEvaluator
}

// NewRuleRegistry creates registry with built-in evaluators.
func NewRuleRegistry() *RuleRegistry {
	r := &RuleRegistry{evaluators: make(map[RuleType]RuleEvaluator)}
	r.evaluators[RuleTypeThreshold] = thresholdEvaluator{}
	r.evaluators[RuleTypeFirstEvent] = firstEventEvaluator{}
	r.evaluators[RuleTypeAllPublishedBlocksApproved] = allPublishedApprovedEvaluator{}
	return r
}

// EvaluatorFor returns evaluator for rule type.
func (r *RuleRegistry) EvaluatorFor(rule AchievementRule) (RuleEvaluator, error) {
	e, ok := r.evaluators[rule.Type]
	if !ok {
		return nil, ErrInvalidRule
	}
	return e, nil
}

type thresholdEvaluator struct{}

func (thresholdEvaluator) Evaluate(ctx context.Context, def AchievementDefinition, eval EvaluationContext, stats ProgressStatsPort, _ RoadmapStatsPort) (bool, error) {
	if def.Rule.Gte == nil {
		return false, nil
	}
	switch def.Rule.Metric {
	case "approved_blocks_count":
		n, err := stats.CountApprovedBlocks(ctx, eval.StudentID)
		if err != nil {
			return false, err
		}
		return n >= *def.Rule.Gte, nil
	default:
		return false, nil
	}
}

type firstEventEvaluator struct{}

func (firstEventEvaluator) Evaluate(ctx context.Context, def AchievementDefinition, eval EvaluationContext, stats ProgressStatsPort, _ RoadmapStatsPort) (bool, error) {
	if def.Rule.Eq == nil {
		return false, nil
	}
	switch def.Rule.Metric {
	case "material_views_count":
		n, err := stats.CountMaterialViews(ctx, eval.StudentID)
		if err != nil {
			return false, err
		}
		return n == *def.Rule.Eq, nil
	default:
		return false, nil
	}
}

type allPublishedApprovedEvaluator struct{}

func (allPublishedApprovedEvaluator) Evaluate(ctx context.Context, _ AchievementDefinition, eval EvaluationContext, stats ProgressStatsPort, roadmap RoadmapStatsPort) (bool, error) {
	published, err := roadmap.CountPublishedBlocks(ctx)
	if err != nil {
		return false, err
	}
	if published == 0 {
		return false, nil
	}
	approved, err := stats.CountApprovedBlocks(ctx, eval.StudentID)
	if err != nil {
		return false, err
	}
	return approved >= published, nil
}
