package handler

import (
	"net/http"

	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/service"
	"github.com/gin-gonic/gin"
)

// ProcDefHandler 流程定义处理器。
type ProcDefHandler struct {
	engine *service.Engine
}

// NewProcDefHandler 创建流程定义处理器。
func NewProcDefHandler(eng *service.Engine) *ProcDefHandler {
	return &ProcDefHandler{engine: eng}
}

// @Summary      流程定义保存/升级
// @Description
// @Tags         流程定义
// @Produce      json
// @Param        Resource  formData string  true  "流程定义资源(json)" example(json字符串)
// @Param        CreateUserID  formData string  true  "创建者ID" example(0001)
// @Success      200  {object}  int 流程ID
// @Failure      400  {object}  string 报错信息
// @Router       /def/save [post]
func (h *ProcDefHandler) Save(c *gin.Context) {
	var req model.ProcessSaveReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	procID, err := h.engine.ProcessSave(c.Request.Context(), req.Resource, req.CreateUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, procID)
}

// @Summary      获取特定source下所有流程
// @Description  引擎可能被多个系统、组件等使用，source表示从哪个来源创建的流程
// @Tags         流程定义
// @Produce      json
// @Param        source  query string  true  "来源" example(办公系统)
// @Success      200  {object}  []entity.ProcDef 流程定义列表
// @Failure      400  {object}  string 报错信息
// @Router       /def/list [get]
func (h *ProcDefHandler) ListBySource(c *gin.Context) {
	var req model.ProcessListReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	procDef, err := h.engine.GetProcessList(c.Request.Context(), req.Source)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, procDef)
}

// @Summary      获取流程定义
// @Description  返回的是Node数组，流程是由N个Node组成的
// @Tags         流程定义
// @Produce      json
// @Param        id  query int  true  "流程ID" example(1)
// @Success      200  {object}  model.Process "流程定义"
// @Failure      400  {object}  string 报错信息
// @Router       /def/get [get]
func (h *ProcDefHandler) GetByID(c *gin.Context) {
	var req model.ProcessDefGetReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	nodes, err := h.engine.GetProcessDefine(c.Request.Context(), req.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, nodes)
}
