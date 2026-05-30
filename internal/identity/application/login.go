package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-mentorship-platform/backend/internal/identity/domain"
	"github.com/google/uuid"
)

// LoginService authenticates a user and issues tokens.
type LoginService struct {
	credentials  domain.CredentialsRepository
	roles        domain.RoleRepository
	preferences  domain.UserPreferencesRepository
	refreshStore domain.RefreshTokenRepository
	hasher       domain.PasswordHasher
	tokens       domain.TokenIssuer
	refreshGen   domain.RefreshTokenGenerator
	clock        Clock
	refreshTTL   time.Duration
	accessTTL    time.Duration
}

// NewLoginService builds LoginService.
func NewLoginService(
	credentials domain.CredentialsRepository,
	roles domain.RoleRepository,
	preferences domain.UserPreferencesRepository,
	refreshStore domain.RefreshTokenRepository,
	hasher domain.PasswordHasher,
	tokens domain.TokenIssuer,
	refreshGen domain.RefreshTokenGenerator,
	clock Clock,
	accessTTL, refreshTTL time.Duration,
) *LoginService {
	return &LoginService{
		credentials:  credentials,
		roles:        roles,
		preferences:  preferences,
		refreshStore: refreshStore,
		hasher:       hasher,
		tokens:       tokens,
		refreshGen:   refreshGen,
		clock:        clock,
		refreshTTL:   refreshTTL,
		accessTTL:    accessTTL,
	}
}

// Login authenticates by email and password.
func (s *LoginService) Login(ctx context.Context, emailRaw, password string) (TokenPair, UserInfo, error) {
	password = strings.TrimSpace(password)
	email, err := domain.ParseEmail(emailRaw)
	if err != nil {
		return TokenPair{}, UserInfo{}, domain.ErrInvalidCredentials
	}

	creds, err := s.credentials.FindByEmail(ctx, email)
	if err != nil {
		return TokenPair{}, UserInfo{}, domain.ErrInvalidCredentials
	}

	if !creds.Account.CanLogin() {
		return TokenPair{}, UserInfo{}, domain.ErrAccountInactive
	}

	if err := s.hasher.Verify(creds.PasswordHash, password); err != nil {
		return TokenPair{}, UserInfo{}, domain.ErrInvalidCredentials
	}

	roleSet, err := s.roles.ListByUser(ctx, creds.UserID)
	if err != nil {
		return TokenPair{}, UserInfo{}, fmt.Errorf("list roles: %w", err)
	}
	if roleSet.Len() == 0 {
		return TokenPair{}, UserInfo{}, domain.ErrInvalidCredentials
	}

	activeRole, requiresSelection, err := s.resolveActiveRole(ctx, creds.UserID, roleSet)
	if err != nil {
		return TokenPair{}, UserInfo{}, err
	}

	pair, err := s.issueTokens(ctx, creds, roleSet, activeRole)
	if err != nil {
		return TokenPair{}, UserInfo{}, err
	}
	pair.RequiresRoleSelection = requiresSelection

	user := toUserInfo(creds, roleSet, activeRole)
	return pair, user, nil
}

func (s *LoginService) resolveActiveRole(ctx context.Context, userID domain.UserID, roles domain.RoleSet) (*domain.Role, bool, error) {
	prefs, err := s.preferences.Get(ctx, userID)
	if err != nil {
		return nil, false, fmt.Errorf("get preferences: %w", err)
	}

	if roles.Len() == 1 {
		role := roles[0]
		if prefs.ActiveRole == nil || *prefs.ActiveRole != role {
			if err := s.preferences.SetActiveRole(ctx, userID, &role); err != nil {
				return nil, false, fmt.Errorf("set active role: %w", err)
			}
		}
		return &role, false, nil
	}

	if prefs.ActiveRole != nil && roles.Contains(*prefs.ActiveRole) {
		role := *prefs.ActiveRole
		return &role, false, nil
	}

	return nil, true, nil
}

func (s *LoginService) issueTokens(
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

	tokenID := uuid.NewString()
	refresh := domain.RefreshToken{
		ID:        domain.RefreshTokenID(tokenID),
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

func toUserInfo(creds domain.Credentials, roles domain.RoleSet, active *domain.Role) UserInfo {
	var activeStr *string
	if active != nil {
		v := string(*active)
		activeStr = &v
	}
	return UserInfo{
		ID:         creds.UserID.String(),
		Email:      creds.Email.String(),
		Status:     creds.Account.Status,
		Roles:      roles.Strings(),
		ActiveRole: activeStr,
	}
}
