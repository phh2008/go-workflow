package easyworkflow

import (
	"context"
	"log/slog"

	"github.com/Bunny3th/easy-workflow/internal/entity"
	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/service"
	"github.com/Bunny3th/easy-workflow/internal/web"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
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

// RegisterEvents 注册事件处理器。
func (e *Engine) RegisterEvents(eventStructs ...any) {
	e.internal.RegisterEvents(eventStructs...)
}

// DB 返回底层的 GORM 数据库实例。
func (e *Engine) DB() *gorm.DB {
	return e.internal.DB()
}

// ResolveVariables 解析流程实例变量。
func (e *Engine) ResolveVariables(ctx context.Context, instID int, variables []string) (map[string]string, error) {
	return e.internal.ResolveVariables(ctx, instID, variables)
}

// --- 流程定义 ---

// ProcessParse 解析流程定义 JSON 资源。
func (e *Engine) ProcessParse(ctx context.Context, resource string) (*model.Process, error) {
	return e.internal.ProcessParse(ctx, resource)
}

// ProcessSave 保存流程定义，返回流程ID。
func (e *Engine) ProcessSave(ctx context.Context, resource string, createUserID string) (int, error) {
	return e.internal.ProcessSave(ctx, resource, createUserID)
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
func (e *Engine) InstanceStart(ctx context.Context, procID int, businessID, comment, variablesJSON string) (int, error) {
	return e.internal.InstanceStart(ctx, procID, businessID, comment, variablesJSON)
}

// InstanceRevoke 撤销流程实例。
func (e *Engine) InstanceRevoke(ctx context.Context, instID int, force bool, revokeUserID string) error {
	return e.internal.InstanceRevoke(ctx, instID, force, revokeUserID)
}

// GetInstanceInfo 获取流程实例信息。
func (e *Engine) GetInstanceInfo(ctx context.Context, instID int) (model.InstanceView, error) {
	return e.internal.GetInstanceInfo(ctx, instID)
}

// GetInstanceStartByUser 获取用户发起的流程实例列表（含分页和总数）。
func (e *Engine) GetInstanceStartByUser(ctx context.Context, userID, processName string, pageNo, pageSize int) (*model.PageData[model.InstanceView], error) {
	return e.internal.GetInstanceStartByUser(ctx, userID, processName, pageNo, pageSize)
}

// --- 任务 ---

// TaskPass 任务通过。
func (e *Engine) TaskPass(ctx context.Context, taskID int, comment, varJSON string, directlyToRejected bool) error {
	return e.internal.TaskPass(ctx, taskID, comment, varJSON, directlyToRejected)
}

// TaskReject 任务驳回。
func (e *Engine) TaskReject(ctx context.Context, taskID int, comment, varJSON string) error {
	return e.internal.TaskReject(ctx, taskID, comment, varJSON)
}

// TaskTransfer 任务转交。
func (e *Engine) TaskTransfer(ctx context.Context, taskID int, users []string) error {
	return e.internal.TaskTransfer(ctx, taskID, users)
}

// TaskFreeReject 自由驳回（驳回到上游指定节点）。
func (e *Engine) TaskFreeReject(ctx context.Context, taskID int, nodeID, comment, varJSON string) error {
	return e.internal.TaskFreeReject(ctx, taskID, nodeID, comment, varJSON)
}

// GetTaskInfo 获取任务信息。
func (e *Engine) GetTaskInfo(ctx context.Context, taskID int) (model.TaskView, error) {
	return e.internal.GetTaskInfo(ctx, taskID)
}

// GetTaskToDoList 获取待办任务列表（含分页和总数）。
func (e *Engine) GetTaskToDoList(ctx context.Context, userID, processName string, asc bool, pageNo, pageSize int) (*model.PageData[model.TaskView], error) {
	return e.internal.GetTaskToDoList(ctx, userID, processName, asc, pageNo, pageSize)
}

// GetTaskFinishedList 获取已办任务列表（含分页和总数）。
func (e *Engine) GetTaskFinishedList(ctx context.Context, userID, processName string, ignoreStartByMe, asc bool, pageNo, pageSize int) (*model.PageData[model.TaskView], error) {
	return e.internal.GetTaskFinishedList(ctx, userID, processName, ignoreStartByMe, asc, pageNo, pageSize)
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

// --- 计划任务 ---

// ScheduleTask 注册定时任务。
func (e *Engine) ScheduleTask(ctx context.Context, name string, startAt, stopAt time.Time, intervalSec int64, fn func() error) error {
	return e.internal.ScheduleTask(ctx, name, startAt, stopAt, intervalSec, fn)
}

// GetScheduledTaskList 获取已注册的定时任务列表。
func (e *Engine) GetScheduledTaskList(ctx context.Context) map[string]*service.SchedulerTask {
	return e.internal.GetScheduledTaskList(ctx)
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
