package controllers

import (
	"net/http"
	"strconv"
	"github.com/jcsvwinston/fleetmanager/internal/services"
	"github.com/jcsvwinston/GoFrame/pkg/router"
)

type FleetHandler struct {
	service *services.FleetService
}

func NewFleetHandler(service *services.FleetService) *FleetHandler {
	return &FleetHandler{service: service}
}

func (h *FleetHandler) Mount(r *router.Mux) {
	r.Get("/api/fleets", h.List)
	r.Post("/api/fleets", h.Create)
	r.Delete("/api/fleets/{id}", h.Delete)
}

func (h *FleetHandler) List(c *router.Context) error {
	res, err := h.service.List(c.Request.Context())
	if err != nil { return err }
	return c.JSON(http.StatusOK, res)
}

func (h *FleetHandler) Create(c *router.Context) error {
	var in struct { Name string `json:"name"` }
	if err := c.Bind(&in); err != nil { return err }
	res, err := h.service.Create(c.Request.Context(), in.Name)
	if err != nil { return err }
	return c.JSON(http.StatusCreated, res)
}

func (h *FleetHandler) Delete(c *router.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil { return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"}) }
	if err := h.service.Delete(c.Request.Context(), id); err != nil { return err }
	return c.NoContent()
}
