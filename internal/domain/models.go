package domain

import "time"

type User struct {
	ID       string `json:"id"`
	UserName string `json:"username"`
	Role     string `json:"role"`
}

type Dataset struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Title     string     `json:"title" db:"title"`
	FilePath  string     `json:"file_path" db:"file_path"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	IndexedAt *time.Time `json:"indexed_at" db:"indexed_at"`
	Content   string     `json:"content,omitempty"`
	Author    string     `json:"author,omitempty"`
}

type DatasetListResponse struct {
	Datasets []Dataset `json:"datasets"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

type AskRequest struct {
	Question string `json:"question" binding:"required"`
}

type AskResponse struct {
	Answer    string     `json:"answer"`
	Citations []Citation `json:"citations"`
}

type Citation struct {
	ChunkID          int     `json:"chunk_id"`
	Score            float64 `json:"score"`
	OriginalScore    float64 `json:"original_score,omitempty"`
	ScoreImprovement float64 `json:"score_improvement,omitempty"`
}

type CreateDatasetRequest struct {
	Title string `form:"title" binding:"required"`
}

type UpdateDatasetRequest struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

type DatasetResponse struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Author    string     `json:"author"`
	UserID    string     `json:"user_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	IndexedAt *time.Time `json:"indexed_at,omitempty"`
}

type IndexResponse struct {
	Success bool   `json:"success"`
	Chunks  int    `json:"chunks"`
	Message string `json:"message,omitempty"`
}
