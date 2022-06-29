package handlers

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

var (
	VersionID string = "Default"
)

func (hS *HttpServer) CreateRoutes() {
	// projects:
	hS.Router.GET("/projects", hS.ProjectsGetList)
	hS.Router.POST("/projects", hS.ProjectCreate)
	hS.Router.GET("/projects/:projectIdOrName", hS.ProjectGet)
	hS.Router.PUT("/projects/:projectIdOrName", hS.ProjectUpdate)
	hS.Router.DELETE("/projects/:projectIdOrName", hS.ProjectDelete)

	// clusters:
	hS.Router.GET("/projects/:projectIdOrName/clusters", hS.ClustersGet)
	hS.Router.POST("/projects/:projectIdOrName/clusters", hS.ClusterCreate)
	hS.Router.GET("/projects/:projectIdOrName/clusters/:clusterIdOrName", hS.ClustersGetByName)
	hS.Router.GET("/projects/:projectIdOrName/clusters/:clusterIdOrName/status", hS.ClustersStatusGetByName)
	hS.Router.PUT("/projects/:projectIdOrName/clusters/:clusterIdOrName", hS.ClustersUpdate)
	hS.Router.DELETE("/projects/:projectIdOrName/clusters/:clusterIdOrName", hS.ClustersDelete)

	// service type:
	hS.Router.POST("/configs", hS.ConfigsCreateService)
	hS.Router.GET("/configs", hS.ConfigsGetServices)
	hS.Router.GET("/configs/:serviceType", hS.ConfigsGetService)
	hS.Router.PUT("/configs/:serviceType", hS.ConfigsUpdateService)
	hS.Router.DELETE("/configs/:serviceType", hS.ConfigsDeleteService)

	// service type versions:
	hS.Router.GET("/configs/:serviceType/versions", hS.ConfigsGetVersions)
	hS.Router.POST("/configs/:serviceType/versions", hS.ConfigsCreateVersion)
	hS.Router.GET("/configs/:serviceType/versions/:versionId", hS.ConfigsGetVersion)
	hS.Router.PUT("/configs/:serviceType/versions/:versionId", hS.ConfigsUpdateVersion)
	hS.Router.DELETE("/configs/:serviceType/versions/:versionId", hS.ConfigsDeleteVersion)
	hS.Router.POST("/configs/:serviceType/versions/:versionId/configs", hS.ConfigsCreateConfigParam)

	// images:
	hS.Router.Handle("GET", "/images", hS.ImagesGetList)
	hS.Router.Handle("GET", "/images/:imageName", hS.ImageGet)
	hS.Router.Handle("POST", "/images", hS.ImageCreate)
	hS.Router.Handle("PUT", "/images/:imageName", hS.ImageUpdate)
	hS.Router.Handle("DELETE", "/images/:imageName", hS.ImageDelete)

	// flavors:
	hS.Router.Handle("POST", "/flavors", hS.FlavorCreate)
	hS.Router.Handle("GET", "/flavors", hS.FlavorsGetList)
	hS.Router.Handle("GET", "/flavors/:flavorIdOrName", hS.FlavorGet)
	hS.Router.Handle("PUT", "/flavors/:flavorIdOrName", hS.FlavorUpdate)
	hS.Router.Handle("DELETE", "/flavors/:flavorIdOrName", hS.FlavorDelete)

	// templates:
	hS.Router.GET("/templates", hS.TemplatesGetList)
	hS.Router.POST("/templates", hS.TemplateCreate)
	hS.Router.GET("/templates/:templateID", hS.TemplateGet)
	hS.Router.PUT("/templates/:templateID", hS.TemplateUpdate)
	hS.Router.DELETE("/templates/:templateID", hS.TemplateDelete)

	// project templates
	hS.Router.GET("/projects/:projectIdOrName/templates", hS.TemplatesGetList)
	hS.Router.POST("/projects/:projectIdOrName/templates", hS.TemplateCreate)
	hS.Router.GET("/projects/:projectIdOrName/templates/:templateID", hS.TemplateGet)
	hS.Router.PUT("/projects/:projectIdOrName/templates/:templateID", hS.TemplateUpdate)
	hS.Router.DELETE("/projects/:projectIdOrName/templates/:templateID", hS.TemplateDelete)

	// swagger UI:
	hS.Router.ServeFiles("/api/*filepath", http.Dir("./api/rest"))

	// logs routes :
	hS.Router.Handle("GET", "/logs/launcher", hS.ServeAnsibleServiceLog)
	hS.Router.Handle("GET", "/logs/http_server", hS.ServeHttpServerLog)
	hS.Router.Handle("GET", "/logs/projects/:projectIdOrName/clusters/:clusterIdOrName", hS.ServeClusterLog)

	// service version:
	hS.Router.Handle("GET", "/version", hS.GetVersion)
}

func (hS *HttpServer) GetVersion(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "/version GET"
	hS.Logger.Info("Get " + request)
	ResponseOK(w, VersionID, request)
}
