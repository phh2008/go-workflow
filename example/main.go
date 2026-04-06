package main

import (
	"log/slog"
	"time"

	"github.com/Bunny3th/easy-workflow"
	"github.com/Bunny3th/easy-workflow/example/event"
	"github.com/Bunny3th/easy-workflow/example/process"
	"github.com/Bunny3th/easy-workflow/example/schedule"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func main() {
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
	process.CreateExampleProcess(eng)

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
	// 这里需要注意：如果你的业务系统也同时使用了swagger
	// 你希望业务系统的swagger页面与easy-workflow内置web api的swagger同时开启
	// 必须做到：
	// 1、业务swagger与工作流swagger必须使用同一个访问路由
	// 2、业务系统web api与easy-workflow内置web api必须使用不同的端口

	// 本项目采用gin运行web api，首先生成一个gin.Engine
	ginEngine := gin.New()
	// 这里定义中间件
	ginEngine.Use(gin.Logger())   // gin的默认log，默认输出是os.Stdout，即屏幕
	ginEngine.Use(gin.Recovery()) // 从任何panic中恢复，并在出现panic时返回http 500
	eng.StartWebAPI(ginEngine, easyworkflow.WebConfig{
		BaseURL:     "/process",
		ShowSwagger: true,
		SwaggerURL:  "/swagger/*any",
		Addr:        ":8180",
	})
}
