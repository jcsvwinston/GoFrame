package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"example.com/showcase_clean/internal/services"
	gferrors "github.com/jcsvwinston/GoFrame/pkg/errors"
	"github.com/jcsvwinston/GoFrame/pkg/router"
)

type CategoryPayload struct {
	Name string `json:"name"`
}

type CategoryHandler struct {
	service *services.CategoryService
}

func NewCategoryHandler(service *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) Mount(r *router.Mux) {
	r.Resource("/categories", router.ResourceHandlers{
		List:     h.List,
		Create:   h.Create,
		Retrieve: h.Get,
		Update:   h.Update,
		Delete:   h.Delete,
	})
}

func (h *CategoryHandler) List(c *router.Context) error {
	records, err := h.service.List(c.Request.Context(), services.ListCategoryInput{
		Query: strings.TrimSpace(c.Query("q")),
	})
	if err != nil {
		return gferrors.InternalError("unable to list categories")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data":  records,
		"count": len(records),
	})
}

func (h *CategoryHandler) Get(c *router.Context) error {
	id, err := parseResourceID(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	record, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		return gferrors.NotFound("Category", strconv.FormatUint(uint64(id), 10))
	}

	return c.JSON(http.StatusOK, map[string]any{"data": record})
}

func (h *CategoryHandler) Create(c *router.Context) error {
	payload, err := decodeCategoryPayload(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	record, err := h.service.Create(c.Request.Context(), services.CreateCategoryInput{Name: payload.Name})
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]any{"data": record})
}

func (h *CategoryHandler) Update(c *router.Context) error {
	id, err := parseResourceID(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	payload, err := decodeCategoryPayload(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	record, err := h.service.Update(c.Request.Context(), id, services.UpdateCategoryInput{Name: payload.Name})
	if err != nil {
		return gferrors.NotFound("Category", strconv.FormatUint(uint64(id), 10))
	}

	return c.JSON(http.StatusOK, map[string]any{"data": record})
}

func (h *CategoryHandler) Delete(c *router.Context) error {
	id, err := parseResourceID(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		return gferrors.NotFound("Category", strconv.FormatUint(uint64(id), 10))
	}

	return c.NoContent()
}

func decodeCategoryPayload(r *http.Request) (CategoryPayload, error) {
	defer r.Body.Close()

	var payload CategoryPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return payload, errors.New("request body must be valid JSON")
	}

	payload.Name = strings.TrimSpace(payload.Name)
	if payload.Name == "" {
		return payload, errors.New("name is required")
	}

	return payload, nil
}

