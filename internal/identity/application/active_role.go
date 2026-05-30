package application

import (
	"context"
	"fmt"
	"time"

	"github.com/go-mentorship-platform/backend/internal/identity/domain"
	"github.com/google/uuid"
)

// ActiveRoleService updates the persisted active role and re-issues tokens.
type ActiveRoleService struct {
	credentials  domain.CredentialsRepository
	roles        domain.RoleRepository
	preferences  domain.UserPreferencesRepository
	refreshStore domain.RefreshTokenRepository
	tokens       domain.TokenIssuer
	refreshGen   domain.RefreshTokenGenerator
	clock        Clock
	refreshTTL   time.Duration
	accessTTL    time.Duration
}

// NewActiveRoleService builds ActiveRoleService.
func NewActiveRoleService(
	credentials domain.CredentialsRepository,
	roles domain.RoleRepository,
	preferences domain.UserPreferencesRepository,
	refreshStore domain.RefreshTokenRepository,
	tokens domain.TokenIssuer,
	refreshGen domain.RefreshTokenGenerator,
	clock Clock,
	accessTTL, refreshTTL time.Duration,
) *ActiveRoleService {
	return &ActiveRoleService{
		credentials:  credentials,
		roles:        roles,
		preferences:  preferences,
		refreshStore: refreshStore,
		tokens:       tokens,
		refreshGen:   refreshGen,
		clock:        clock,
		refreshTTL:   refreshTTL,
		accessTTL:    accessTTL,
	}
}

// SetActiveRole validates and persists the active role, then issues new tokens.
func (s *ActiveRoleService) SetActiveRole(ctx context.Context, userID domain.UserID, roleRaw string) (TokenPair, UserInfo, error) {
	role, err := domain.ParseRole(roleRaw)
	if err != nil {
		return TokenPair{}, UserInfo{}, domain.ErrInvalidRole
	}

	roleSet, err := s.roles.ListByUser(ctx, userID)
	if err != nil {
		return TokenPair{}, UserInfo{}, fmt.Errorf("list roles: %w", err)
	}
	if !roleSet.Contains(role) {
		return TokenPair{}, UserInfo{}, domain.ErrActiveRoleNotAllowed
	}

	if err := s.preferences.SetActiveRole(ctx, userID, &role); err != nil {
		return TokenPair{}, UserInfo{}, fmt.Errorf("set active role: %w", err)
	}

	creds, err := s.credentials.FindByID(ctx, userID)
	if err != nil {
		return TokenPair{}, UserInfo{}, err
	}
	if !creds.Account.CanLogin() {
		return TokenPair{}, UserInfo{}, domain.ErrAccountInactive
	}

	pair, err := s.issueTokens(ctx, creds, roleSet, &role)
	if err != nil {
		return TokenPair{}, UserInfo{}, err
	}

	user := toUserInfo(creds, roleSet, &role)
	return pair, user, nil
}

func (s *ActiveRoleService) issueTokens(
	ctx context.Context,
	creds domain.Credentials,
	roles domain.RoleSet,
	activeRole *domain.Role,
) (TokenPair, error) {
	now := s.clock.Now()

	rawRefresh, hash, err := s.refreshGen.Generate()
	if err != nil {
		return TokenPair{}, fmt.Errorf("refresh token: %w", err)
	}

	refresh := domain.RefreshToken{
		ID:        domain.RefreshTokenID(uuid.NewString()),
		UserID:    creds.UserID,
		TokenHash: hash,
		ExpiresAt: now.Add(s.refreshTTL),
	}
	if err := s.refreshStore.Store(ctx, refresh); err != nil {
		return TokenPair{}, fmt.Errorf("store refresh: %w", err)
	}

	access, err := s.tokens.Issue(domain.AccessTokenClaims{
		UserID:     creds.UserID,
		Roles:      roles,
		ActiveRole: activeRole,
		ExpiresAt:  now.Add(s.accessTTL),
		TokenID:    uuid.NewString(),
	})
	if err != nil {
		return TokenPair{}, fmt.Errorf("issue access: %w", err)
	}

	return TokenPair{
		AccessToken:          access,
		AccessTokenExpiresIn: int64(s.accessTTL.Seconds()),
		RefreshToken:         rawRefresh,
	}, nil
}
