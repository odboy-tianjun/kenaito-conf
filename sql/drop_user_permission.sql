-- 删除已废弃的用户应用授权表。
-- 权限现由 application.owner_ids（应用负责人）和业务组角色管理。
-- 已有环境升级时手动执行：mysql -u<user> -p <database> < sql/drop_user_permission.sql

DROP TABLE IF EXISTS `user_permission`;
