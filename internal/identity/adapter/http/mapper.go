package http

import "github.com/go-mentorship-platform/backend/internal/identity/application"

func toTokenResponse(pair application.TokenPair) TokenResponse {
	return TokenResponse{
		AccessToken:           pair.AccessToken,
		TokenType:             "Bearer",
		ExpiresIn:             pair.AccessTokenExpiresIn,
		RefreshToken:          pair.RefreshToken,
		RequiresRoleSelection: pair.RequiresRoleSelection,
	}
}

func toUserResponse(user application.UserInfo) UserResponse {
	return UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		Status:     string(user.Status),
		Roles:      user.Roles,
		ActiveRole: user.ActiveRole,
	}
}
