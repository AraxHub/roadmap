package telegram

// Update — минимальный кусок Telegram Update.
type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message"`
}

// Message — входящее сообщение.
type Message struct {
	MessageID int64  `json:"message_id"`
	Text      string `json:"text"`
	Chat      Chat   `json:"chat"`
}

// Chat — чат отправителя.
type Chat struct {
	ID int64 `json:"id"`
}
