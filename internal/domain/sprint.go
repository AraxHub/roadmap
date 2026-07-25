package domain

// Sprint — спринт (группа модулей).
type Sprint struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Position    int    `json:"position"`
	IsPublished bool   `json:"is_published"`
}
