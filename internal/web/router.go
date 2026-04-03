package web

import (
	"github.com/Bunny3th/easy-workflow/internal/service"
	"github.com/Bunny3th/easy-workflow/internal/web/handler"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// registerRoutes 注册所有路由。
func registerRoutes(eng *service.Engine, ginEngine *gin.Engine, cfg WebConfig) {
	if cfg.ShowSwagger {
		ginEngine.GET(cfg.SwaggerURL, ginSwagger.WrapHandler(swaggerFiles.Handler, func(c *ginSwagger.Config) {
			c.InstanceName = "easyworkflow"
		}))
	}

	procDefHandler := handler.NewProcDefHandler(eng)
	procInstHandler := handler.NewProcInstHandler(eng)
	taskHandler := handler.NewTaskHandler(eng)

	router := ginEngine.Group(cfg.BaseURL)

	// 流程定义
	router.POST("/def/save", procDefHandler.Save)
	router.GET("/def/list", procDefHandler.ListBySource)
	router.GET("/def/get", procDefHandler.GetByID)

	// 流程实例
	router.POST("/inst/start", procInstHandler.Start)
	router.GET("/inst/start/by", procInstHandler.StartByUser)
	router.POST("/inst/revoke", procInstHandler.Revoke)
	router.GET("/inst/task_history", procInstHandler.TaskHistory)

	// 任务
	router.POST("/task/pass", taskHandler.Pass)
	router.POST("/task/pass/directly", taskHandler.PassDirectly)
	router.POST("/task/reject", taskHandler.Reject)
	router.POST("/task/reject/free", taskHandler.FreeReject)
	router.POST("/task/transfer", taskHandler.Transfer)
	router.GET("/task/todo", taskHandler.ToDoList)
	router.GET("/task/finished", taskHandler.FinishedList)
	router.GET("/task/upstream", taskHandler.UpstreamNodeList)
	router.GET("/task/action", taskHandler.WhatCanIDo)
	router.GET("/task/info", taskHandler.Info)
}
