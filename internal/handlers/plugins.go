package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/services"
)

type ListPluginsResponse struct {
	Plugins []services.Plugin `json:"plugins"`
}

// ListPlugins godoc
// @Summary List plugins
// @Description List all plugins
// @Tags plugins
// @Accept json
// @Produce json
// @Success 200 {object} ListPluginsResponse
// @Param Authorization header string true "Bearer token"
// @Param tenantId path string true "Tenant ID"
// @Router /plugin-discovery [get]
func ListPlugins(c *gin.Context, e common.Env) (any, error) {
	pluginGateway := services.GetPluginsDiscovery(e.GetHttpClient())
	plugins, err := pluginGateway.ListPlugins()
	if err != nil {
		return nil, err
	}
	return ListPluginsResponse{Plugins: *plugins}, nil

}
