-- =====================================================
-- ToughRadius v9 数据库迁移脚本 (升级)
-- 功能: 移除 TR069/CWMP 相关表和字段
-- 版本: v8 -> v9
-- 作者: ToughRADIUS Team
-- 日期: 2025-11-08
-- =====================================================

-- 开启事务
BEGIN;

-- 设置客户端编码
SET client_encoding = 'UTF8';

-- 记录迁移开始
DO $$
BEGIN
    RAISE NOTICE '开始执行 ToughRadius v9 迁移脚本...';
    RAISE NOTICE '当前时间: %', NOW();
END $$;

-- =====================================================
-- 1. 备份关键数据 (可选，用于回滚)
-- =====================================================

-- 创建临时备份表 (如果需要保留数据)
DO $$
BEGIN
    -- 备份 CPE 设备的 CWMP 字段数据
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'net_cpe') THEN
        RAISE NOTICE '备份 net_cpe 表的 CWMP 相关字段...';
        
        CREATE TABLE IF NOT EXISTS _backup_v8_net_cpe_cwmp AS
        SELECT 
            id,
            name,
            sn,
            cwmp_url,
            cwmp_username,
            cwmp_password,
            cwmp_status,
            cwmp_last_inform,
            created_at,
            updated_at
        FROM net_cpe
        WHERE cwmp_url IS NOT NULL OR cwmp_status IS NOT NULL;
        
        RAISE NOTICE '备份完成，共 % 条记录', (SELECT COUNT(*) FROM _backup_v8_net_cpe_cwmp);
    END IF;
END $$;

-- =====================================================
-- 2. 删除 CWMP 相关表
-- =====================================================

-- 2.1 删除 CWMP 预设任务表
DROP TABLE IF EXISTS cwmp_preset_task CASCADE;
RAISE NOTICE '已删除表: cwmp_preset_task';

-- 2.2 删除 CWMP 预设表
DROP TABLE IF EXISTS cwmp_preset CASCADE;
RAISE NOTICE '已删除表: cwmp_preset';

-- 2.3 删除 CWMP 固件配置表
DROP TABLE IF EXISTS cwmp_firmware_config CASCADE;
RAISE NOTICE '已删除表: cwmp_firmware_config';

-- 2.4 删除 CWMP 恢复出厂设置表
DROP TABLE IF EXISTS cwmp_factory_reset CASCADE;
RAISE NOTICE '已删除表: cwmp_factory_reset';

-- 2.5 删除 CWMP 配置会话表
DROP TABLE IF EXISTS cwmp_config_session CASCADE;
RAISE NOTICE '已删除表: cwmp_config_session';

-- 2.6 删除 CWMP 配置表
DROP TABLE IF EXISTS cwmp_config CASCADE;
RAISE NOTICE '已删除表: cwmp_config';

-- =====================================================
-- 3. 修改 net_cpe 表 (移除 CWMP 字段)
-- =====================================================

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'net_cpe') THEN
        RAISE NOTICE '开始修改 net_cpe 表，移除 CWMP 相关字段...';
        
        -- 删除 cwmp_url 字段
        IF EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'net_cpe' AND column_name = 'cwmp_url'
        ) THEN
            ALTER TABLE net_cpe DROP COLUMN cwmp_url;
            RAISE NOTICE '已删除字段: net_cpe.cwmp_url';
        END IF;
        
        -- 删除 cwmp_username 字段
        IF EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'net_cpe' AND column_name = 'cwmp_username'
        ) THEN
            ALTER TABLE net_cpe DROP COLUMN cwmp_username;
            RAISE NOTICE '已删除字段: net_cpe.cwmp_username';
        END IF;
        
        -- 删除 cwmp_password 字段
        IF EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'net_cpe' AND column_name = 'cwmp_password'
        ) THEN
            ALTER TABLE net_cpe DROP COLUMN cwmp_password;
            RAISE NOTICE '已删除字段: net_cpe.cwmp_password';
        END IF;
        
        -- 删除 cwmp_status 字段
        IF EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'net_cpe' AND column_name = 'cwmp_status'
        ) THEN
            ALTER TABLE net_cpe DROP COLUMN cwmp_status;
            RAISE NOTICE '已删除字段: net_cpe.cwmp_status';
        END IF;
        
        -- 删除 cwmp_last_inform 字段
        IF EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'net_cpe' AND column_name = 'cwmp_last_inform'
        ) THEN
            ALTER TABLE net_cpe DROP COLUMN cwmp_last_inform;
            RAISE NOTICE '已删除字段: net_cpe.cwmp_last_inform';
        END IF;
        
        RAISE NOTICE 'net_cpe 表修改完成';
    ELSE
        RAISE NOTICE '警告: net_cpe 表不存在，跳过字段删除';
    END IF;
END $$;

-- =====================================================
-- 4. 清理相关索引 (如果存在)
-- =====================================================

DO $$
BEGIN
    RAISE NOTICE '检查并清理 CWMP 相关索引...';
    
    -- 删除可能存在的 CWMP 相关索引
    DROP INDEX IF EXISTS idx_cwmp_config_sn;
    DROP INDEX IF EXISTS idx_cwmp_config_session_sn;
    DROP INDEX IF EXISTS idx_cwmp_preset_name;
    DROP INDEX IF EXISTS idx_cwmp_preset_task_status;
    DROP INDEX IF EXISTS idx_net_cpe_cwmp_status;
    
    RAISE NOTICE '索引清理完成';
END $$;

-- =====================================================
-- 5. 清理相关序列 (如果存在)
-- =====================================================

DO $$
BEGIN
    RAISE NOTICE '检查并清理 CWMP 相关序列...';
    
    DROP SEQUENCE IF EXISTS cwmp_config_id_seq CASCADE;
    DROP SEQUENCE IF EXISTS cwmp_config_session_id_seq CASCADE;
    DROP SEQUENCE IF EXISTS cwmp_preset_id_seq CASCADE;
    DROP SEQUENCE IF EXISTS cwmp_preset_task_id_seq CASCADE;
    DROP SEQUENCE IF EXISTS cwmp_factory_reset_id_seq CASCADE;
    DROP SEQUENCE IF EXISTS cwmp_firmware_config_id_seq CASCADE;
    
    RAISE NOTICE '序列清理完成';
END $$;

-- =====================================================
-- 6. 创建迁移记录表
-- =====================================================

CREATE TABLE IF NOT EXISTS schema_migrations (
    id SERIAL PRIMARY KEY,
    version VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN DEFAULT TRUE
);

-- 记录本次迁移
INSERT INTO schema_migrations (version, description, success)
VALUES (
    'v9.0.0-remove-tr069',
    'Remove TR069/CWMP module: tables, fields, indexes, sequences',
    TRUE
)
ON CONFLICT (version) DO UPDATE
SET applied_at = CURRENT_TIMESTAMP,
    success = TRUE;

RAISE NOTICE '迁移记录已保存到 schema_migrations 表';

-- =====================================================
-- 7. 统计信息
-- =====================================================

DO $$
DECLARE
    backup_count INTEGER;
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE '迁移统计信息:';
    RAISE NOTICE '========================================';
    RAISE NOTICE '已删除表: 6 个 (cwmp_config, cwmp_config_session, cwmp_preset, cwmp_preset_task, cwmp_factory_reset, cwmp_firmware_config)';
    RAISE NOTICE '已修改表: 1 个 (net_cpe - 移除 5 个 CWMP 字段)';
    
    SELECT COUNT(*) INTO backup_count FROM _backup_v8_net_cpe_cwmp WHERE 1=1;
    RAISE NOTICE '备份数据: % 条 CPE CWMP 记录 (表: _backup_v8_net_cpe_cwmp)', backup_count;
    
    RAISE NOTICE '========================================';
    RAISE NOTICE '迁移完成时间: %', NOW();
    RAISE NOTICE '========================================';
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE '警告: 备份表可能不存在或为空';
END $$;

-- =====================================================
-- 8. 提交事务
-- =====================================================

COMMIT;

-- 最终提示
DO $$
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '✅ ToughRadius v9 数据库迁移成功完成！';
    RAISE NOTICE '';
    RAISE NOTICE '后续步骤:';
    RAISE NOTICE '1. 运行 v9_migration_verify.sql 验证迁移结果';
    RAISE NOTICE '2. 如需回滚，运行 v9_migration_down.sql';
    RAISE NOTICE '3. 备份数据保存在 _backup_v8_net_cpe_cwmp 表中';
    RAISE NOTICE '';
END $$;
