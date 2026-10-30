package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	dbmodel "github.com/FinalInvictus/Newsletter-Backend-in-Go.git/internal/db"
)

type articleInput struct {
	Slug              string          `json:"slug"`
	Title             string          `json:"title"`
	Description       string          `json:"description"`
	Contents          json.RawMessage `json:"contents"`
	CoverImage        *string         `json:"cover_image"`
	Tags              []string        `json:"tags"`
	CategoryID        *uuid.UUID      `json:"category_id"`
	Status            string          `json:"status"`
	Type              string          `json:"type"`
	EstimatedReadTime int             `json:"estimated_read_time"`
}

// validateArticle checks required fields and supported values before an article reaches PostgreSQL.
func validateArticle(input articleInput) string {
	for _, field := range []struct{ value, name string }{
		{input.Slug, "slug"}, {input.Title, "title"}, {input.Description, "description"},
		{input.Status, "status"}, {input.Type, "type"},
	} {
		if message := requireText(field.value, field.name); message != "" {
			return message
		}
	}
	if len(input.Contents) == 0 || !json.Valid(input.Contents) {
		return "contents must be valid JSON"
	}
	if input.Status != "draft" && input.Status != "published" && input.Status != "archived" {
		return "status must be draft, published, or archived"
	}
	if input.EstimatedReadTime <= 0 {
		return "estimated_read_time must be greater than zero"
	}
	if input.Type != "article" && input.Type != "letter" && input.Type != "note" {
		return "type must be article, letter, or note"
	}
	return ""
}

// RegisterArticleRoutes binds public article reads and editorial create, update, and delete endpoints.
func (h *Handler) RegisterArticleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /articles", h.listPublishedArticles)
	mux.HandleFunc("GET /articles/all", h.listAllArticles)
	mux.HandleFunc("GET /articles/{slug}", h.getArticleBySlug)
	mux.HandleFunc("POST /articles", h.createArticle)
	mux.HandleFunc("PUT /articles/{id}", h.updateArticle)
	mux.HandleFunc("DELETE /articles/{id}", h.deleteArticle)
}

// createArticle validates and stores a new article, assigning its first publication timestamp when needed.
func (h *Handler) createArticle(w http.ResponseWriter, r *http.Request) {
	var input articleInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if message := validateArticle(input); message != "" {
		writeError(w, http.StatusBadRequest, message)
		return
	}

	article := dbmodel.Article{
		Slug: strings.TrimSpace(input.Slug), Title: strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description), Contents: dbmodel.JSONB(input.Contents),
		CoverImage: input.CoverImage, Tags: pq.StringArray(input.Tags), CategoryID: input.CategoryID,
		Status: input.Status, Type: input.Type, EstimatedReadTime: input.EstimatedReadTime,
	}
	if article.Tags == nil {
		article.Tags = []string{}
	}
	if article.Status == "published" {
		now := time.Now().UTC()
		article.PublishedAt = &now
	}
	if err := h.DB.WithContext(r.Context()).Create(&article).Error; err != nil {
		writeDBError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, article)
}

// updateArticle replaces an existing article's editable fields and records its first publication time.
func (h *Handler) updateArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "id must be a valid UUID")
		return
	}
	var input articleInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if message := validateArticle(input); message != "" {
		writeError(w, http.StatusBadRequest, message)
		return
	}

	ctxDB := h.DB.WithContext(r.Context())
	var article dbmodel.Article
	if err := ctxDB.First(&article, "id = ?", id).Error; err != nil {
		writeDBError(w, err)
		return
	}
	previousStatus := article.Status
	article.Slug = strings.TrimSpace(input.Slug)
	article.Title = strings.TrimSpace(input.Title)
	article.Description = strings.TrimSpace(input.Description)
	article.Contents = dbmodel.JSONB(input.Contents)
	article.CoverImage = input.CoverImage
	article.Tags = pq.StringArray(input.Tags)
	if article.Tags == nil {
		article.Tags = []string{}
	}
	article.CategoryID = input.CategoryID
	article.Status = input.Status
	article.Type = input.Type
	article.EstimatedReadTime = input.EstimatedReadTime
	if previousStatus != "published" && article.Status == "published" {
		now := time.Now().UTC()
		article.PublishedAt = &now
	} else if article.Status != "published" {
		article.PublishedAt = nil
	}
	if err := ctxDB.Save(&article).Error; err != nil {
		writeDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, article)
}

// deleteArticle removes the requested article and reports when its identifier does not exist.
func (h *Handler) deleteArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "id must be a valid UUID")
		return
	}
	result := h.DB.WithContext(r.Context()).Where("id = ?", id).Delete(&dbmodel.Article{})
	if result.Error != nil {
		writeDBError(w, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		writeError(w, http.StatusNotFound, "article not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// listPublishedArticles returns only reader-visible articles, newest first.
func (h *Handler) listPublishedArticles(w http.ResponseWriter, r *http.Request) {
	var articles []dbmodel.Article
	if err := h.DB.WithContext(r.Context()).Where("status = ?", "published").Order("created_at DESC").Find(&articles).Error; err != nil {
		writeDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, articles)
}

// listAllArticles returns every article for the editorial interface, newest first.
func (h *Handler) listAllArticles(w http.ResponseWriter, r *http.Request) {
	var articles []dbmodel.Article
	if err := h.DB.WithContext(r.Context()).Order("created_at DESC").Find(&articles).Error; err != nil {
		writeDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, articles)
}

// getArticleBySlug finds one article by its public, human-readable URL slug.
func (h *Handler) getArticleBySlug(w http.ResponseWriter, r *http.Request) {
	var article dbmodel.Article
	err := h.DB.WithContext(r.Context()).Where("slug = ? AND status = ?", r.PathValue("slug"), "published").First(&article).Error
	if err != nil {
		writeDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, article)
}
