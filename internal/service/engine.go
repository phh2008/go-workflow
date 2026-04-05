package service

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/Bunny3th/easy-workflow/internal/entity"
	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/repository"
	"gorm.io/gorm"
)

// Config 引擎配置。
type Config struct {
	IgnoreEventError bool         // 是否忽略事件错误
	Logger           *slog.Logger // 日志记录器，为 nil 时使用默认日志
}

// eventMethod 事件方法包装，记录方法所在的 receiver 和反射方法信息。
type eventMethod struct {
	receiver any
	method   any // reflect.Method - 使用 any 以避免在结构体定义中引入 reflect 依赖
}

// Engine 工作流引擎，封装所有状态和业务逻辑。
type Engine struct {
	db               *gorm.DB
	logger           *slog.Logger
	repo             repository.Repository
	eventPool        map[string]*eventMethod
	eventPoolMu      sync.RWMutex
	ignoreEventErr   bool
	procCache        map[int]map[string]model.Node
	procCacheMu      sync.RWMutex
	expressionEval   *ExpressionEvaluator
}

// NewEngine 创建并初始化工作流引擎。
// db 由调用方创建，NewEngine 负责执行 AutoMigrate 和初始化内部组件。
func NewEngine(db *gorm.DB, cfg Config) (*Engine, error) {
	log := cfg.Logger
	if log == nil {
		log = slog.Default()
	}

	// 执行 AutoMigrate
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("初始化数据表失败: %w", err)
	}

	eng := &Engine{
		db:             db,
		logger:         log,
		repo:           repository.NewFlowRepo(db),
		eventPool:      make(map[string]*eventMethod),
		ignoreEventErr: cfg.IgnoreEventError,
		procCache:      make(map[int]map[string]model.Node),
		expressionEval: NewExpressionEvaluator(),
	}

	log.Info("easy workflow 引擎初始化完成")
	return eng, nil
}

// Close 关闭引擎。数据库连接由调用方负责关闭。
func (e *Engine) Close() error {
	return nil
}

// DB 返回底层的 GORM 数据库实例，供需要直接操作数据库的场景使用。
func (e *Engine) DB() *gorm.DB {
	return e.db
}

// autoMigrate 执行所有实体表的自动迁移。
func autoMigrate(db *gorm.DB) error {
	tables := []any{
		&entity.ProcDef{},
		&entity.HistProcDef{},
		&entity.ProcInst{},
		&entity.HistProcInst{},
		&entity.ProcTask{},
		&entity.HistProcTask{},
		&entity.ProcExecution{},
		&entity.HistProcExecution{},
		&entity.ProcInstVariable{},
		&entity.HistProcInstVariable{},
	}
	for _, t := range tables {
		if err := db.AutoMigrate(t); err != nil {
			return err
		}
	}
	return nil
}
