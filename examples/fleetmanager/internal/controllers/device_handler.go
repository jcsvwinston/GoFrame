package controllers

import (
	"net/http"
	"strconv"
	"github.com/jcsvwinston/fleetmanager/internal/services"
	"github.com/jcsvwinston/nucleus/pkg/router"
)

type DeviceHandler struct {
	service *services.DeviceService
}

func NewDeviceHandler(service *services.DeviceService) *DeviceHandler {
	return &DeviceHandler{service: service}
}

func (h *DeviceHandler) Mount(r *router.Mux) {
	r.Get("/api/devices", h.List)
	r.Post("/api/devices", h.Create)
	r.Delete("/api/devices/{id}", h.Delete)
}

func (h *DeviceHandler) List(c *router.Context) error {
	res, err := h.service.List(c.Request.Context())
	if err != nil { return err }
	return c.JSON(http.StatusOK, res)
}

func (h *DeviceHandler) Create(c *router.Context) error {
	var in struct { Name string `json:"name"` }
	if err := c.Bind(&in); err != nil { return err }
	res, err := h.service.Create(c.Request.Context(), in.Name)
	if err != nil { return err }
	return c.JSON(http.StatusCreated, res)
}

func (h *DeviceHandler) Delete(c *router.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil { return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"}) }
	if err := h.service.Delete(c.Request.Context(), id); err != nil { return err }
	return c.NoContent()
}
