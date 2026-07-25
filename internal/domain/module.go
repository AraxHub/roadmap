package domain

// Module — модуль внутри спринта.
type Module struct {
	ID          string `json:"id"`
	SprintID    string `json:"sprint_id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Position    int    `json:"position"`
	IsPublished bool   `json:"is_published"`
}
