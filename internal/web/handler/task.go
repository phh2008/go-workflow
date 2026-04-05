package handler

import (
	"net/http"
	"strings"

	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/service"
	"github.com/gin-gonic/gin"
)

// TaskHandler 任务处理器。
type TaskHandler struct {
	engine *service.Engine
}

// NewTaskHandler 创建任务处理器。
func NewTaskHandler(eng *service.Engine) *TaskHandler {
	return &TaskHandler{engine: eng}
}

// @Summary      任务通过
// @Description  任务通过后根据流程定义，进入下一个节点进行处理
// @Tags         任务
// @Produce      json
// @Param        taskId  formData int  true  "任务ID" example(1)
// @Param        comment  formData string  false  "评论意见" example("同意请假")
// @Param        variableJson  formData string  false  "变量(Json)" example({"User":"001"})
// @Success      200  {object}  string "ok"
// @Failure      400  {object}  string 报错信息
// @Router       /task/pass [post]
func (h *TaskHandler) Pass(c *gin.Context) {
	var req model.TaskActionReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if err := h.engine.TaskPass(c.Request.Context(), model.TaskPassParams{
		TaskID:             req.TaskID,
		Comment:            req.Comment,
		VariableJSON:       req.VariableJSON,
		DirectlyToRejected: false,
	}); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, "ok")
}

// @Summary      任务通过后流程直接返回到上一个驳回我的节点
// @Description  此功能只有在非会签节点时才能使用
// @Tags         任务
// @Produce      json
// @Param        taskId  formData int  true  "任务ID" example(1)
// @Param        comment  formData string  false  "评论意见" example("同意请假")
// @Param        variableJson  formData string  false  "变量(Json)" example({"User":"001"})
// @Success      200  {object}  string "ok"
// @Failure      400  {object}  string 报错信息
// @Router       /task/pass/directly [post]
func (h *TaskHandler) PassDirectly(c *gin.Context) {
	var req model.TaskActionReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if err := h.engine.TaskPass(c.Request.Context(), model.TaskPassParams{
		TaskID:             req.TaskID,
		Comment:            req.Comment,
		VariableJSON:       req.VariableJSON,
		DirectlyToRejected: true,
	}); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, "ok")
}

// @Summary      任务驳回
// @Description
// @Tags         任务
// @Produce      json
// @Param        taskId  formData int  true  "任务ID" example(1)
// @Param        comment  formData string  false  "评论意见" example("不同意")
// @Param        variableJson  formData string  false  "变量(Json)" example({"User":"001"})
// @Success      200  {object}  string "ok"
// @Failure      400  {object}  string 报错信息
// @Router       /task/reject [post]
func (h *TaskHandler) Reject(c *gin.Context) {
	var req model.TaskActionReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if err := h.engine.TaskReject(c.Request.Context(), model.TaskRejectParams{
		TaskID:       req.TaskID,
		Comment:      req.Comment,
		VariableJSON: req.VariableJSON,
	}); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, "ok")
}

// @Summary      自由任务驳回
// @Description  驳回到上游任意一个节点
// @Tags         任务
// @Produce      json
// @Param        taskId  formData int  true  "任务ID" example(1)
// @Param        comment  formData string  false  "评论意见" example("不同意")
// @Param        variableJson  formData string  false  "变量(Json)" example({"User":"001"})
// @Param        rejectToNodeId  formData string  false  "驳回到哪个节点" example("流程开始节点")
// @Success      200  {object}  string "ok"
// @Failure      400  {object}  string 报错信息
// @Router       /task/reject/free [post]
func (h *TaskHandler) FreeReject(c *gin.Context) {
	var req model.TaskFreeRejectReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if err := h.engine.TaskFreeReject(c.Request.Context(), model.TaskFreeRejectParams{
		TaskID:         req.TaskID,
		RejectToNodeID: req.RejectToNodeID,
		Comment:        req.Comment,
		VariableJSON:   req.VariableJSON,
	}); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, "ok")
}

// @Summary      将任务转交给他人处理
// @Description
// @Tags         任务
// @Produce      json
// @Param        taskId  formData int  true  "任务ID" example(1)
// @Param        users  formData string  true  "用户,多个用户使用逗号分隔" example("U001,U002,U003")
// @Success      200  {object}  string "ok"
// @Failure      400  {object}  string 报错信息
// @Router       /task/transfer [post]
func (h *TaskHandler) Transfer(c *gin.Context) {
	var req model.TaskTransferReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	// 支持逗号分隔和 JSON 数组两种格式
	users := req.Users
	if len(users) == 1 && strings.Contains(users[0], ",") {
		users = strings.Split(users[0], ",")
	}
	if err := h.engine.TaskTransfer(c.Request.Context(), model.TaskTransferParams{
		TaskID: req.TaskID,
		Users:  users,
	}); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, "ok")
}

// @Summary      获取待办任务
// @Description  返回的是任务数组
// @Tags         任务
// @Produce      json
// @Param        pageNo   query int  false  "页码" example(1)
// @Param        pageSize query int  false  "每页数量" example(10)
// @Param        userId   query string  false  "用户ID 传入空则获取所有用户的待办任务" example("U001")
// @Param        processName query string  false  "指定流程名称，非必填" example("请假")
// @Param        asc     query bool  false  "是否按照任务生成时间升序排列" example(true)
// @Success      200  {object}  []model.TaskView 任务数组
// @Failure      400  {object}  string 报错信息
// @Router       /task/todo [get]
func (h *TaskHandler) ToDoList(c *gin.Context) {
	var req model.TaskListReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.engine.GetTaskToDoList(c.Request.Context(), model.TaskToDoListParams{
		UserID:      req.UserID,
		ProcessName: req.ProcessName,
		Asc:         req.Asc,
		PageNo:      req.GetPageNo(),
		PageSize:    req.GetPageSize(),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary      获取已办任务
// @Description  返回的是任务数组
// @Tags         任务
// @Produce      json
// @Param        pageNo   query int  false  "页码" example(1)
// @Param        pageSize query int  false  "每页数量" example(10)
// @Param        userId   query string  false  "用户ID 传入空则获取所有用户的已完成任务(此时IgnoreStartByMe参数强制为False)" example("U001")
// @Param        processName query string  false  "指定流程名称，非必填" example("请假")
// @Param        ignoreStartByMe query bool  false  "忽略由我开启流程,而生成处理人是我自己的任务" example(true)
// @Param        asc     query bool  false  "是否按照任务完成时间升序排列" example(true)
// @Success      200  {object}  []model.TaskView 任务数组
// @Failure      400  {object}  string 报错信息
// @Router       /task/finished [get]
func (h *TaskHandler) FinishedList(c *gin.Context) {
	var req model.TaskFinishedListReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.engine.GetTaskFinishedList(c.Request.Context(), model.TaskFinishedListParams{
		UserID:          req.UserID,
		ProcessName:     req.ProcessName,
		IgnoreStartByMe: req.IgnoreStartByMe,
		Asc:             req.Asc,
		PageNo:          req.GetPageNo(),
		PageSize:        req.GetPageSize(),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary      获取本任务所在节点的所有上游节点
// @Description  此功能为自由驳回使用
// @Tags         任务
// @Produce      json
// @Param        taskid  query int  true  "任务ID" example(8)
// @Success      200  {object}  []model.Node 节点任务数组
// @Failure      400  {object}  string 报错信息
// @Router       /task/upstream [get]
func (h *TaskHandler) UpstreamNodeList(c *gin.Context) {
	var req model.TaskInfoReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	nodes, err := h.engine.TaskUpstreamNodeList(c.Request.Context(), req.TaskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, nodes)
}

// @Summary      当前任务可以执行哪些操作
// @Description  前端无法提前知道当前任务可以做哪些操作，此方法目的是解决这个困扰
// @Tags         任务
// @Produce      json
// @Param        taskid  query int  true  "任务ID" example(1)
// @Success      200  {object}  model.TaskAction "可执行任务"
// @Failure      400  {object}  string 报错信息
// @Router       /task/action [get]
func (h *TaskHandler) WhatCanIDo(c *gin.Context) {
	var req model.TaskInfoReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	action, err := h.engine.WhatCanIDo(c.Request.Context(), req.TaskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, action)
}

// @Summary      任务信息
// @Description
// @Tags         任务
// @Produce      json
// @Param        taskid  query int  true  "任务ID" example(1)
// @Success      200  {object}  model.TaskView "任务信息"
// @Failure      400  {object}  string 报错信息
// @Router       /task/info [get]
func (h *TaskHandler) Info(c *gin.Context) {
	var req model.TaskInfoReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	taskInfo, err := h.engine.GetTaskInfo(c.Request.Context(), req.TaskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, taskInfo)
}
