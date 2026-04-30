package dtos

import "time"

type CreateArticleInput struct {
	Title     string `json:"title" validate:"required,min=3"`
	Content   string `json:"content"`
	Published bool   `json:"published"`
}

type CreateLeadInput struct {
	Name      string `json:"name" validate:"required,min=2"`
	Email     string `json:"email" validate:"required,email"`
	Company   string `json:"company"`
	WantsDemo bool   `json:"wants_demo"`
}

type ArticleDTO struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Published bool      `json:"published"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LeadDTO struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Company   string    `json:"company"`
	WantsDemo bool      `json:"wants_demo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LeadCapturePageData struct {
	Title       string
	Submitted   bool
	Form        CreateLeadInput
	FieldErrors map[string]string
}

type DemoTaskPayload struct {
	Kind     string `json:"kind"`
	Target   string `json:"target"`
	Source   string `json:"source"`
	QueuedAt string `json:"queued_at"`
}
