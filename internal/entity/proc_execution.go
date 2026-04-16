package entity

// ProcExecution 流程节点执行关系表，记录流程中各节点的执行路径和关系。
type ProcExecution struct {
	ID          int       `gorm:"primaryKey;column:id;type:INT UNSIGNED NOT NULL AUTO_INCREMENT;"`
	ProcID      int       `gorm:"column:proc_id;type:INT UNSIGNED NOT NULL;index:idx_proc_id;comment:流程ID"`
	ProcVersion int       `gorm:"column:proc_version;type:INT UNSIGNED NOT NULL;comment:流程版本号"`
	NodeID      string    `gorm:"column:node_id;type:VARCHAR(200) NOT NULL;comment:节点ID"`
	NodeName    string    `gorm:"column:node_name;type:VARCHAR(200) NOT NULL;comment:节点名称"`
	PrevNodeID  string    `gorm:"column:prev_node_id;type:VARCHAR(200);default:NULL;comment:上级节点ID"`
	NodeType    int       `gorm:"column:node_type;type:TINYINT NOT NULL;comment:流程类型 0:开始节点 1:任务节点 2:网关节点 3:结束节点"`
	IsCosigned  int       `gorm:"column:is_cosigned;type:TINYINT NOT NULL;comment:是否会签"`
	CreatedAt   LocalTime `gorm:"column:created_at;type:DATETIME;default:NOW();comment:创建时间"`
}

func (ProcExecution) TableName() string {
	return "proc_execution"
}

// HistProcExecution 流程节点执行关系历史表，流程结束时数据归档到此表。
type HistProcExecution struct {
	ID          int       `gorm:"primaryKey;column:id;type:INT UNSIGNED NOT NULL AUTO_INCREMENT;"`
	ProcID      int       `gorm:"column:proc_id;type:INT UNSIGNED NOT NULL;index:idx_proc_id;comment:流程ID"`
	ProcVersion int       `gorm:"column:proc_version;type:INT UNSIGNED NOT NULL;comment:流程版本号"`
	NodeID      string    `gorm:"column:node_id;type:VARCHAR(200) NOT NULL;comment:节点ID"`
	NodeName    string    `gorm:"column:node_name;type:VARCHAR(200) NOT NULL;comment:节点名称"`
	PrevNodeID  string    `gorm:"column:prev_node_id;type:VARCHAR(200);default:NULL;comment:上级节点ID"`
	NodeType    int       `gorm:"column:node_type;type:TINYINT NOT NULL;comment:流程类型 0:开始节点 1:任务节点 2:网关节点 3:结束节点"`
	IsCosigned  int       `gorm:"column:is_cosigned;type:TINYINT NOT NULL;comment:是否会签"`
	CreatedAt   LocalTime `gorm:"column:created_at;type:DATETIME;default:NOW();comment:创建时间"`
}

func (HistProcExecution) TableName() string {
	return "hist_proc_execution"
}
