package services

import (
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
	"net/http"
	"strings"
	"sync"
)

var pluginsDiscovery *PluginDiscovery

var oncePluginsDiscovery sync.Once

func GetPluginsDiscovery(client common.HttpClient) *PluginDiscovery {
	oncePluginsDiscovery.Do(initPluginsDiscovery(client))
	return pluginsDiscovery
}

func initPluginsDiscovery(client common.HttpClient) func() {
	return func() {
		pluginsDiscovery = &PluginDiscovery{url: config.ServerConfiguration.PluginDiscovery.Url, httpClient: client}
	}
}

type pluginsDiscoveryEndpoint string

const (
	listPlugins pluginsDiscoveryEndpoint = "/api/plugins"
)

type Plugin struct {
	Name  string `json:"name"`
	Route string `json:"route"`
	URL   string `json:"url"`
}

type PluginDiscovery struct {
	url        string
	httpClient common.HttpClient
}

func (c *PluginDiscovery) ListPlugins() (*[]Plugin, error) {
	var res []Plugin
	url := strings.Join([]string{c.url, string(listPlugins)}, "")
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return handleResponse(resp, &res)
}
