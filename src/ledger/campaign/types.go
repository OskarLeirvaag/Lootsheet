// Package campaign provides CRUD operations for campaign management.
package campaign

// Record represents a campaign row in the database.
type Record struct {
	ID        string
	Name      string
	CreatedAt string
	UpdatedAt string
}
