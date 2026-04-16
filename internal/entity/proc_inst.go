package entity

// ProcInst 流程实例表，记录每个流程的运行实例信息。
type ProcInst struct {
	BaseModel
	ProcID        int    `gorm:"column:proc_id;type:INT NOT NULL;index:idx_proc_id;comment:流程ID"`
	ProcVersion   int    `gorm:"column:proc_version;type:INT UNSIGNED NOT NULL;comment:流程版本号"`
	BusinessID    string `gorm:"column:business_id;type:VARCHAR(200);default:NULL;comment:业务ID"`
	Starter       string `gorm:"column:starter;type:VARCHAR(200) NOT NULL;index:idx_starter;comment:流程发起人用户ID"`
	CurrentNodeID string `gorm:"column:current_node_id;type:VARCHAR(200) NOT NULL;comment:当前进行节点ID"`
	Status        int    `gorm:"column:status;type:TINYINT;default:0;comment:0:未完成(审批中) 1:已完成(通过) 2:撤销"`
}

func (ProcInst) TableName() string {
	return "proc_inst"
}

// HistProcInst 流程实例历史表，流程结束时数据归档到此表。
type HistProcInst struct {
	BaseModel
	ProcInstID    int    `gorm:"column:proc_inst_id;type:INT UNSIGNED NOT NULL;index:idx_proc_inst_id;comment:流程实例ID"`
	ProcID        int    `gorm:"column:proc_id;type:INT NOT NULL;index:idx_proc_id;comment:流程ID"`
	ProcVersion   int    `gorm:"column:proc_version;type:INT UNSIGNED NOT NULL;comment:流程版本号"`
	BusinessID    string `gorm:"column:business_id;type:VARCHAR(200);default:NULL;comment:业务ID"`
	Starter       string `gorm:"column:starter;type:VARCHAR(200) NOT NULL;index:idx_starter;comment:流程发起人用户ID"`
	CurrentNodeID string `gorm:"column:current_node_id;type:VARCHAR(200) NOT NULL;comment:当前进行节点ID"`
	Status        int    `gorm:"column:status;type:TINYINT;default:0;comment:0:未完成(审批中) 1:已完成(通过) 2:撤销"`
}

func (HistProcInst) TableName() string {
	return "hist_proc_inst"
}
