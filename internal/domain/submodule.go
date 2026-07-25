package domain

// Submodule — подмодуль (урок) внутри модуля.
type Submodule struct {
	ID          string `json:"id"`
	ModuleID    string `json:"module_id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Position    int    `json:"position"`
	IsPublished bool   `json:"is_published"`
}
