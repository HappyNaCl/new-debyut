package events

// OrganizationRenamedEvent tells other services an organization's public name
// changed, e.g. so the event search index can refresh organizer names.
type OrganizationRenamedEvent struct {
	OrganizationId string
	Name           string
}
