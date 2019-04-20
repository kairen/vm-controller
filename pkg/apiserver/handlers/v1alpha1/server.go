package v1alpha1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kairen/vm-controller/pkg/apiserver/driver"
	"github.com/kairen/vm-controller/pkg/apiserver/types"
)

type ServerHandler struct {
	hypervisor driver.Interface
}

func (h *ServerHandler) Create(c *gin.Context) {
	s := &types.Server{}
	if err := c.ShouldBindJSON(s); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{})
		return
	}

	createServer, err := h.hypervisor.Create(s)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{})
		fmt.Println(err)
		return
	}
	c.JSON(http.StatusOK, createServer)
}

func (h *ServerHandler) List(c *gin.Context) {
	servers, err := h.hypervisor.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, []string{})
		return
	}
	c.JSON(http.StatusOK, servers)
}

func (h *ServerHandler) Get(c *gin.Context) {
	server, err := h.hypervisor.Get(c.Param("uuid"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{})
		return
	}

	if server == nil {
		c.JSON(http.StatusNotFound, map[string]string{})
		return
	}
	c.JSON(http.StatusOK, server)
}

func (h *ServerHandler) GetStatus(c *gin.Context) {
	status, err := h.hypervisor.GetStatus(c.Param("uuid"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{})
		return
	}

	if status == nil {
		c.JSON(http.StatusNotFound, map[string]string{})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *ServerHandler) Delete(c *gin.Context) {
	if err := h.hypervisor.Delete(c.Param("uuid")); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{})
		return
	}
	c.JSON(http.StatusOK, map[string]string{})
}

func (h *ServerHandler) CheckName(c *gin.Context) {
	check := h.hypervisor.CheckName(c.Param("name"))
	switch check {
	case -1:
		c.JSON(http.StatusInternalServerError, map[string]string{})
	case 0:
		c.JSON(http.StatusForbidden, map[string]string{})
	case 1:
		c.JSON(http.StatusOK, map[string]string{})
	}
}
