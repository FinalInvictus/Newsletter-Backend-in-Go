package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	dbmodel "newsletter-backend/internal/db"
)

type commentInput struct {
	AuthorName *string `json:"author_name"`
	Content    string  `json:"content"`
}

func (h *Handler) RegisterCommentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /articles/{id}/comments", h.listComments)
	mux.HandleFunc("POST /articles/{id}/comments", h.createComment)
}

func (h *Handler) listComments(w http.ResponseWriter, r *http.Request) {
	articleID, ok := parseUUID(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "id must be a valid UUID")
		return
	}
	var comments []dbmodel.Comment
	err := h.DB.WithContext(r.Context()).Where("article_id = ?", articleID).Order("created_at ASC").Find(&comments).Error
	if err != nil {
		writeDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, comments)
}

func (h *Handler) createComment(w http.ResponseWriter, r *http.Request) {
	articleID, ok := parseUUID(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusBadRequest, "id must be a valid UUID")
		return
	}
	var input commentInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if message := requireText(input.Content, "content"); message != "" {
		writeError(w, http.StatusBadRequest, message)
		return
	}
	if input.AuthorName != nil {
		trimmed := strings.TrimSpace(*input.AuthorName)
		input.AuthorName = &trimmed
		if trimmed == "" {
			input.AuthorName = nil
		}
	}

	var article dbmodel.Article
	err := h.DB.WithContext(r.Context()).Select("id").First(&article, "id = ?", articleID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "article not found")
			return
		}
		writeDBError(w, err)
		return
	}
	comment := dbmodel.Comment{ArticleID: articleID, AuthorName: input.AuthorName, Content: strings.TrimSpace(input.Content)}
	if err := h.DB.WithContext(r.Context()).Create(&comment).Error; err != nil {
		writeDBError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, comment)
}
