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

type CommentPayload struct {
	Name string `json:"name"`
}

type CommentHandler struct {
	service *services.CommentService
}

func NewCommentHandler(service *services.CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}

func (h *CommentHandler) Mount(r *router.Mux) {
	r.Resource("/comments", router.ResourceHandlers{
		List:     h.List,
		Create:   h.Create,
		Retrieve: h.Get,
		Update:   h.Update,
		Delete:   h.Delete,
	})
}

func (h *CommentHandler) List(c *router.Context) error {
	records, err := h.service.List(c.Request.Context(), services.ListCommentInput{
		Query: strings.TrimSpace(c.Query("q")),
	})
	if err != nil {
		return gferrors.InternalError("unable to list comments")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data":  records,
		"count": len(records),
	})
}

func (h *CommentHandler) Get(c *router.Context) error {
	id, err := parseResourceID(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	record, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		return gferrors.NotFound("Comment", strconv.FormatUint(uint64(id), 10))
	}

	return c.JSON(http.StatusOK, map[string]any{"data": record})
}

func (h *CommentHandler) Create(c *router.Context) error {
	payload, err := decodeCommentPayload(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	record, err := h.service.Create(c.Request.Context(), services.CreateCommentInput{Name: payload.Name})
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]any{"data": record})
}

func (h *CommentHandler) Update(c *router.Context) error {
	id, err := parseResourceID(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	payload, err := decodeCommentPayload(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	record, err := h.service.Update(c.Request.Context(), id, services.UpdateCommentInput{Name: payload.Name})
	if err != nil {
		return gferrors.NotFound("Comment", strconv.FormatUint(uint64(id), 10))
	}

	return c.JSON(http.StatusOK, map[string]any{"data": record})
}

func (h *CommentHandler) Delete(c *router.Context) error {
	id, err := parseResourceID(c.Request)
	if err != nil {
		return gferrors.BadRequest(err.Error())
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		return gferrors.NotFound("Comment", strconv.FormatUint(uint64(id), 10))
	}

	return c.NoContent()
}

func decodeCommentPayload(r *http.Request) (CommentPayload, error) {
	defer r.Body.Close()

	var payload CommentPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return payload, errors.New("request body must be valid JSON")
	}

	payload.Name = strings.TrimSpace(payload.Name)
	if payload.Name == "" {
		return payload, errors.New("name is required")
	}

	return payload, nil
}

