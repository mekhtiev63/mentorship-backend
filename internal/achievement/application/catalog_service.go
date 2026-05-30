package application

import (
	"context"
	"time"

	"github.com/go-mentorship-platform/backend/internal/achievement/domain"
)

// AchievementCatalogService serves achievement catalog and user grants read API.
type AchievementCatalogService struct {
	definitions domain.AchievementDefinitionRepository
	grants      domain.UserAchievementRepository
	buddy       domain.BuddyScopePort
}

// NewAchievementCatalogService builds AchievementCatalogService.
func NewAchievementCatalogService(
	definitions domain.AchievementDefinitionRepository,
	grants domain.UserAchievementRepository,
	buddy domain.BuddyScopePort,
) *AchievementCatalogService {
	return &AchievementCatalogService{
		definitions: definitions,
		grants:      grants,
		buddy:       buddy,
	}
}

// ListCatalog returns active achievement definitions.
func (s *AchievementCatalogService) ListCatalog(ctx context.Context) ([]DefinitionDTO, error) {
	defs, err := s.definitions.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]DefinitionDTO, len(defs))
	for i, d := range defs {
		out[i] = DefinitionDTO{
			Code:        string(d.Code),
			Title:       d.Title,
			Description: d.Description,
		}
	}
	return out, nil
}

// ListUserAchievements returns grants for a user after access check.
func (s *AchievementCatalogService) ListUserAchievements(ctx context.Context, requesterID, targetUserID string, requesterIsAdmin bool) ([]UserAchievementDTO, error) {
	reqID, err := domain.ParseUserID(requesterID)
	if err != nil {
		return nil, err
	}
	targetID, err := domain.ParseUserID(targetUserID)
	if err != nil {
		return nil, err
	}
	if !requesterIsAdmin && reqID != targetID {
		ok, err := s.buddy.IsActiveBuddyOf(ctx, reqID, targetID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, domain.ErrForbidden
		}
	}
	return s.listGrants(ctx, targetID)
}

func (s *AchievementCatalogService) listGrants(ctx context.Context, userID domain.UserID) ([]UserAchievementDTO, error) {
	grants, err := s.grants.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	defs, err := s.definitions.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	meta := make(map[domain.AchievementCode]domain.AchievementDefinition, len(defs))
	for _, d := range defs {
		meta[d.Code] = d
	}
	out := make([]UserAchievementDTO, 0, len(grants))
	for _, g := range grants {
		d := meta[g.AchievementCode]
		out = append(out, UserAchievementDTO{
			Code:        string(g.AchievementCode),
			Title:       d.Title,
			Description: d.Description,
			GrantedAt:   g.GrantedAt.UTC().Format(time.RFC3339),
		})
	}
	return out, nil
}
