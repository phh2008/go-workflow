package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/Bunny3th/easy-workflow"
	"github.com/Bunny3th/easy-workflow/example/event"
	"github.com/Bunny3th/easy-workflow/example/process"
	"github.com/Bunny3th/easy-workflow/example/schedule"
	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/gin-gonic/gin"
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
	eng.RegisterEvents(&event.MyEvent{})

	//----------------------------生成一个示例流程----------------------------
	process.CreateExampleProcess(eng)

	// 开启工作流计划任务：每10秒钟执行一次自动完成任务(免审)
	start, _ := time.ParseInLocation("2006-01-02 15:04:05", "2023-10-27 00:00:00", time.Local)
	end, _ := time.ParseInLocation("2006-01-02 15:04:05", "2199-10-27 00:00:00", time.Local)
	go eng.ScheduleTask(context.Background(), model.ScheduleTaskParams{
		Name:        "自动完成任务",
		StartAt:     start,
		StopAt:      end,
		IntervalSec: 10,
		Func:        schedule.AutoFinishTask(eng),
	})

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
