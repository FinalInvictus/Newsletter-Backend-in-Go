package httpapi

import (
	"net/http"
	"strings"

	dbmodel "github.com/FinalInvictus/Newsletter-Backend-in-Go.git/internal/db"
)

type categoryInput struct {
	Name string  `json:"name"`
	Slug *string `json:"slug"`
}

// RegisterCategoryRoutes binds category listing and creation endpoints.
func (h *Handler) RegisterCategoryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /categories", h.listCategories)
	mux.HandleFunc("POST /categories", h.createCategory)
}

// listCategories returns categories in name order for display and article assignment.
func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	var categories []dbmodel.Category
	if err := h.DB.WithContext(r.Context()).Order("name ASC").Find(&categories).Error; err != nil {
		writeDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

// createCategory validates and persists one new editorial category.
func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var input categoryInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if message := requireText(input.Name, "name"); message != "" {
		writeError(w, http.StatusBadRequest, message)
		return
	}
	category := dbmodel.Category{Name: strings.TrimSpace(input.Name), Slug: input.Slug}
	if err := h.DB.WithContext(r.Context()).Create(&category).Error; err != nil {
		writeDBError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, category)
}
