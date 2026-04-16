package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	easyworkflow "github.com/Bunny3th/easy-workflow"
	"github.com/Bunny3th/easy-workflow/example/event"
	"github.com/Bunny3th/easy-workflow/example/process"
	"github.com/Bunny3th/easy-workflow/example/schedule"
	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func main() {
	// 启用 Debug 级别日志
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))

	//----------------------------创建数据库连接----------------------------
	db, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:3306)/easy_workflow_v2?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
		Logger:         gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		slog.Error("数据库连接失败", "error", err)
		return
	}

	//----------------------------开启流程引擎----------------------------
	eng, err := easyworkflow.New(db, easyworkflow.Config{})
	if err != nil {
		slog.Error("引擎初始化失败", "error", err)
		return
	}

	// 注册事件
	evt := event.NewMyEvent(eng)
	eng.RegisterNodeEvent("MyEvent_End", evt.OnEnd)
	eng.RegisterNodeEvent("MyEvent_Notify", evt.OnNotify)
	eng.RegisterNodeEvent("MyEvent_ResolveRoles", evt.OnResolveRoles)
	eng.RegisterNodeEvent("MyEvent_TaskForceNodePass", evt.OnTaskForceNodePass)
	eng.RegisterProcEvent("MyEvent_Revoke", evt.OnRevoke)

	//----------------------------生成一个示例流程----------------------------
	procID := process.CreateExampleProcess(eng)

	//----------------------------自动演示：走完一个完整流程----------------------------
	if procID > 0 {
		runDemo(eng, procID)
	}

	//----------------------------注册定时任务----------------------------
	s, err := gocron.NewScheduler()
	if err != nil {
		slog.Error("创建调度器失败", "error", err)
		return
	}
	// 每10秒钟执行一次自动完成任务(免审)
	_, err = s.NewJob(
		gocron.DurationJob(10*time.Second),
		gocron.NewTask(schedule.AutoFinishTask(eng)),
	)
	if err != nil {
		slog.Error("注册定时任务失败", "error", err)
		return
	}
	s.Start()

	//----------------------------开启web api----------------------------
	ginEngine := gin.New()
	ginEngine.Use(gin.Logger())
	ginEngine.Use(gin.Recovery())
	if err := eng.StartWebAPI(ginEngine, easyworkflow.WebConfig{
		BaseURL:     "/process",
		ShowSwagger: true,
		SwaggerURL:  "/swagger/*any",
		Addr:        ":8180",
	}); err != nil {
		slog.Error("启动 Web API 失败", "error", err)
		os.Exit(1)
	}
}

// runDemo 自动演示一个完整的请假流程（路径B：长请假 >= 3天）。
//
// 流程路径：Start → GW-Day($days>=3) → Manager(张经理) → GW-Parallel
//
//	→ HR(人事老刘) + DeputyBoss(会签: 赵总、钱总、孙总，2人通过即通过)
//	→ GW-Parallel2 → Boss(李老板) → END
func runDemo(eng *easyworkflow.Engine, procID int) {
	ctx := context.Background()
	slog.Info("========== 开始自动演示流程 ==========")

	// 1. 启动流程实例
	variablesJSON := `[{"key":"starter","value":"张三"},{"key":"days","value":"5"}]`
	instID, err := eng.InstanceStart(ctx, model.InstanceStartReq{
		ProcessID:     procID,
		BusinessID:    "biz_demo_001",
		Comment:       "张三申请5天请假",
		VariablesJSON: variablesJSON,
	})
	if err != nil {
		slog.Error("启动流程实例失败", "error", err)
		return
	}
	slog.Info("流程实例已启动", "实例ID", instID)

	// 查询当前所有待办任务，确认流程走到了哪个节点
	allTasks, err := eng.GetTaskToDoList(ctx, model.TaskListReq{PageQuery: model.PageQuery{PageNo: 1, PageSize: 20}})
	if err != nil {
		slog.Error("查询待办任务失败", "error", err)
		return
	}
	if len(allTasks.Data) == 0 {
		slog.Info("当前无待办任务，流程可能已走到结束节点")
		return
	}
	for _, t := range allTasks.Data {
		slog.Info("待办任务", "任务ID", t.TaskID, "审批人", t.UserID, "节点", t.NodeName)
	}

	// 2. 按顺序模拟各节点审批
	// 流程经过 Start(自动) → GW-Day(条件判断) → Manager 节点，现在 Manager 有待办任务
	steps := []struct {
		user    string
		comment string
	}{
		{"张经理", "主管同意请假"},
		{"人事老刘", "人事同意请假"},
		{"赵总", "副总同意请假（1/2）"},
		{"钱总", "副总同意请假（2/2，触发自动通过孙总）"},
		{"李老板", "老板同意请假"},
	}

	for _, step := range steps {
		taskID, err := getFirstToDoTaskID(eng, ctx, step.user)
		if err != nil {
			slog.Error("查询待办任务失败", "error", err, "审批人", step.user)
			return
		}
		if taskID == 0 {
			slog.Error("未找到待办任务", "审批人", step.user)
			// 打印当前所有待办帮助定位
			allTasks, _ := eng.GetTaskToDoList(ctx, model.TaskListReq{PageQuery: model.PageQuery{PageNo: 1, PageSize: 20}})
			if len(allTasks.Data) > 0 {
				slog.Info("当前存在的待办任务:")
				for _, t := range allTasks.Data {
					slog.Info("  ", "任务ID", t.TaskID, "审批人", t.UserID, "节点", t.NodeName)
				}
			}
			return
		}

		err = eng.TaskPass(ctx, model.TaskActionReq{
			TaskID:  taskID,
			Comment: step.comment,
		}, false)
		if err != nil {
			slog.Error("审批任务失败", "error", err, "审批人", step.user, "任务ID", taskID)
			return
		}
		slog.Info("任务已通过", "审批人", step.user, "任务ID", taskID, "备注", step.comment)
	}

	// 3. 验证流程是否完成
	instInfo, err := eng.GetInstanceInfo(ctx, instID)
	if err != nil {
		slog.Error("查询流程实例信息失败", "error", err)
		return
	}
	if instInfo.Status == 2 {
		slog.Info("========== 流程演示完成，流程已正常结束 ==========")
	} else {
		slog.Info("流程当前状态", "实例ID", instID, "状态", fmt.Sprintf("code=%d", instInfo.Status))
	}
}

// getFirstToDoTaskID 查询指定用户的第一个待办任务ID。
func getFirstToDoTaskID(eng *easyworkflow.Engine, ctx context.Context, userID string) (int, error) {
	result, err := eng.GetTaskToDoList(ctx, model.TaskListReq{
		UserID:    userID,
		PageQuery: model.PageQuery{PageNo: 1, PageSize: 1},
	})
	if err != nil {
		return 0, err
	}
	if len(result.Data) == 0 {
		return 0, nil
	}
	return result.Data[0].TaskID, nil
}
