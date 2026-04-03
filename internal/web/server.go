package web

import (
	"github.com/Bunny3th/easy-workflow/internal/service"
	"github.com/gin-gonic/gin"
)

// WebConfig Web API 配置
type WebConfig struct {
	BaseURL     string // API 基础路径
	ShowSwagger bool   // 是否显示 Swagger 文档
	SwaggerURL  string // Swagger 文档 URL
	Addr        string // 监听地址
}

// StartWebAPI 启动 Web API 服务。
func StartWebAPI(eng *service.Engine, ginEngine *gin.Engine, cfg WebConfig) error {
	registerRoutes(eng, ginEngine, cfg)
	return ginEngine.Run(cfg.Addr)
}
