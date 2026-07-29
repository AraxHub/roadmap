package domain

// Типы блоков контента подмодуля.
const (
	BlockTypeMarkdown = "markdown"
	BlockTypeAnswer   = "answer"
	BlockTypeImage    = "image"
)

// ContentBlock — элемент полотна подмодуля (JSON в submodule_contents.blocks).
type ContentBlock struct {
	ID      string `json:"id"`
	Type    string `json:"type"` // markdown | answer | image
	MD      string `json:"md,omitempty"`
	Title   string `json:"title,omitempty"` // для answer: заголовок спойлера (по умолчанию «Ответ»)
	ImageID string `json:"image_id,omitempty"`
	Alt     string `json:"alt,omitempty"`
}

// MaxContentImageBytes — лимит размера одной картинки (10 МБ).
const MaxContentImageBytes = 10 * 1024 * 1024

// AllowedContentImageMIME — допустимые MIME для content_images.
var AllowedContentImageMIME = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/webp": true,
	"image/gif":  true,
}
