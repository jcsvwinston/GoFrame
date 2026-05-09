package controllers

import (
	"net/http"
	"strconv"
	"github.com/jcsvwinston/fleetmanager/internal/services"
	"github.com/jcsvwinston/nucleus/pkg/router"
)

type TelemetryHandler struct {
	service *services.TelemetryService
}

func NewTelemetryHandler(service *services.TelemetryService) *TelemetryHandler {
	return &TelemetryHandler{service: service}
}

func (h *TelemetryHandler) Mount(r *router.Mux) {
	r.Get("/api/telemetries", h.List)
	r.Post("/api/telemetries", h.Create)
	r.Delete("/api/telemetries/{id}", h.Delete)
}

func (h *TelemetryHandler) List(c *router.Context) error {
	res, err := h.service.List(c.Request.Context())
	if err != nil { return err }
	return c.JSON(http.StatusOK, res)
}

func (h *TelemetryHandler) Create(c *router.Context) error {
	var in struct { Name string `json:"name"` }
	if err := c.Bind(&in); err != nil { return err }
	res, err := h.service.Create(c.Request.Context(), in.Name)
	if err != nil { return err }
	return c.JSON(http.StatusCreated, res)
}

func (h *TelemetryHandler) Delete(c *router.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil { return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"}) }
	if err := h.service.Delete(c.Request.Context(), id); err != nil { return err }
	return c.NoContent()
}
