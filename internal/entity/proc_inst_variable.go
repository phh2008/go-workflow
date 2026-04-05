package entity

// ProcInstVariable 流程实例变量表，存储流程实例运行过程中的变量。
type ProcInstVariable struct {
	ID         int    `gorm:"primaryKey;column:id;type:INT UNSIGNED NOT NULL AUTO_INCREMENT;"`
	ProcInstID int    `gorm:"index:ix_proc_inst_id;column:proc_inst_id;type:INT UNSIGNED NOT NULL;comment:流程实例ID"`
	Key        string `gorm:"column:key;type:VARCHAR(250) NOT NULL;comment:变量key"`
	Value      string `gorm:"column:value;type:VARCHAR(250) NOT NULL;comment:变量value"`
}

func (ProcInstVariable) TableName() string {
	return "proc_inst_variable"
}

// HistProcInstVariable 流程实例变量历史表，流程结束时数据归档到此表。
type HistProcInstVariable struct {
	ID         int    `gorm:"primaryKey;column:id;type:INT UNSIGNED NOT NULL AUTO_INCREMENT;"`
	ProcInstID int    `gorm:"index:ix_proc_inst_id;column:proc_inst_id;type:INT UNSIGNED NOT NULL;comment:流程实例ID"`
	Key        string `gorm:"column:key;type:VARCHAR(250) NOT NULL;comment:变量key"`
	Value      string `gorm:"column:value;type:VARCHAR(250) NOT NULL;comment:变量value"`
}

func (HistProcInstVariable) TableName() string {
	return "hist_proc_inst_variable"
}
