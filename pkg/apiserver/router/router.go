package router

import (
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	_ "github.com/kairen/vm-controller/api"
	"github.com/kairen/vm-controller/pkg/apiserver/handlers/v1alpha1"
)

type Router struct {
	engine *gin.Engine
}

func New() *Router {
	gin.DisableConsoleColor()
	engine := gin.Default()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	return &Router{engine: engine}
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

func (r *Router) LinkSwaggerAPI(swagger bool) {
	if swagger {
		r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}

func (r *Router) LinkHandler(handler *v1alpha1.Handler) {
	r.engine.GET("/version", handler.Version)
	r.engine.GET("/healthz", handler.Healthz)

	apiv1alpha1 := r.engine.Group("/api/v1alpha1")
	{
		apiv1alpha1.POST("/servers", handler.Server.Create)
		apiv1alpha1.GET("/servers", handler.Server.List)
		apiv1alpha1.GET("/servers/:uuid", handler.Server.Get)
		apiv1alpha1.GET("/servers/:uuid/status", handler.Server.GetStatus)
		apiv1alpha1.DELETE("/servers/:uuid", handler.Server.Delete)
		apiv1alpha1.GET("/check/:name", handler.Server.CheckName)
	}
}
