package domain

import "time"

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type Dataset struct {
	ID           string     `json:"id" db:"id"`
	UserID       string     `json:"user_id" db:"user_id"`
	Author       string     `json:"author,omitempty" db:"author"`
	Title        string     `json:"title" db:"title"`
	FilePath     string     `json:"file_path" db:"file_path"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	IndexedAt    *time.Time `json:"indexed_at" db:"indexed_at"`
	TopicID      *string    `json:"topic_id,omitempty" db:"topic_id"`
	AssignmentID *string    `json:"assignment_id,omitempty" db:"assignment_id"`
	Content      string     `json:"content,omitempty"`
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
	Title        string `form:"title" binding:"required"`
	AssignmentID string `form:"assignment_id" binding:"required"`
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

type Topic struct {
	ID          string    `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	CreatedByID string    `json:"created_by_id" db:"created_by_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type TopicAssignment struct {
	ID          string    `json:"id" db:"id"`
	TopicID     string    `json:"topic_id" db:"topic_id"`
	StudentID   string    `json:"student_id" db:"student_id"`
	StudentName string    `json:"student_name" db:"student_name"`
	AssignedBy  string    `json:"assigned_by" db:"assigned_by"`
	AssignedAt  time.Time `json:"assigned_at" db:"assigned_at"`
}

type StudentInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type CreateTopicRequest struct {
	Title       string        `json:"title" binding:"required"`
	Description string        `json:"description"`
	Students    []StudentInfo `json:"students" binding:"required,min=1,dive"`
}

type AddStudentsRequest struct {
	Students []StudentInfo `json:"students" binding:"required,min=1,dive"`
}

type SearchStudentsRequest struct {
	Query string `json:"query" binding:"required"`
}

type SearchStudentsResponse struct {
	Students []StudentInfo `json:"students"`
	Total    int           `json:"total"`
}

type TopicResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AssignedTopicResponse struct {
	ID           string        `json:"id"`
	Topic        TopicResponse `json:"topic"`
	AssignmentID string        `json:"assignment_id"`
	AssignedAt   time.Time     `json:"assigned_at"`
	HasDataset   bool          `json:"has_dataset"`
}

type TopicStudentResponse struct {
	ID         string      `json:"id"`
	Student    StudentInfo `json:"student"`
	AssignedAt time.Time   `json:"assigned_at"`
}
