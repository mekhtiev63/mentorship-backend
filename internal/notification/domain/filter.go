package domain

// ReadStatusFilter filters by read state.
type ReadStatusFilter string

const (
	ReadStatusUnread ReadStatusFilter = "unread"
	ReadStatusRead   ReadStatusFilter = "read"
)

// NotificationListFilter lists inbox rows for a recipient.
type NotificationListFilter struct {
	ReadStatus       *ReadStatusFilter
	NotificationType *NotificationType
}
