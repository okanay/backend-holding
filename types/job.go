package types

import (
	"time"

	"github.com/google/uuid"
)

// ====================
// ENUM TİPLERİ
// ====================

// JobStatus - İş ilanı durumu için enum tipi
type JobStatus string

const (
	JobStatusDraft     JobStatus = "draft"
	JobStatusPublished JobStatus = "published"
	JobStatusClosed    JobStatus = "closed"
	JobStatusDeleted   JobStatus = "deleted"
)

// ====================
// VERİTABANI MODELLERİ
// ====================

// Job - İş ilanı ana yapısı (job_postings tablosu)
type Job struct {
	ID         uuid.UUID   `db:"id" json:"id"`
	UserID     uuid.UUID   `db:"user_id" json:"userId"`
	Slug       string      `db:"slug" json:"slug"`
	Status     JobStatus   `db:"status" json:"status"`
	Deadline   *time.Time  `db:"deadline" json:"deadline,omitempty"`
	CreatedAt  time.Time   `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time   `db:"updated_at" json:"updatedAt"`
	Details    *JobDetails `json:"details,omitempty"`
	Categories []string    `json:"categories,omitempty"`
}

// JobDetails - İş ilanı detayları (job_posting_details tablosu)
type JobDetails struct {
	ID              uuid.UUID `db:"id" json:"id"`
	Title           string    `db:"title" json:"title"`
	Description     string    `db:"description" json:"description,omitempty"`
	Image           string    `db:"image" json:"image,omitempty"`
	Location        string    `db:"location" json:"location,omitempty"`
	WorkMode        string    `db:"work_mode" json:"workMode,omitempty"`
	EmploymentType  string    `db:"employment_type" json:"employmentType,omitempty"`
	ExperienceLevel string    `db:"experience_level" json:"experienceLevel,omitempty"`
	HTML            string    `db:"html" json:"html"`
	JSON            string    `db:"json" json:"json"`
	FormType        string    `db:"form_type" json:"formType"`
	Applicants      int       `db:"applicants" json:"applicants"`
}

// JobCategory - İş kategorisi (job_categories tablosu)
type JobCategory struct {
	Name        string    `db:"name" json:"name"`
	DisplayName string    `db:"display_name" json:"displayName"`
	UserID      uuid.UUID `db:"user_id" json:"userId"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`
}

// JobApplication - İş başvurusu (job_applications tablosu)
type JobApplication struct {
	ID        uuid.UUID `db:"id" json:"id"`
	JobID     uuid.UUID `db:"job_id" json:"jobId"`
	FullName  string    `db:"full_name" json:"fullName"`
	Email     string    `db:"email" json:"email"`
	Phone     string    `db:"phone" json:"phone"`
	FormType  string    `db:"form_type" json:"formType"`
	FormJSON  string    `db:"form_json" json:"formJson"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

// JobsTrackingCode - İş başvuru takip kodu (jobs_tracking_codes tablosu)
type JobsTrackingCode struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	TrackingCode string    `db:"tracking_code" json:"trackingCode"`
	ExpiresAt    time.Time `db:"expires_at" json:"expiresAt"`
	IsUsed       bool      `db:"is_used" json:"isUsed"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at" json:"updatedAt"`
}

// JobsTrackingSession - İş başvuru takip oturumu (jobs_tracking_sessions tablosu)
type JobsTrackingSession struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	SessionToken string    `db:"session_token" json:"sessionToken"`
	IPAddress    string    `db:"ip_address" json:"ipAddress"`
	UserAgent    string    `db:"user_agent" json:"userAgent"`
	ExpiresAt    time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at" json:"updatedAt"`
}

// ====================
// VIEW MODELLERİ
// ====================

// JobView - İş ilanı görünümü (API yanıtı için)
type JobView struct {
	ID        uuid.UUID  `json:"id"`
	Slug      string     `json:"slug"`
	Status    JobStatus  `json:"status"`
	Deadline  *time.Time `json:"deadline,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`

	// İlişkili alanlar (job_details tablosundan gelen)
	Details    JobDetailsView    `json:"details"`
	Categories []JobCategoryView `json:"categories,omitempty"`
}

// JobDetailsView - İş ilanı detayları görünümü
type JobDetailsView struct {
	Title           string `json:"title"`
	Description     string `json:"description,omitempty"`
	Image           string `json:"image,omitempty"`
	Location        string `json:"location,omitempty"`
	WorkMode        string `json:"workMode,omitempty"`
	EmploymentType  string `json:"employmentType,omitempty"`
	ExperienceLevel string `json:"experienceLevel,omitempty"`
	HTML            string `json:"html"`
	JSON            string `json:"json"`
	FormType        string `json:"formType"`
	Applicants      int    `json:"applicants"`
}

// JobCategoryView - Kategori görünümü
type JobCategoryView struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	CreatedAt   time.Time `json:"createdAt"`
}

// JobApplicationView - Başvuru görünümü
type JobApplicationView struct {
	ID        uuid.UUID `json:"id"`
	JobID     uuid.UUID `json:"jobId"`
	JobTitle  string    `json:"jobTitle,omitempty"`
	FullName  string    `json:"fullName"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	FormType  string    `json:"formType"`
	FormJSON  string    `json:"formJson"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// ====================
// INPUT MODELLERİ
// ====================

// JobInput - İş ilanı ortak input yapısı (Create ve Update için)
type JobInput struct {
	Title           string     `json:"title,omitempty"`
	Description     string     `json:"description,omitempty"`
	Image           string     `json:"image,omitempty"`
	Slug            string     `json:"slug,omitempty"`
	Status          JobStatus  `json:"status,omitempty"`
	Location        string     `json:"location,omitempty"`
	WorkMode        string     `json:"workMode,omitempty"`
	EmploymentType  string     `json:"employmentType,omitempty"`
	ExperienceLevel string     `json:"experienceLevel,omitempty"`
	HTML            string     `json:"html,omitempty"`
	JSON            string     `json:"json,omitempty"`
	FormType        string     `json:"formType,omitempty"`
	Categories      []string   `json:"categories,omitempty"`
	Deadline        *time.Time `json:"deadline,omitempty"`
}

// JobStatusInput - İş ilanı durumu güncelleme
type JobStatusInput struct {
	Status JobStatus `json:"status" binding:"required"`
}

// JobCategoryInput - Kategori ortak input yapısı
type JobCategoryInput struct {
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"displayName" binding:"required"`
}

// JobApplicationInput - Başvuru giriş verisi
type JobApplicationInput struct {
	FullName string `json:"fullName" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	FormType string `json:"formType" binding:"required"`
	FormJSON string `json:"formJson" binding:"required"`
}

// JobApplicationStatusInput - Başvuru durumu güncelleme
type JobApplicationStatusInput struct {
	Status string `json:"status" binding:"required"`
}

// JobTrackingCodeInput - Takip kodu isteği
type JobTrackingCodeInput struct {
	Email string `json:"email" binding:"required,email"`
}

// JobTrackingVerifyInput - Takip kodu doğrulama
type JobTrackingVerifyInput struct {
	Email        string `json:"email" binding:"required,email"`
	TrackingCode string `json:"trackingCode" binding:"required"`
}

// ====================
// ARAMA PARAMETRELERİ
// ====================

// JobSearchParams - İlan arama parametreleri
type JobSearchParams struct {
	Status    JobStatus `form:"status"`
	Category  string    `form:"category"`
	Query     string    `form:"q"` // Başlık/açıklama içinde arama
	Location  string    `form:"location"`
	WorkMode  string    `form:"workMode"`
	Page      int       `form:"page,default=1"`
	Limit     int       `form:"limit,default=10"`
	SortBy    string    `form:"sortBy,default=createdAt"`
	SortOrder string    `form:"sortOrder,default=desc"`
}

// JobApplicationSearchParams - Başvuru arama parametreleri
type JobApplicationSearchParams struct {
	JobID     uuid.UUID `form:"jobId"`
	Status    string    `form:"status"`
	FullName  string    `form:"fullName"`
	Email     string    `form:"email"`
	StartDate string    `form:"startDate"` // YYYY-MM-DD formatında
	EndDate   string    `form:"endDate"`   // YYYY-MM-DD formatında
	Page      int       `form:"page,default=1"`
	Limit     int       `form:"limit,default=10"`
	SortBy    string    `form:"sortBy,default=createdAt"`
	SortOrder string    `form:"sortOrder,default=desc"`
}
