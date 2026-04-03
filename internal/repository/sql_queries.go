package repository

// 流程定义相关 SQL

const sqlGetProcessIDByName = "SELECT id, version FROM proc_def WHERE name = ? AND source = ?"

const sqlGetProcessResource = "SELECT resource FROM proc_def WHERE id = ?"

const sqlListProcessDef = "SELECT * FROM proc_def WHERE source = ?"

// 流程实例相关 SQL

// sqlGetInstanceInfo 通过 CTE 合并 proc_inst 和 hist_proc_inst 获取流程实例信息
const sqlGetInstanceInfo = "" +
	"WITH tmp_procinst AS (\n" +
	"    SELECT id, proc_id, proc_version, business_id, starter, current_node_id,\n" +
	"           create_time, `status`\n" +
	"      FROM proc_inst\n" +
	"     WHERE id = ?\n" +
	"    UNION ALL\n" +
	"    SELECT proc_inst_id AS id, proc_id, proc_version, business_id, starter, current_node_id,\n" +
	"           create_time, `status`\n" +
	"      FROM hist_proc_inst\n" +
	"     WHERE proc_inst_id = ?\n" +
	")\n" +
	"SELECT a.id, a.proc_id, a.proc_version, a.business_id, a.starter,\n" +
	"       a.current_node_id, a.create_time, a.`status`, b.name\n" +
	"  FROM tmp_procinst a\n" +
	"  LEFT JOIN proc_def b ON a.proc_id = b.id"

// sqlListInstanceStartByUser 获取特定用户发起的流程实例列表（含分页）
const sqlListInstanceStartByUser = "" +
	"WITH tmp_procinst AS (\n" +
	"    SELECT id, proc_id, proc_version, business_id, starter, current_node_id,\n" +
	"           create_time, `status`\n" +
	"      FROM proc_inst\n" +
	"     WHERE CASE WHEN '' = @userid THEN TRUE ELSE starter = @userid END\n" +
	"    UNION ALL\n" +
	"    SELECT proc_inst_id AS id, proc_id, proc_version, business_id, starter, current_node_id,\n" +
	"           create_time, `status`\n" +
	"      FROM hist_proc_inst\n" +
	"     WHERE CASE WHEN '' = @userid THEN TRUE ELSE starter = @userid END\n" +
	")\n" +
	"SELECT a.id, a.proc_id, a.proc_version, a.business_id,\n" +
	"       a.starter, a.current_node_id, a.create_time, a.`status`, b.name\n" +
	"  FROM tmp_procinst a\n" +
	"  JOIN proc_def b ON a.proc_id = b.id\n" +
	" WHERE CASE WHEN '' = @procname THEN TRUE ELSE b.name = @procname END\n" +
	" ORDER BY a.id\n" +
	" LIMIT @index, @rows"

// sqlCountInstanceStartByUser 获取特定用户发起的流程实例总数
const sqlCountInstanceStartByUser = "" +
	"WITH tmp_procinst AS (\n" +
	"    SELECT id\n" +
	"      FROM proc_inst\n" +
	"     WHERE CASE WHEN '' = @userid THEN TRUE ELSE starter = @userid END\n" +
	"    UNION ALL\n" +
	"    SELECT proc_inst_id AS id\n" +
	"      FROM hist_proc_inst\n" +
	"     WHERE CASE WHEN '' = @userid THEN TRUE ELSE starter = @userid END\n" +
	")\n" +
	"SELECT COUNT(*)\n" +
	"  FROM tmp_procinst a\n" +
	"  JOIN proc_def b ON a.proc_id = b.id\n" +
	" WHERE CASE WHEN '' = @procname THEN TRUE ELSE b.name = @procname END"

const sqlGetProcessIDByInstID = "SELECT proc_id FROM proc_inst WHERE id = ?"

const sqlGetProcessNameByInstID = "" +
	"SELECT b.name\n" +
	"  FROM proc_inst a\n" +
	"  JOIN proc_def b ON a.proc_id = b.id\n" +
	" WHERE a.id = ?"

// 任务相关 SQL

// sqlGetTaskInfo 通过 CTE 合并 proc_task 和 hist_proc_task 获取任务信息
const sqlGetTaskInfo = "" +
	"WITH tmp_task AS (\n" +
	"    SELECT id, proc_id, proc_inst_id, business_id, starter, node_id, node_name, prev_node_id,\n" +
	"           is_cosigned, batch_code, user_id, `status`, is_finished, `comment`,\n" +
	"           proc_inst_create_time, create_time, finished_time\n" +
	"      FROM proc_task\n" +
	"     WHERE id = ?\n" +
	"    UNION ALL\n" +
	"    SELECT task_id AS id, proc_id, proc_inst_id, business_id, starter, node_id, node_name,\n" +
	"           prev_node_id, is_cosigned, batch_code, user_id, `status`, is_finished, `comment`,\n" +
	"           proc_inst_create_time, create_time, finished_time\n" +
	"      FROM hist_proc_task\n" +
	"     WHERE id = ?\n" +
	")\n" +
	"SELECT a.id, a.proc_id, b.name, a.proc_inst_id, a.business_id, a.starter,\n" +
	"       a.node_id, a.node_name, a.prev_node_id, a.is_cosigned,\n" +
	"       a.batch_code, a.user_id, a.`status`, a.is_finished, a.`comment`,\n" +
	"       a.proc_inst_create_time, a.create_time, a.finished_time\n" +
	"  FROM tmp_task a\n" +
	"  LEFT JOIN proc_def b ON a.proc_id = b.id"

// sqlGetTaskToDoList 获取特定用户待办任务列表（含可选过滤和分页）
const sqlGetTaskToDoList = "" +
	"SELECT a.id, a.proc_id, b.name, a.proc_inst_id,\n" +
	"       a.business_id, a.starter, a.node_id, a.node_name, a.prev_node_id,\n" +
	"       a.is_cosigned, a.batch_code, a.user_id, a.`status`, a.is_finished, a.`comment`,\n" +
	"       a.proc_inst_create_time, a.create_time, a.finished_time\n" +
	"  FROM proc_task a\n" +
	"  JOIN proc_def b ON a.proc_id = b.id\n" +
	" WHERE CASE WHEN '' = @userid THEN TRUE ELSE a.user_id = @userid END\n" +
	"   AND a.is_finished = 0\n" +
	"   AND CASE WHEN '' = @procname THEN TRUE ELSE b.name = @procname END\n" +
	" ORDER BY a.id\n" +
	"   {{SORT}}\n" +
	" LIMIT @index, @rows"

// sqlCountTaskToDo 获取特定用户待办任务总数
const sqlCountTaskToDo = "" +
	"SELECT COUNT(*)\n" +
	"  FROM proc_task a\n" +
	"  JOIN proc_def b ON a.proc_id = b.id\n" +
	" WHERE CASE WHEN '' = @userid THEN TRUE ELSE a.user_id = @userid END\n" +
	"   AND a.is_finished = 0\n" +
	"   AND CASE WHEN '' = @procname THEN TRUE ELSE b.name = @procname END"

// sqlGetTaskFinishedList 获取特定用户已完成任务列表（CTE 合并历史数据）
const sqlGetTaskFinishedList = "" +
	"WITH tmp_task AS (\n" +
	"    SELECT id, proc_id, proc_inst_id, business_id, starter, node_id, node_name, prev_node_id,\n" +
	"           is_cosigned, batch_code, user_id, `status`, is_finished, `comment`,\n" +
	"           proc_inst_create_time, create_time, finished_time\n" +
	"      FROM proc_task\n" +
	"     WHERE CASE WHEN '' = @userid THEN TRUE ELSE user_id = @userid END\n" +
	"    UNION ALL\n" +
	"    SELECT task_id AS id, proc_id, proc_inst_id, business_id, starter, node_id, node_name,\n" +
	"           prev_node_id, is_cosigned, batch_code, user_id, `status`, is_finished, `comment`,\n" +
	"           proc_inst_create_time, create_time, finished_time\n" +
	"      FROM hist_proc_task\n" +
	"     WHERE CASE WHEN '' = @userid THEN TRUE ELSE user_id = @userid END\n" +
	")\n" +
	"SELECT a.id, a.proc_id, b.name, a.proc_inst_id, a.business_id, a.starter,\n" +
	"       a.node_id, a.node_name, a.prev_node_id, a.is_cosigned,\n" +
	"       a.batch_code, a.user_id, a.`status`, a.is_finished, a.`comment`,\n" +
	"       a.proc_inst_create_time, a.create_time, a.finished_time\n" +
	"  FROM tmp_task a\n" +
	"  JOIN proc_def b ON a.proc_id = b.id\n" +
	" WHERE a.is_finished = 1\n" +
	"   AND a.`status` != 0\n" +
	"   AND CASE WHEN '' = @procname THEN TRUE ELSE b.name = @procname END\n" +
	"   AND CASE WHEN true = @ignorestartbyme THEN a.starter != @userid ELSE TRUE END\n" +
	" ORDER BY a.finished_time\n" +
	"   {{SORT}}\n" +
	" LIMIT @index, @rows"

// sqlCountTaskFinished 获取特定用户已完成任务总数
const sqlCountTaskFinished = "" +
	"WITH tmp_task AS (\n" +
	"    SELECT id, proc_id, proc_inst_id, business_id, starter, node_id, node_name, prev_node_id,\n" +
	"           is_cosigned, batch_code, user_id, `status`, is_finished, `comment`,\n" +
	"           proc_inst_create_time, create_time, finished_time\n" +
	"      FROM proc_task\n" +
	"     WHERE CASE WHEN '' = @userid THEN TRUE ELSE user_id = @userid END\n" +
	"    UNION ALL\n" +
	"    SELECT task_id AS id, proc_id, proc_inst_id, business_id, starter, node_id, node_name,\n" +
	"           prev_node_id, is_cosigned, batch_code, user_id, `status`, is_finished, `comment`,\n" +
	"           proc_inst_create_time, create_time, finished_time\n" +
	"      FROM hist_proc_task\n" +
	"     WHERE CASE WHEN '' = @userid THEN TRUE ELSE user_id = @userid END\n" +
	")\n" +
	"SELECT COUNT(*)\n" +
	"  FROM tmp_task a\n" +
	"  JOIN proc_def b ON a.proc_id = b.id\n" +
	" WHERE a.is_finished = 1\n" +
	"   AND a.`status` != 0\n" +
	"   AND CASE WHEN '' = @procname THEN TRUE ELSE b.name = @procname END\n" +
	"   AND CASE WHEN true = @ignorestartbyme THEN a.starter != @userid ELSE TRUE END"

// sqlGetInstanceTaskHistory 获取流程实例下任务历史记录（CTE 合并历史数据）
const sqlGetInstanceTaskHistory = "" +
	"WITH tmp_task AS (\n" +
	"    SELECT id, proc_id, proc_inst_id, business_id, starter, node_id, node_name, prev_node_id,\n" +
	"           is_cosigned, batch_code, user_id, `status`, is_finished, `comment`,\n" +
	"           proc_inst_create_time, create_time, finished_time\n" +
	"      FROM proc_task\n" +
	"     WHERE proc_inst_id = ?\n" +
	"    UNION ALL\n" +
	"    SELECT task_id AS id, proc_id, proc_inst_id, business_id, starter, node_id, node_name,\n" +
	"           prev_node_id, is_cosigned, batch_code, user_id, `status`, is_finished, `comment`,\n" +
	"           proc_inst_create_time, create_time, finished_time\n" +
	"      FROM hist_proc_task\n" +
	"     WHERE proc_inst_id = ?\n" +
	")\n" +
	"SELECT a.id, a.proc_id, b.name, a.proc_inst_id, a.business_id, a.starter,\n" +
	"       a.node_id, a.node_name, a.prev_node_id, a.is_cosigned,\n" +
	"       a.batch_code, a.user_id, a.`status`, a.is_finished, a.`comment`,\n" +
	"       a.proc_inst_create_time, a.create_time, a.finished_time\n" +
	"  FROM tmp_task a\n" +
	"  JOIN proc_def b ON a.proc_id = b.id\n" +
	" ORDER BY a.id"

// sqlTaskNodeStatus 获取任务节点审批状态（总任务数、通过数、驳回数）
const sqlTaskNodeStatus = "" +
	"SELECT COUNT(*) AS total_task,\n" +
	"       SUM(CASE `status` WHEN 1 THEN 1 ELSE 0 END) AS total_passed,\n" +
	"       SUM(CASE `status` WHEN 2 THEN 1 ELSE 0 END) AS total_rejected\n" +
	"  FROM proc_task\n" +
	" WHERE proc_inst_id = ?\n" +
	"   AND node_id = ?\n" +
	"   AND batch_code = ?"

const sqlGetNotFinishUsers = "" +
	"SELECT DISTINCT user_id\n" +
	"  FROM proc_task\n" +
	" WHERE proc_inst_id = ? AND node_id = ? AND is_finished = 0"

const sqlGetPrevNodeBatchCode = "" +
	"SELECT a.batch_code\n" +
	"  FROM proc_task a\n" +
	"  JOIN (\n" +
	"      SELECT prev_node_id, proc_inst_id FROM proc_task WHERE id = ?\n" +
	"  ) b ON a.node_id = b.prev_node_id AND a.proc_inst_id = b.proc_inst_id\n" +
	" ORDER BY a.id DESC\n" +
	" LIMIT 1"

const sqlHasRejectInBatch = "SELECT id FROM proc_task WHERE batch_code = ? AND `status` = 2 LIMIT 1"

const sqlTaskRevoke = "" +
	"UPDATE proc_task\n" +
	"   SET `status` = 0, is_finished = 0, finished_time = NULL, comment = NULL\n" +
	" WHERE id = ?"

const sqlTaskRevokeBatch = "" +
	"UPDATE proc_task SET is_finished = 0\n" +
	" WHERE `status` = 0\n" +
	"   AND batch_code IN (SELECT batch_code FROM proc_task WHERE id = ?)"

// 执行关系相关 SQL

const sqlGetStartNodeID = "SELECT node_id FROM proc_execution WHERE proc_id = ? AND node_type = 0"

const sqlIsNodeFinished = "" +
	"SELECT CASE WHEN total = finished THEN 1 ELSE 0 END AS finished\n" +
	"  FROM (\n" +
	"      SELECT COUNT(*) AS total, SUM(is_finished) AS finished\n" +
	"        FROM proc_task\n" +
	"       WHERE proc_inst_id = ? AND node_id = ?\n" +
	"       GROUP BY proc_inst_id, node_id\n" +
	"  ) a"

const sqlGetUpstreamNodes = "" +
	"WITH RECURSIVE tmp(node_id, node_name, prev_node_id, node_type) AS (\n" +
	"    SELECT node_id, node_name, prev_node_id, node_type\n" +
	"      FROM proc_execution\n" +
	"     WHERE node_id = ?\n" +
	"    UNION ALL\n" +
	"    SELECT a.node_id, a.node_name, a.prev_node_id, a.node_type\n" +
	"      FROM proc_execution a\n" +
	"      JOIN tmp b ON a.node_id = b.prev_node_id\n" +
	")\n" +
	"SELECT DISTINCT node_id, node_name, prev_node_id, node_type\n" +
	"  FROM tmp\n" +
	" WHERE node_type != 2 AND node_id != ?"

// 变量相关 SQL

const sqlGetVariable = "" +
	"SELECT `value`\n" +
	"  FROM proc_inst_variable\n" +
	" WHERE proc_inst_id = ? AND `key` = ?\n" +
	" LIMIT 1"

const sqlGetVariableForUpsert = "" +
	"SELECT *\n" +
	"  FROM proc_inst_variable\n" +
	" WHERE proc_inst_id = ? AND `key` = ?\n" +
	" ORDER BY id\n" +
	" LIMIT 1"

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
	"    proc_inst_create_time, create_time, finished_time\n" +
	")\n" +
	"SELECT id, proc_id, proc_inst_id, business_id, starter,\n" +
	"       node_id, node_name, prev_node_id, is_cosigned, batch_code,\n" +
	"       user_id, `status`, is_finished, `comment`,\n" +
	"       proc_inst_create_time, create_time, finished_time\n" +
	"  FROM proc_task\n" +
	" WHERE proc_inst_id = ?"

const sqlDeleteTasks = "DELETE FROM proc_task WHERE proc_inst_id = ?"

const sqlUpdateInstanceStatus = "UPDATE proc_inst SET `status` = ? WHERE id = ?"

const sqlArchiveInstance = "" +
	"INSERT INTO hist_proc_inst (\n" +
	"    proc_inst_id, proc_id, proc_version, business_id,\n" +
	"    starter, current_node_id, create_time, `status`\n" +
	")\n" +
	"SELECT id, proc_id, proc_version, business_id,\n" +
	"       starter, current_node_id, create_time, `status`\n" +
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
	"INSERT INTO hist_proc_def (proc_id, name, `version`, resource, user_id, source, create_time)\n" +
	"SELECT id, name, `version`, resource, user_id, source, create_time\n" +
	"  FROM proc_def\n" +
	" WHERE name = ? AND source = ?"

const sqlArchiveExecutions = "" +
	"INSERT INTO hist_proc_execution (\n" +
	"    proc_id, proc_version, node_id, node_name,\n" +
	"    prev_node_id, node_type, is_cosigned, create_time\n" +
	")\n" +
	"SELECT proc_id, proc_version, node_id, node_name,\n" +
	"       prev_node_id, node_type, is_cosigned, create_time\n" +
	"  FROM proc_execution\n" +
	" WHERE proc_id = ?"
