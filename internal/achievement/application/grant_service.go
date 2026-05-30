package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-mentorship-platform/backend/internal/achievement/domain"
)

// AchievementGrantService grants achievements from domain events.
type AchievementGrantService struct {
	tx          domain.Transactor
	definitions domain.AchievementDefinitionRepository
	grants      domain.UserAchievementRepository
	progress    domain.ProgressStatsPort
	roadmap     domain.RoadmapStatsPort
	registry    *domain.RuleRegistry
	events      domain.EventRecorder
}

// NewAchievementGrantService builds AchievementGrantService.
func NewAchievementGrantService(
	tx domain.Transactor,
	definitions domain.AchievementDefinitionRepository,
	grants domain.UserAchievementRepository,
	progress domain.ProgressStatsPort,
	roadmap domain.RoadmapStatsPort,
	events domain.EventRecorder,
) *AchievementGrantService {
	return &AchievementGrantService{
		tx:          tx,
		definitions: definitions,
		grants:      grants,
		progress:    progress,
		roadmap:     roadmap,
		registry:    domain.NewRuleRegistry(),
		events:      events,
	}
}

// HandleProgressEvent evaluates rules and grants inside a transaction.
func (s *AchievementGrantService) HandleProgressEvent(ctx context.Context, sourceEventID, eventName string, payload json.RawMessage) error {
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		return s.processEvent(ctx, sourceEventID, eventName, payload)
	})
}

func (s *AchievementGrantService) processEvent(ctx context.Context, sourceEventID, eventName string, payload json.RawMessage) error {
	studentRaw, err := studentIDFromPayload(payload)
	if err != nil || studentRaw == "" {
		return err
	}
	studentID, err := domain.ParseUserID(studentRaw)
	if err != nil {
		return err
	}
	srcID, err := domain.ParseSourceEventID(sourceEventID)
	if err != nil {
		return err
	}

	defs, err := s.definitions.ListActiveByEvent(ctx, eventName)
	if err != nil {
		return err
	}
	evalCtx := domain.EvaluationContext{EventName: eventName, StudentID: studentID}
	for _, def := range defs {
			if !def.Rule.MatchesEvent(eventName) {
				continue
			}
			exists, err := s.grants.Exists(ctx, studentID, def.Code)
			if err != nil {
				return err
			}
			if exists {
				continue
			}
			evaluator, err := s.registry.EvaluatorFor(def.Rule)
			if err != nil {
				continue
			}
			ok, err := evaluator.Evaluate(ctx, def, evalCtx, s.progress, s.roadmap)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			created, err := s.grants.Grant(ctx, domain.UserAchievement{
				UserID:          studentID,
				AchievementCode: def.Code,
				GrantedAt:       time.Now().UTC(),
				SourceEventID:   srcID,
			})
			if err != nil {
				return err
			}
			if created {
				_ = s.events.Record(ctx, domain.EventGranted, map[string]any{
					"userId":           string(studentID),
					"achievementCode":  string(def.Code),
					"sourceEventId":    string(srcID),
					"triggerEventName": eventName,
				})
			}
	}
	return nil
}
