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

type AuthorPayload struct {
	Name string `json:"name"`
}

type AuthorHandler struct {
	service *services.AuthorService
}

func NewAuthorHandler(service *services.AuthorService) *AuthorHandler {
	return &AuthorHandler{service: service}
}

func (h *AuthorHandler) Mount(r *router.Mux) {
	r.Resource("/authors", router.ResourceHandlers{
		List:     h.List,
		Create:   h.Create,
		Retrieve: h.Get,
		Update:   h.Update,
		Delete:   h.Delete,
	})
}

func (h *AuthorHandler) List(c *router.Context) error {
	records, err := h.service.List(c.Request.Context(), services.ListAuthorInput{
		Query: strings.TrimSpace(c.Query("q")),
	})
	if err != nil {
		return gferrors.InternalError("unable to list authors")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data":  records,
		"count": len(records),
	})
}

func (h *AuthorHandler) Get(c *router.Context) error {
	id, err := parseResourceID(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	record, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		return gferrors.NotFound("Author", strconv.FormatUint(uint64(id), 10))
	}

	return c.JSON(http.StatusOK, map[string]any{"data": record})
}

func (h *AuthorHandler) Create(c *router.Context) error {
	payload, err := decodeAuthorPayload(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	record, err := h.service.Create(c.Request.Context(), services.CreateAuthorInput{Name: payload.Name})
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]any{"data": record})
}

func (h *AuthorHandler) Update(c *router.Context) error {
	id, err := parseResourceID(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	payload, err := decodeAuthorPayload(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	record, err := h.service.Update(c.Request.Context(), id, services.UpdateAuthorInput{Name: payload.Name})
	if err != nil {
		return gferrors.NotFound("Author", strconv.FormatUint(uint64(id), 10))
	}

	return c.JSON(http.StatusOK, map[string]any{"data": record})
}

func (h *AuthorHandler) Delete(c *router.Context) error {
	id, err := parseResourceID(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		return gferrors.NotFound("Author", strconv.FormatUint(uint64(id), 10))
	}

	return c.NoContent()
}

func decodeAuthorPayload(r *http.Request) (AuthorPayload, error) {
	defer r.Body.Close()

	var payload AuthorPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return payload, errors.New("request body must be valid JSON")
	}

	payload.Name = strings.TrimSpace(payload.Name)
	if payload.Name == "" {
		return payload, errors.New("name is required")
	}

	return payload, nil
}

func parseResourceID(r *http.Request) (uint, error) {
	raw := strings.TrimSpace(r.PathValue("id"))
	if raw == "" {
		return 0, errors.New("resource id is required")
	}

	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("resource id must be a positive integer")
	}

	return uint(id), nil
}

func writeError(w http.ResponseWriter, err error) {
	router.Error(w, err)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
