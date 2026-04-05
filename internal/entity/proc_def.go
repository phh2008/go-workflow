package entity

// CommonID 通用 ID 嵌入结构体，用于历史表的主键定义。
type CommonID struct {
	ID int `gorm:"primaryKey;column:id;type:INT UNSIGNED NOT NULL AUTO_INCREMENT;"`
}

// ProcDef 流程定义表，存储流程的定义信息。
type ProcDef struct {
	ID        int       `gorm:"primaryKey;column:id;type:INT UNSIGNED NOT NULL AUTO_INCREMENT;comment:流程ID"`
	Name      string    `gorm:"column:name;type:VARCHAR(250) NOT NULL;comment:流程名字;uniqueIndex:uix_name_source"`
	Version   int       `gorm:"column:version;type:INT UNSIGNED NOT NULL DEFAULT 1;default:1;comment:版本号"`
	Resource  string    `gorm:"column:resource;type:TEXT NOT NULL;comment:流程定义模板"`
	UserID    string    `gorm:"column:user_id;type:VARCHAR(250) NOT NULL;comment:创建者ID"`
	Source    string    `gorm:"column:source;type:VARCHAR(250) NOT NULL;uniqueIndex:uix_name_source;comment:来源(引擎可能被多个系统、组件等使用，这里记下从哪个来源创建的流程);"`
	CreatTime LocalTime `gorm:"column:create_time;type:DATETIME DEFAULT NOW();default:NOW();comment:创建时间"`
}

func (ProcDef) TableName() string {
	return "proc_def"
}

// HistProcDef 流程定义历史表，流程结束时数据归档到此表。
type HistProcDef struct {
	CommonID
	ProcID    int       `gorm:"column:proc_id;type:INT UNSIGNED NOT NULL;comment:流程ID"`
	Name      string    `gorm:"column:name;type:VARCHAR(250) NOT NULL;comment:流程名字;"`
	Version   int       `gorm:"column:version;type:INT UNSIGNED NOT NULL DEFAULT 1;default:1;comment:版本号"`
	Resource  string    `gorm:"column:resource;type:TEXT NOT NULL;comment:流程定义模板"`
	UserID    string    `gorm:"column:user_id;type:VARCHAR(250) NOT NULL;comment:创建者ID"`
	Source    string    `gorm:"column:source;type:VARCHAR(250) NOT NULL;comment:来源(引擎可能被多个系统、组件等使用，这里记下从哪个来源创建的流程);"`
	CreatTime LocalTime `gorm:"column:create_time;type:DATETIME DEFAULT NOW();default:NOW();comment:创建时间"`
}

func (HistProcDef) TableName() string {
	return "hist_proc_def"
}
