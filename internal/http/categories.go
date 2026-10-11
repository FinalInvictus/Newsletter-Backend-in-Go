package httpapi

import (
	"net/http"
	"strings"

	dbmodel "newsletter-backend/internal/db"
)

type categoryInput struct {
	Name string  `json:"name"`
	Slug *string `json:"slug"`
}

func (h *Handler) RegisterCategoryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /categories", h.listCategories)
	mux.HandleFunc("POST /categories", h.createCategory)
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	var categories []dbmodel.Category
	if err := h.DB.WithContext(r.Context()).Order("name ASC").Find(&categories).Error; err != nil {
		writeDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

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
