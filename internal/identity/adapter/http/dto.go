package http

// LoginRequest is the login JSON body.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LogoutRequest is the logout JSON body.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// SetActiveRoleRequest is the active role JSON body.
type SetActiveRoleRequest struct {
	ActiveRole string `json:"active_role"`
}

// TokenResponse is returned with access credentials.
type TokenResponse struct {
	AccessToken           string `json:"access_token"`
	TokenType             string `json:"token_type"`
	ExpiresIn             int64  `json:"expires_in"`
	RefreshToken          string `json:"refresh_token,omitempty"`
	RequiresRoleSelection bool   `json:"requires_role_selection,omitempty"`
}

// UserResponse describes the authenticated user.
type UserResponse struct {
	ID         string   `json:"id"`
	Email      string   `json:"email"`
	Status     string   `json:"status"`
	Roles      []string `json:"roles"`
	ActiveRole *string  `json:"active_role"`
}

// LoginResponse combines tokens and user info.
type LoginResponse struct {
	Tokens TokenResponse `json:"tokens"`
	User   UserResponse  `json:"user"`
}

// MeResponse wraps user info for GET /auth/me.
type MeResponse struct {
	User UserResponse `json:"user"`
}

// ActiveRoleResponse is returned after updating active role.
type ActiveRoleResponse struct {
	Tokens TokenResponse `json:"tokens"`
	User   UserResponse  `json:"user"`
}
