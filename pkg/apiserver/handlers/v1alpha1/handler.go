package v1alpha1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kairen/vm-controller/pkg/apiserver/driver"
	"github.com/kairen/vm-controller/pkg/version"
)

type Handler struct {
	hypervisor driver.Interface
	Server     *ServerHandler
}

func New(driver driver.Interface) *Handler {
	h := &Handler{hypervisor: driver}
	h.Server = &ServerHandler{hypervisor: driver}
	return h
}

func (h *Handler) Version(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"version": version.GetVersion(),
	})
}

func (h *Handler) Healthz(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}
