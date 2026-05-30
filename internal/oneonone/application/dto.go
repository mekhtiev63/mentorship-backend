package application

import (
	"encoding/json"
	"time"

	"github.com/go-mentorship-platform/backend/internal/oneonone/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// RequestDTO is API representation of a 1:1 request.
type RequestDTO struct {
	ID              string          `json:"id"`
	StudentID       string          `json:"student_id"`
	BuddyID         string          `json:"buddy_id"`
	Status          string          `json:"status"`
	Message         string          `json:"message"`
	PreferredSlots  json.RawMessage `json:"preferred_slots"`
	CalendarEventID *string         `json:"calendar_event_id,omitempty"`
	RejectReason    *string         `json:"reject_reason,omitempty"`
	ApprovedBy      *string         `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time      `json:"approved_at,omitempty"`
	BonusDebitedAt  *time.Time      `json:"bonus_debited_at,omitempty"`
	BonusReference  *string         `json:"bonus_reference,omitempty"`
	CancelledAt     *time.Time      `json:"cancelled_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	CostPoints      int64           `json:"cost_points"`
}

// ListResult is paginated requests.
type ListResult struct {
	Items []RequestDTO    `json:"items"`
	Meta  pagination.Meta `json:"meta"`
}

func toDTO(req domain.OneOnOneRequest) RequestDTO {
	slots := json.RawMessage(req.PreferredSlots)
	if len(slots) == 0 {
		slots = json.RawMessage("[]")
	}
	dto := RequestDTO{
		ID:             string(req.ID),
		StudentID:      string(req.StudentID),
		BuddyID:        string(req.BuddyID),
		Status:         string(req.Status),
		Message:        req.Message,
		PreferredSlots: slots,
		RejectReason:   req.RejectReason,
		ApprovedAt:     req.ApprovedAt,
		BonusDebitedAt: req.BonusDebitedAt,
		BonusReference: req.BonusReference,
		CancelledAt:    req.CancelledAt,
		CreatedAt:      req.CreatedAt,
		UpdatedAt:      req.UpdatedAt,
		CalendarEventID: req.CalendarEventID,
		CostPoints:     domain.OneOnOneCostPoints,
	}
	if req.ApprovedBy != nil {
		s := string(*req.ApprovedBy)
		dto.ApprovedBy = &s
	}
	return dto
}
