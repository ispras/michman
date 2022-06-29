package handlers

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

var (
	VersionID = "Default"
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
	hS.Router.POST("/configs", hS.ConfigsCreateServiceType)
	hS.Router.GET("/configs", hS.ConfigsServiceTypesGetList)
	hS.Router.GET("/configs/:serviceTypeIdOrName", hS.ConfigsServiceTypeGet)
	hS.Router.PUT("/configs/:serviceTypeIdOrName", hS.ConfigsUpdateServiceType)
	hS.Router.DELETE("/configs/:serviceTypeIdOrName", hS.ConfigsDeleteServiceType)

	// service type versions:
	hS.Router.GET("/configs/:serviceTypeIdOrName/versions", hS.ConfigsGetVersions)
	hS.Router.POST("/configs/:serviceTypeIdOrName/versions", hS.ConfigsCreateVersion)
	hS.Router.GET("/configs/:serviceTypeIdOrName/versions/:versionId", hS.ConfigsGetVersion)
	hS.Router.PUT("/configs/:serviceTypeIdOrName/versions/:versionId", hS.ConfigsUpdateVersion)
	hS.Router.DELETE("/configs/:serviceTypeIdOrName/versions/:versionId", hS.ConfigsDeleteVersion)
	hS.Router.POST("/configs/:serviceTypeIdOrName/versions/:versionId/configs", hS.ConfigsCreateConfigParam)

	// images:
	hS.Router.GET("/images", hS.ImagesGetList)
	hS.Router.GET("/images/:imageName", hS.ImageGet)
	hS.Router.POST("/images", hS.ImageCreate)
	hS.Router.PUT("/images/:imageName", hS.ImageUpdate)
	hS.Router.DELETE("/images/:imageName", hS.ImageDelete)

	// flavors:
	hS.Router.POST("/flavors", hS.FlavorCreate)
	hS.Router.GET("/flavors", hS.FlavorsGetList)
	hS.Router.GET("/flavors/:flavorIdOrName", hS.FlavorGet)
	hS.Router.PUT("/flavors/:flavorIdOrName", hS.FlavorUpdate)
	hS.Router.DELETE("/flavors/:flavorIdOrName", hS.FlavorDelete)

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
	hS.Router.GET("/logs/launcher", hS.ServeAnsibleServiceLog)
	hS.Router.GET("/logs/http_server", hS.ServeHttpServerLog)
	hS.Router.GET("/logs/projects/:projectIdOrName/clusters/:clusterIdOrName", hS.ServeClusterLog)

	// service version:
	hS.Router.GET("/version", hS.GetVersion)
}

func (hS *HttpServer) GetVersion(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	request := "/version GET"
	hS.Logger.Info("Get " + request)
	ResponseOK(w, VersionID, request)
}
