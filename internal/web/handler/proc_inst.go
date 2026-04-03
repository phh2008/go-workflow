package handler

import (
	"net/http"

	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/service"
	"github.com/gin-gonic/gin"
)

// ProcInstHandler 流程实例处理器。
type ProcInstHandler struct {
	engine *service.Engine
}

// NewProcInstHandler 创建流程实例处理器。
func NewProcInstHandler(eng *service.Engine) *ProcInstHandler {
	return &ProcInstHandler{engine: eng}
}

// @Summary      开始流程
// @Description  注意，VariablesJson格式是key-value对象集合:[{"Key":"starter","Value":"U0001"}]
// @Tags         流程实例
// @Produce      json
// @Param        processId  formData int  true  "流程ID" example(1)
// @Param        businessId  formData string  true  "业务ID" example(订单001)
// @Param        comment  formData string  false  "评论意见" example("家中有事请假三天,请领导批准")
// @Param        variablesJson  formData string  false  "变量(Json)" example([{"Key":"starter","Value":"U0001"}])
// @Success      200  {object}  int 流程实例ID
// @Failure      400  {object}  string 报错信息
// @Router       /inst/start [post]
func (h *ProcInstHandler) Start(c *gin.Context) {
	var req model.InstanceStartReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	id, err := h.engine.InstanceStart(c.Request.Context(), req.ProcessID, req.BusinessID, req.Comment, req.VariablesJSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, id)
}

// @Summary      撤销流程
// @Description  注意，Force 是否强制撤销，若为false,则只有流程回到发起人这里才能撤销
// @Tags         流程实例
// @Produce      json
// @Param        instanceId  formData int  true  "流程实例ID" example(1)
// @Param        revokeUserId  formData string  true  "撤销发起用户ID" example("U001")
// @Param        force  formData bool  true  "是否强制撤销" example(false)
// @Success      200  {object}  string "ok"
// @Failure      400  {object}  string 报错信息
// @Router       /inst/revoke [post]
func (h *ProcInstHandler) Revoke(c *gin.Context) {
	var req model.InstanceRevokeReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if err := h.engine.InstanceRevoke(c.Request.Context(), req.InstanceID, req.Force, req.RevokeUserID); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, "ok")
}

// @Summary      流程实例中任务执行记录
// @Description
// @Tags         流程实例
// @Produce      json
// @Param        instid  query int  true  "流程实例ID" example(1)
// @Success      200  {object}  []model.TaskView "任务列表"
// @Failure      400  {object}  string 报错信息
// @Router       /inst/task_history [get]
func (h *ProcInstHandler) TaskHistory(c *gin.Context) {
	var req model.InstanceTaskHistoryReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	tasklist, err := h.engine.GetInstanceTaskHistory(c.Request.Context(), req.InstanceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, tasklist)
}

// @Summary      获取起始人为特定用户的流程实例
// @Description
// @Tags         流程实例
// @Produce      json
// @Param        pageNo   query int  false  "页码" example(1)
// @Param        pageSize query int  false  "每页数量" example(10)
// @Param        userId   query string  false  "用户ID 传入空则获取所有用户的流程实例" example("U001")
// @Param        processName query string  false  "指定流程名称，非必填" example("请假")
// @Success      200  {object}  []model.InstanceView "流程实例列表"
// @Failure      400  {object}  string 报错信息
// @Router       /inst/start/by [get]
func (h *ProcInstHandler) StartByUser(c *gin.Context) {
	var req model.InstanceListReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.engine.GetInstanceStartByUser(c.Request.Context(), req.UserID, req.ProcessName, req.GetPageNo(), req.GetPageSize())
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}
