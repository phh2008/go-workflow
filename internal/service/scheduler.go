package service

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/Bunny3th/easy-workflow/internal/model"
)

// SchedulerTask 被计划的任务。
type SchedulerTask struct {
	StartAt        time.Time     // 任务开始时间
	StopAt         time.Time     // 任务结束时间
	IntervalSecond int64         // 重复执行间隔（秒），最小1秒
	Func           func() error  // 需要运行的任务方法
	LastRunTime    time.Time     // 上一次运行时间
	LastResult     string        // 上一次运行结果
	LastDuration   time.Duration // 上一次运行时长
}

// ScheduleTask 登记计划任务，任务会被添加进计划任务池，并进入运行状态。
// 注意：请使用 go 关键字运行此函数，因为任务运行周期可能很长。
func (e *Engine) ScheduleTask(ctx context.Context, params model.ScheduleTaskParams) error {
	_ = ctx
	e.scheduledTasksMu.Lock()
	defer e.scheduledTasksMu.Unlock()

	if _, ok := e.scheduledTasks[params.Name]; ok {
		return errors.New("此任务已被加入任务池，无需重复操作")
	}

	now := time.Now()

	if now.After(params.StopAt) {
		return errors.New("任务结束时间小于当前时间，任务不会被运行")
	}

	if params.IntervalSec < 1 {
		return errors.New("重复执行间隔最小1秒")
	}

	if params.StopAt.Before(params.StartAt) {
		return errors.New("开始时间应小于结束时间")
	}

	e.scheduledTasks[params.Name] = &SchedulerTask{
		StartAt:        params.StartAt,
		StopAt:         params.StopAt,
		IntervalSecond: params.IntervalSec,
		Func:           params.Func,
	}

	go e.runScheduledTask(params.Name)

	return nil
}

// GetScheduledTaskList 获取任务计划池中的所有任务信息。
func (e *Engine) GetScheduledTaskList(ctx context.Context) map[string]*SchedulerTask {
	_ = ctx
	e.scheduledTasksMu.RLock()
	defer e.scheduledTasksMu.RUnlock()
	// 返回副本
	result := make(map[string]*SchedulerTask, len(e.scheduledTasks))
	maps.Copy(result, e.scheduledTasks)
	return result
}

// runScheduledTask 运行任务计划。
func (e *Engine) runScheduledTask(name string) {

	e.scheduledTasksMu.RLock()
	task, ok := e.scheduledTasks[name]
	e.scheduledTasksMu.RUnlock()

	if !ok {
		return
	}

	now := time.Now()

	if now.After(task.StopAt) {
		return
	}

	var waitDuration time.Duration
	if now.After(task.StartAt) {
		waitDuration = 0
	} else {
		waitDuration = task.StartAt.Sub(now)
	}

	timer := time.NewTimer(waitDuration)
	defer timer.Stop()

	<-timer.C

	ticker := time.NewTicker(time.Duration(task.IntervalSecond) * time.Second)
	defer ticker.Stop()

	defer func() {
		if err := recover(); err != nil {
			task.LastResult = fmt.Sprint(err)
			task.LastDuration = time.Since(task.LastRunTime)
		}
	}()

	for {
		<-ticker.C

		e.scheduledTasksMu.RLock()
		_, ok := e.scheduledTasks[name]
		e.scheduledTasksMu.RUnlock()
		if !ok {
			return
		}

		if time.Now().After(task.StopAt) {
			return
		}

		task.LastRunTime = time.Now()

		err := task.Func()
		if err == nil {
			task.LastResult = "ok"
		} else {
			task.LastResult = err.Error()
		}
		task.LastDuration = time.Since(task.LastRunTime)
	}
}
