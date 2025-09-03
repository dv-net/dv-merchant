package models

type NotificationCategory string

const (
	NotificationCategorySystem NotificationCategory = "system"
	NotificationCategoryEvent  NotificationCategory = "event"
	NotificationCategoryReport NotificationCategory = "report"
)

func (nc *NotificationCategory) String() string {
	if nc == nil {
		return ""
	}

	return string(*nc)
}

func (nc *NotificationCategory) IsFullDisableAllowed() bool {
	if nc == nil {
		return false
	}

	return *nc != NotificationCategorySystem
}
