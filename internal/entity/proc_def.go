package entity

import "gorm.io/plugin/soft_delete"

// BaseModel 公共基础结构体，包含审计字段和软删除标记，业务表和历史表共用。
type BaseModel struct {
	ID        int                   `gorm:"primaryKey;column:id;type:INT UNSIGNED NOT NULL AUTO_INCREMENT;"`
	CreatedAt LocalTime             `gorm:"column:created_at;type:DATETIME DEFAULT NOW();default:NOW();comment:创建时间"`
	UpdateAt  LocalTime             `gorm:"column:update_time;type:DATETIME DEFAULT NOW();default:NOW();comment:更新时间"`
	CreatedBy string                `gorm:"column:created_by;type:VARCHAR(250) DEFAULT '';default:'';comment:创建人"`
	UpdatedBy string                `gorm:"column:updated_by;type:VARCHAR(250) DEFAULT '';default:'';comment:更新人"`
	Deleted   soft_delete.DeletedAt `gorm:"softDelete:flag;default:0;comment:是否删除:0-否,1-是"`
}

// ProcDef 流程定义表，存储流程的定义信息。
type ProcDef struct {
	BaseModel
	Name     string `gorm:"column:name;type:VARCHAR(250) NOT NULL;comment:流程名字;uniqueIndex:uix_name_source"`
	Version  int    `gorm:"column:version;type:INT UNSIGNED NOT NULL DEFAULT 1;default:1;comment:版本号"`
	Resource string `gorm:"column:resource;type:TEXT NOT NULL;comment:流程定义模板"`
	Source   string `gorm:"column:source;type:VARCHAR(250) NOT NULL;uniqueIndex:uix_name_source;comment:来源(引擎可能被多个系统、组件等使用，这里记下从哪个来源创建的流程);"`
}

func (ProcDef) TableName() string {
	return "proc_def"
}

// HistProcDef 流程定义历史表，流程结束时数据归档到此表。
type HistProcDef struct {
	BaseModel
	ProcID   int    `gorm:"column:proc_id;type:INT UNSIGNED NOT NULL;comment:流程ID"`
	Name     string `gorm:"column:name;type:VARCHAR(250) NOT NULL;comment:流程名字;"`
	Version  int    `gorm:"column:version;type:INT UNSIGNED NOT NULL DEFAULT 1;default:1;comment:版本号"`
	Resource string `gorm:"column:resource;type:TEXT NOT NULL;comment:流程定义模板"`
	Source   string `gorm:"column:source;type:VARCHAR(250) NOT NULL;comment:来源(引擎可能被多个系统、组件等使用，这里记下从哪个来源创建的流程);"`
}

func (HistProcDef) TableName() string {
	return "hist_proc_def"
}
