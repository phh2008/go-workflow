package entity

// ProcTask 任务表，记录流程中每个待处理或已处理的任务。
type ProcTask struct {
	BaseModel
	ProcID             int       `gorm:"column:proc_id;type:INT UNSIGNED NOT NULL;index:idx_proc_id;comment:流程ID,冗余字段"`
	ProcInstID         int       `gorm:"column:proc_inst_id;type:INT UNSIGNED NOT NULL;index:idx_proc_inst_id;comment:流程实例ID"`
	BusinessID         string    `gorm:"column:business_id;type:VARCHAR(200);default:NULL;comment:业务ID,冗余字段"`
	Starter            string    `gorm:"column:starter;type:VARCHAR(200) NOT NULL;index:idx_starter;comment:流程发起人用户ID,冗余字段"`
	NodeID             string    `gorm:"column:node_id;type:VARCHAR(200) NOT NULL;comment:节点ID"`
	NodeName           string    `gorm:"column:node_name;type:VARCHAR(200) NOT NULL;comment:节点名称"`
	PrevNodeID         string    `gorm:"column:prev_node_id;type:VARCHAR(200);default:NULL;comment:上个处理节点ID,注意这里和execution中的上一个节点不一样，这里是实际审批处理时上个已处理节点的ID"`
	IsCosigned         int       `gorm:"column:is_cosigned;type:TINYINT;default:0;comment:0:任意一人通过即可 1:会签"`
	BatchCode          string    `gorm:"column:batch_code;type:VARCHAR(50);default:NULL;index:idx_batch_code;comment:批次码.节点会被驳回，一个节点可能产生多批task,用此码做分别"`
	UserID             string    `gorm:"column:user_id;type:VARCHAR(200) NOT NULL;index:idx_user_id;comment:分配用户ID"`
	Status             int       `gorm:"column:status;type:TINYINT;default:0;comment:任务状态:0:初始 1:通过 2:驳回"`
	IsFinished         int       `gorm:"column:is_finished;type:TINYINT;default:0;comment:0:任务未完成 1:处理完成.任务未必都是用户处理的，比如会签时一人驳回，其他任务系统自动设为已处理"`
	Comment            string    `gorm:"column:comment;type:TEXT;default:NULL;comment:任务备注"`
	ProcInstCreateTime LocalTime `gorm:"column:proc_inst_create_time;type:DATETIME NOT NULL;comment:流程实例创建时间,冗余字段"`
	FinishedTime       LocalTime `gorm:"column:finished_time;type:DATETIME;default:NULL;index:idx_finished_time;comment:处理任务时间"`
}

func (ProcTask) TableName() string {
	return "proc_task"
}

// HistProcTask 任务历史表，流程结束时数据归档到此表。
type HistProcTask struct {
	BaseModel
	TaskID             int       `gorm:"column:task_id;type:INT UNSIGNED NOT NULL;index:idx_task_id;comment:任务ID"`
	ProcID             int       `gorm:"column:proc_id;type:INT UNSIGNED NOT NULL;index:idx_proc_id;comment:流程ID,冗余字段"`
	ProcInstID         int       `gorm:"column:proc_inst_id;type:INT UNSIGNED NOT NULL;index:idx_proc_inst_id;comment:流程实例ID"`
	BusinessID         string    `gorm:"column:business_id;type:VARCHAR(200);default:NULL;comment:业务ID,冗余字段"`
	Starter            string    `gorm:"column:starter;type:VARCHAR(200) NOT NULL;index:idx_starter;comment:流程发起人用户ID,冗余字段"`
	NodeID             string    `gorm:"column:node_id;type:VARCHAR(200) NOT NULL;comment:节点ID"`
	NodeName           string    `gorm:"column:node_name;type:VARCHAR(200) NOT NULL;comment:节点名称"`
	PrevNodeID         string    `gorm:"column:prev_node_id;type:VARCHAR(200);default:NULL;comment:上个处理节点ID,注意这里和execution中的上一个节点不一样，这里是实际审批处理时上个已处理节点的ID"`
	IsCosigned         int       `gorm:"column:is_cosigned;type:TINYINT;default:0;comment:0:任意一人通过即可 1:会签"`
	BatchCode          string    `gorm:"column:batch_code;type:VARCHAR(50);default:NULL;index:idx_batch_code;comment:批次码.节点会被驳回，一个节点可能产生多批task,用此码做分别"`
	UserID             string    `gorm:"column:user_id;type:VARCHAR(200) NOT NULL;index:idx_user_id;comment:分配用户ID"`
	Status             int       `gorm:"column:status;type:TINYINT;default:0;comment:任务状态:0:初始 1:通过 2:驳回"`
	IsFinished         int       `gorm:"column:is_finished;type:TINYINT;default:0;comment:0:任务未完成 1:处理完成.任务未必都是用户处理的，比如会签时一人驳回，其他任务系统自动设为已处理"`
	Comment            string    `gorm:"column:comment;type:TEXT;default:NULL;comment:任务备注"`
	ProcInstCreateTime LocalTime `gorm:"column:proc_inst_create_time;type:DATETIME NOT NULL;comment:流程实例创建时间,冗余字段"`
	FinishedTime       LocalTime `gorm:"column:finished_time;type:DATETIME;default:NULL;index:idx_finished_time;comment:处理任务时间"`
}

func (HistProcTask) TableName() string {
	return "hist_proc_task"
}
