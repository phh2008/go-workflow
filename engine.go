package easyworkflow

import (
	"context"
	"log/slog"

	"github.com/Bunny3th/easy-workflow/internal/event"
	"github.com/Bunny3th/easy-workflow/internal/entity"
	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/service"
	"github.com/Bunny3th/easy-workflow/internal/web"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Engine 工作流引擎，提供流程定义、实例管理和任务处理等功能。
type Engine struct {
	internal *service.Engine
}

// Config 工作流引擎配置。
type Config struct {
	IgnoreEventError bool         // 是否忽略事件错误
	Logger           *slog.Logger // 日志记录器，为 nil 时使用默认日志
}

// New 创建并初始化工作流引擎。
// db 由调用方创建并配置（包括连接池、日志等），New 仅负责 AutoMigrate 和初始化内部组件。
func New(db *gorm.DB, cfg Config) (*Engine, error) {
	svcCfg := service.Config{
		IgnoreEventError: cfg.IgnoreEventError,
		Logger:           cfg.Logger,
	}
	eng, err := service.NewEngine(db, svcCfg)
	if err != nil {
		return nil, err
	}
	return &Engine{internal: eng}, nil
}

// Close 关闭引擎。数据库连接由调用方负责关闭。
func (e *Engine) Close() error {
	return e.internal.Close()
}

// RegisterNodeEvent 注册节点事件处理器。
// 节点事件用于 NodeStartEvents、NodeEndEvents、TaskFinishEvents。
func (e *Engine) RegisterNodeEvent(name string, handler event.NodeEventHandler) {
	e.internal.RegisterNodeEvent(name, handler)
}

// RegisterProcEvent 注册流程事件处理器。
// 流程事件用于 RevokeEvents（流程撤销事件）。
func (e *Engine) RegisterProcEvent(name string, handler event.ProcEventHandler) {
	e.internal.RegisterProcEvent(name, handler)
}

// DB 返回底层的 GORM 数据库实例。
func (e *Engine) DB() *gorm.DB {
	return e.internal.DB()
}

// ResolveVariables 解析流程实例变量。
func (e *Engine) ResolveVariables(ctx context.Context, params model.ResolveVariablesParams) (map[string]string, error) {
	return e.internal.ResolveVariables(ctx, params)
}

// --- 流程定义 ---

// ProcessParse 解析流程定义 JSON 资源。
func (e *Engine) ProcessParse(ctx context.Context, resource string) (*model.Process, error) {
	return e.internal.ProcessParse(ctx, resource)
}

// ProcessSave 保存流程定义，返回流程ID。
func (e *Engine) ProcessSave(ctx context.Context, req model.ProcessSaveReq) (int, error) {
	return e.internal.ProcessSave(ctx, req)
}

// GetProcessDefine 获取流程定义。
func (e *Engine) GetProcessDefine(ctx context.Context, procID int) (*model.Process, error) {
	return e.internal.GetProcessDefine(ctx, procID)
}

// GetProcessList 获取某个来源下所有流程定义。
func (e *Engine) GetProcessList(ctx context.Context, source string) ([]entity.ProcDef, error) {
	return e.internal.GetProcessList(ctx, source)
}

// --- 流程实例 ---

// InstanceStart 启动流程实例，返回实例ID。
func (e *Engine) InstanceStart(ctx context.Context, req model.InstanceStartReq) (int, error) {
	return e.internal.InstanceStart(ctx, req)
}

// InstanceRevoke 撤销流程实例。
func (e *Engine) InstanceRevoke(ctx context.Context, req model.InstanceRevokeReq) error {
	return e.internal.InstanceRevoke(ctx, req)
}

// GetInstanceInfo 获取流程实例信息。
func (e *Engine) GetInstanceInfo(ctx context.Context, instID int) (model.InstanceView, error) {
	return e.internal.GetInstanceInfo(ctx, instID)
}

// GetInstanceStartByUser 获取用户发起的流程实例列表（含分页和总数）。
func (e *Engine) GetInstanceStartByUser(ctx context.Context, req model.InstanceListReq) (*model.PageData[model.InstanceView], error) {
	return e.internal.GetInstanceStartByUser(ctx, req)
}

// --- 任务 ---

// TaskPass 任务通过。directlyToRejected 为 true 时，直接跳到上一个驳回自己的节点。
func (e *Engine) TaskPass(ctx context.Context, req model.TaskActionReq, directlyToRejected bool) error {
	return e.internal.TaskPass(ctx, req, directlyToRejected)
}

// TaskReject 任务驳回。
func (e *Engine) TaskReject(ctx context.Context, req model.TaskActionReq) error {
	return e.internal.TaskReject(ctx, req)
}

// TaskTransfer 任务转交。
func (e *Engine) TaskTransfer(ctx context.Context, req model.TaskTransferReq) error {
	return e.internal.TaskTransfer(ctx, req)
}

// TaskFreeReject 自由驳回（驳回到上游指定节点）。
func (e *Engine) TaskFreeReject(ctx context.Context, req model.TaskFreeRejectReq) error {
	return e.internal.TaskFreeReject(ctx, req)
}

// GetTaskInfo 获取任务信息。
func (e *Engine) GetTaskInfo(ctx context.Context, taskID int) (model.TaskView, error) {
	return e.internal.GetTaskInfo(ctx, taskID)
}

// GetTaskToDoList 获取待办任务列表（含分页和总数）。
func (e *Engine) GetTaskToDoList(ctx context.Context, req model.TaskListReq) (*model.PageData[model.TaskView], error) {
	return e.internal.GetTaskToDoList(ctx, req)
}

// GetTaskFinishedList 获取已办任务列表（含分页和总数）。
func (e *Engine) GetTaskFinishedList(ctx context.Context, req model.TaskFinishedListReq) (*model.PageData[model.TaskView], error) {
	return e.internal.GetTaskFinishedList(ctx, req)
}

// TaskUpstreamNodeList 获取任务上游节点列表。
func (e *Engine) TaskUpstreamNodeList(ctx context.Context, taskID int) ([]model.Node, error) {
	return e.internal.TaskUpstreamNodeList(ctx, taskID)
}

// GetInstanceTaskHistory 获取流程实例任务历史。
func (e *Engine) GetInstanceTaskHistory(ctx context.Context, instID int) ([]model.TaskView, error) {
	return e.internal.GetInstanceTaskHistory(ctx, instID)
}

// WhatCanIDo 获取当前任务可执行的操作。
func (e *Engine) WhatCanIDo(ctx context.Context, taskID int) (model.TaskAction, error) {
	return e.internal.WhatCanIDo(ctx, taskID)
}

// --- Web API ---

// WebConfig Web API 配置。
type WebConfig struct {
	BaseURL     string // API 基础路径
	ShowSwagger bool   // 是否显示 Swagger 文档
	SwaggerURL  string // Swagger 文档 URL
	Addr        string // 监听地址
}

// StartWebAPI 启动 Web API 服务。
func (e *Engine) StartWebAPI(ginEngine *gin.Engine, cfg WebConfig) error {
	return web.StartWebAPI(e.internal, ginEngine, web.WebConfig{
		BaseURL:     cfg.BaseURL,
		ShowSwagger: cfg.ShowSwagger,
		SwaggerURL:  cfg.SwaggerURL,
		Addr:        cfg.Addr,
	})
}
