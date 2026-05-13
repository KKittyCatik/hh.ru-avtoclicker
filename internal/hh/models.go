package hh

import "time"

type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Employer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Salary struct {
	From int `json:"from"`
	To   int `json:"to"`
}

type Vacancy struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Employer    Employer `json:"employer"`
	Salary      Salary   `json:"salary"`
	Schedule    string   `json:"schedule"`
	URL         string   `json:"alternate_url"`
}

type VacancySearchResponse struct {
	Items []Vacancy `json:"items"`
	Found int       `json:"found"`
	Pages int       `json:"pages"`
	Page  int       `json:"page"`
}

type MessageOption struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Message struct {
	ID                   string          `json:"id"`
	NegotiationID        string          `json:"negotiation_id"`
	Type                 string          `json:"type"`
	From                 string          `json:"from"`
	Text                 string          `json:"text"`
	Options              []MessageOption `json:"options"`
	CreatedAt            time.Time       `json:"created_at"`
	NegotiationCreatedAt time.Time       `json:"negotiation_created_at"`
	QuickReplyOptionID   string          `json:"quick_reply_option_id,omitempty"`
	NeedsHumanInput      bool            `json:"needs_human_input,omitempty"`
	PotentialBot         bool            `json:"potential_bot,omitempty"`
}

type Negotiation struct {
	ID          string    `json:"id"`
	VacancyID   string    `json:"vacancy_id"`
	ResumeID    string    `json:"resume_id"`
	Status      string    `json:"status"`
	IsBot       bool      `json:"is_bot"`
	NeedsReply  bool      `json:"needs_reply"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastMessage Message   `json:"last_message"`
}

type Resume struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Status        string    `json:"status"`
	NextPublishAt time.Time `json:"next_publish_at"`
}

type ApplyResult struct {
	VacancyID string `json:"vacancy_id"`
	Company   string `json:"company"`
	Status    string `json:"status"`
	Reason    string `json:"reason,omitempty"`
}

type QuestionOption struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Question struct {
	ID      string           `json:"id"`
	Type    string           `json:"type"`
	Text    string           `json:"text"`
	Options []QuestionOption `json:"options,omitempty"`
}

type Answer struct {
	QuestionID  string   `json:"question_id"`
	OptionIDs   []string `json:"option_ids,omitempty"`
	Text        string   `json:"text,omitempty"`
	Number      int      `json:"number,omitempty"`
	NeedsReview bool     `json:"needs_review"`
}
