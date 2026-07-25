package domain

// ChainItem — элемент глобальной цепочки обучения (порядок прохождения).
type ChainItem struct {
	SubmoduleID    string
	SubmoduleSlug  string
	SubmoduleTitle string
	ModuleID       string
	ModuleSlug     string
	ModuleTitle    string
	SprintID       string
	SprintSlug     string
	SprintTitle    string
}
