package controllers

import (
	"douyin-action-example/internal/actions/assets"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AssetHandler struct {
}

func NewAssetHandler() *AssetHandler {
	return &AssetHandler{}
}

func (h *AssetHandler) OpenApiSpecYaml(c *gin.Context) {
	c.Header("Content-Type", "text/yaml; charset=utf-8")
	c.Header("Access-Control-Allow-Origin", "*")
	c.String(http.StatusOK, assets.OpenApiSpecYaml)
}
