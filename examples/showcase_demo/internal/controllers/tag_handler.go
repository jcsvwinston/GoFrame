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

type TagPayload struct {
	Name string `json:"name"`
}

type TagHandler struct {
	service *services.TagService
}

func NewTagHandler(service *services.TagService) *TagHandler {
	return &TagHandler{service: service}
}

func (h *TagHandler) Mount(r *router.Mux) {
	r.Resource("/tags", router.ResourceHandlers{
		List:     h.List,
		Create:   h.Create,
		Retrieve: h.Get,
		Update:   h.Update,
		Delete:   h.Delete,
	})
}

func (h *TagHandler) List(c *router.Context) error {
	records, err := h.service.List(c.Request.Context(), services.ListTagInput{
		Query: strings.TrimSpace(c.Query("q")),
	})
	if err != nil {
		return gferrors.InternalError("unable to list tags")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data":  records,
		"count": len(records),
	})
}

func (h *TagHandler) Get(c *router.Context) error {
	id, err := parseResourceID(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	record, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		return gferrors.NotFound("Tag", strconv.FormatUint(uint64(id), 10))
	}

	return c.JSON(http.StatusOK, map[string]any{"data": record})
}

func (h *TagHandler) Create(c *router.Context) error {
	payload, err := decodeTagPayload(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	record, err := h.service.Create(c.Request.Context(), services.CreateTagInput{Name: payload.Name})
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]any{"data": record})
}

func (h *TagHandler) Update(c *router.Context) error {
	id, err := parseResourceID(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	payload, err := decodeTagPayload(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	record, err := h.service.Update(c.Request.Context(), id, services.UpdateTagInput{Name: payload.Name})
	if err != nil {
		return gferrors.NotFound("Tag", strconv.FormatUint(uint64(id), 10))
	}

	return c.JSON(http.StatusOK, map[string]any{"data": record})
}

func (h *TagHandler) Delete(c *router.Context) error {
	id, err := parseResourceID(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		return gferrors.NotFound("Tag", strconv.FormatUint(uint64(id), 10))
	}

	return c.NoContent()
}

func decodeTagPayload(r *http.Request) (TagPayload, error) {
	defer r.Body.Close()

	var payload TagPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return payload, errors.New("request body must be valid JSON")
	}

	payload.Name = strings.TrimSpace(payload.Name)
	if payload.Name == "" {
		return payload, errors.New("name is required")
	}

	return payload, nil
}

