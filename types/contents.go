package types

import (
	"time"

	"github.com/google/uuid"
)

// ====================
// ENUM TİPLERİ
// ====================

// ContentStatus - İçerik durumu için enum tipi
type ContentStatus string

const (
	ContentStatusDraft     ContentStatus = "draft"
	ContentStatusPublished ContentStatus = "published"
	ContentStatusClosed    ContentStatus = "closed"  // Migration'daki 'closed' ile eşleşiyor
	ContentStatusDeleted   ContentStatus = "deleted" // Migration'daki 'deleted' ile eşleşiyor
)

// ====================
// VERİTABANI MODELİ
// ====================

// Content - İçerik ana yapısı (contents tablosu)
// Bu model, veritabanı şemanızdaki (000008_contents.up.sql) contents tablosunu yansıtır.
type Content struct {
	ID          uuid.UUID     `db:"id" json:"id"`
	UserID      *uuid.UUID    `db:"user_id" json:"userId,omitempty"`
	Slug        string        `db:"slug" json:"slug"`
	Identifier  string        `db:"identifier" json:"identifier"`
	Language    string        `db:"language" json:"language"`
	Title       string        `db:"title" json:"title"`
	Description *string       `db:"description" json:"description,omitempty"`
	Category    string        `db:"category" json:"category"`
	ImageURL    *string       `db:"image_url" json:"imageUrl,omitempty"`
	DetailsJSON *string       `db:"details_json" json:"detailsJson,omitempty"`
	ContentJSON string        `db:"content_json" json:"contentJson"`
	ContentHTML string        `db:"content_html" json:"contentHtml"`
	Status      ContentStatus `db:"status" json:"status"`
	CreatedAt   time.Time     `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time     `db:"updated_at" json:"updatedAt"`
}

// ====================
// VIEW MODELLERİ (API Yanıtları İçin)
// ====================

// ContentView - Tek bir içerik dil versiyonunun API'de nasıl görüneceği.
type ContentView struct {
	ID          uuid.UUID       `json:"id"`
	Slug        string          `json:"slug"`
	Identifier  string          `json:"identifier"`
	Language    string          `json:"language"`
	Title       string          `json:"title"`
	Description *string         `json:"description,omitempty"`
	Category    string          `json:"category"`
	ImageURL    *string         `json:"imageUrl,omitempty"`
	DetailsJSON *map[string]any `json:"detailsJson,omitempty"`
	ContentJSON *map[string]any `json:"contentJson"`
	ContentHTML string          `json:"contentHtml"`
	Status      ContentStatus   `json:"status"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// ====================
// INPUT MODELLERİ (İstekler İçin)
// ====================

// ContentInput - Yeni içerik oluşturma veya güncelleme için ortak input yapısı.
// Güncelleme yaparken, sadece gönderilen alanlar güncellenir (PATCH metodolojisi).
// Oluşturma sırasında identifier boş gönderilebilir, backend yeni bir tane oluşturur
// veya bir identifier'a bağlı yeni dil versiyonu ekleniyorsa identifier dolu gönderilir.
type ContentInput struct {
	Slug        string          `json:"slug" binding:"required,min=3,max=255"`
	Identifier  *uuid.UUID      `json:"identifier,omitempty"`
	Language    string          `json:"language" binding:"required,min=2,max=10"`
	Title       string          `json:"title" binding:"required,min=3,max=255"`
	Description *string         `json:"description,omitempty"`
	Category    string          `json:"category" binding:"required"`
	ImageURL    *string         `json:"imageUrl,omitempty" binding:"omitempty,url"`
	DetailsJSON *map[string]any `json:"detailsJson,omitempty"`
	ContentJSON map[string]any  `json:"contentJson" binding:"required"`
	ContentHTML string          `json:"contentHtml" binding:"required"`
	Status      ContentStatus   `json:"status,omitempty" binding:"omitempty,oneof=draft published closed deleted"`
}

// ContentStatusInput - Sadece içerik durumunu güncellemek için input.
type ContentStatusInput struct {
	Status ContentStatus `json:"status" binding:"required,oneof=draft published closed deleted"`
}

// ====================
// ARAMA PARAMETRELERİ (Listeleme İçin)
// ====================

// ContentSearchParams - İçerikleri listelerken kullanılacak arama parametreleri.
type ContentSearchParams struct {
	Status     ContentStatus `form:"status"`
	Language   string        `form:"language"`
	Identifier string        `form:"identifier"`
	Category   string        `form:"category"`
	Query      string        `form:"q"`
	UserID     string        `form:"userId"`
	Page       int           `form:"page,default=1"`
	Limit      int           `form:"limit,default=10"`
	SortBy     string        `form:"sortBy,default=createdAt"`
	SortOrder  string        `form:"sortOrder,default=desc"`
}
