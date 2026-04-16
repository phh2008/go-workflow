package repository

// 归档相关 SQL

const sqlArchiveFinishTasks = "" +
	"UPDATE proc_task\n" +
	"   SET is_finished = 1, finished_time = NOW()\n" +
	" WHERE proc_inst_id = ? AND is_finished = 0"

const sqlArchiveTasks = "" +
	"INSERT INTO hist_proc_task (\n" +
	"    task_id, proc_id, proc_inst_id, business_id, starter,\n" +
	"    node_id, node_name, prev_node_id, is_cosigned, batch_code,\n" +
	"    user_id, `status`, is_finished, `comment`,\n" +
	"    proc_inst_create_time, created_at, update_time, created_by, updated_by, deleted, finished_time\n" +
	")\n" +
	"SELECT id, proc_id, proc_inst_id, business_id, starter,\n" +
	"       node_id, node_name, prev_node_id, is_cosigned, batch_code,\n" +
	"       user_id, `status`, is_finished, `comment`,\n" +
	"       proc_inst_create_time, created_at, update_time, created_by, updated_by, deleted, finished_time\n" +
	"  FROM proc_task\n" +
	" WHERE proc_inst_id = ?"

const sqlDeleteTasks = "DELETE FROM proc_task WHERE proc_inst_id = ?"

const sqlUpdateInstanceStatus = "UPDATE proc_inst SET `status` = ? WHERE id = ?"

const sqlArchiveInstance = "" +
	"INSERT INTO hist_proc_inst (\n" +
	"    proc_inst_id, proc_id, proc_version, business_id,\n" +
	"    starter, current_node_id, created_at, update_time, created_by, updated_by, deleted, `status`\n" +
	")\n" +
	"SELECT id, proc_id, proc_version, business_id,\n" +
	"       starter, current_node_id, created_at, update_time, created_by, updated_by, deleted, `status`\n" +
	"  FROM proc_inst\n" +
	" WHERE id = ?"

const sqlDeleteInstance = "DELETE FROM proc_inst WHERE id = ?"

const sqlArchiveVariables = "" +
	"INSERT INTO hist_proc_inst_variable (proc_inst_id, `key`, `value`)\n" +
	"SELECT proc_inst_id, `key`, `value`\n" +
	"  FROM proc_inst_variable\n" +
	" WHERE proc_inst_id = ?"

const sqlDeleteVariables = "DELETE FROM proc_inst_variable WHERE proc_inst_id = ?"

// 流程定义保存相关 SQL

const sqlArchiveProcDef = "" +
	"INSERT INTO hist_proc_def (proc_id, name, `version`, resource, created_by, source, created_at, update_time, updated_by, deleted)\n" +
	"SELECT id, name, `version`, resource, created_by, source, created_at, update_time, updated_by, deleted\n" +
	"  FROM proc_def\n" +
	" WHERE name = ? AND source = ?"

const sqlArchiveExecutions = "" +
	"INSERT INTO hist_proc_execution (\n" +
	"    proc_id, proc_version, node_id, node_name,\n" +
	"    prev_node_id, node_type, is_cosigned, created_at\n" +
	")\n" +
	"SELECT proc_id, proc_version, node_id, node_name,\n" +
	"       prev_node_id, node_type, is_cosigned, created_at\n" +
	"  FROM proc_execution\n" +
	" WHERE proc_id = ?"
