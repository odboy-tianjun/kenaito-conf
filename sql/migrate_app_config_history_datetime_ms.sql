-- 将配置历史时间字段升级为毫秒精度，与 app_config 保持一致。
-- 新库由 xorm Sync2 按模型定义建表，已有环境执行本脚本升级。

ALTER TABLE `app_config_history`
  MODIFY COLUMN `created_at` DATETIME(3) NULL DEFAULT NULL;
