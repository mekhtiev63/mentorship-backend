package http

type createUserRequest struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

type replaceRolesRequest struct {
	Roles []string `json:"roles"`
}

type buddyAssignmentRequest struct {
	StudentID string `json:"student_id"`
	BuddyID   string `json:"buddy_id"`
}
