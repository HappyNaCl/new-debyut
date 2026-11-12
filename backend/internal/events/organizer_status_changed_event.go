package events

type OrganizerStatusChangedEvent struct {
	UserId      string
	IsOrganizer bool
}
