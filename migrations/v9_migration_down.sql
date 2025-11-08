-- =====================================================
-- ToughRadius v9 数据库迁移回滚脚本
-- 功能: 恢复 TR069/CWMP 相关表和字段
-- 版本: v9 -> v8
-- 作者: ToughRADIUS Team
-- 日期: 2025-11-08
-- 警告: 此脚本仅恢复表结构，数据需要从备份恢复
-- =====================================================

-- 开启事务
BEGIN;

-- 设置客户端编码
SET client_encoding = 'UTF8';

-- 记录回滚开始
DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE '开始执行 ToughRadius v9 -> v8 回滚脚本...';
    RAISE NOTICE '当前时间: %', NOW();
    RAISE NOTICE '警告: 此操作将恢复 TR069/CWMP 模块';
    RAISE NOTICE '========================================';
END $$;

-- =====================================================
-- 1. 恢复 CWMP 配置表
-- =====================================================

CREATE TABLE IF NOT EXISTS cwmp_config (
    id SERIAL PRIMARY KEY,
    sn VARCHAR(128) NOT NULL,
    category VARCHAR(32) NOT NULL DEFAULT '',
    param_path VARCHAR(256) NOT NULL,
    param_name VARCHAR(128) NOT NULL,
    param_value TEXT,
    param_type VARCHAR(32) DEFAULT 'string',
    status VARCHAR(32) DEFAULT 'pending',
    result TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(sn, param_path)
);

CREATE INDEX IF NOT EXISTS idx_cwmp_config_sn ON cwmp_config(sn);
CREATE INDEX IF NOT EXISTS idx_cwmp_config_status ON cwmp_config(status);

COMMENT ON TABLE cwmp_config IS 'CWMP 配置表';
COMMENT ON COLUMN cwmp_config.sn IS '设备序列号';
COMMENT ON COLUMN cwmp_config.category IS '配置分类';
COMMENT ON COLUMN cwmp_config.param_path IS '参数路径';
COMMENT ON COLUMN cwmp_config.param_name IS '参数名称';
COMMENT ON COLUMN cwmp_config.param_value IS '参数值';
COMMENT ON COLUMN cwmp_config.status IS '状态: pending, success, failed';

RAISE NOTICE '✓ 已恢复表: cwmp_config';

-- =====================================================
-- 2. 恢复 CWMP 配置会话表
-- =====================================================

CREATE TABLE IF NOT EXISTS cwmp_config_session (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(128) NOT NULL UNIQUE,
    sn VARCHAR(128) NOT NULL,
    source VARCHAR(32) DEFAULT 'manual',
    status VARCHAR(32) DEFAULT 'created',
    request_data TEXT,
    response_data TEXT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_cwmp_config_session_sn ON cwmp_config_session(sn);
CREATE INDEX IF NOT EXISTS idx_cwmp_config_session_status ON cwmp_config_session(status);

COMMENT ON TABLE cwmp_config_session IS 'CWMP 配置会话表';
COMMENT ON COLUMN cwmp_config_session.session_id IS '会话ID';
COMMENT ON COLUMN cwmp_config_session.sn IS '设备序列号';
COMMENT ON COLUMN cwmp_config_session.source IS '来源: manual, preset, scheduled';
COMMENT ON COLUMN cwmp_config_session.status IS '状态: created, processing, completed, failed';

RAISE NOTICE '✓ 已恢复表: cwmp_config_session';

-- =====================================================
-- 3. 恢复 CWMP 预设表
-- =====================================================

CREATE TABLE IF NOT EXISTS cwmp_preset (
    id SERIAL PRIMARY KEY,
    name VARCHAR(128) NOT NULL UNIQUE,
    description TEXT,
    content TEXT NOT NULL,
    category VARCHAR(32) DEFAULT 'config',
    enabled BOOLEAN DEFAULT TRUE,
    created_by VARCHAR(64),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_cwmp_preset_name ON cwmp_preset(name);
CREATE INDEX IF NOT EXISTS idx_cwmp_preset_category ON cwmp_preset(category);

COMMENT ON TABLE cwmp_preset IS 'CWMP 预设模板表';
COMMENT ON COLUMN cwmp_preset.name IS '预设名称';
COMMENT ON COLUMN cwmp_preset.content IS '预设内容 (JSON格式)';
COMMENT ON COLUMN cwmp_preset.category IS '分类: config, firmware, factory_reset';

RAISE NOTICE '✓ 已恢复表: cwmp_preset';

-- =====================================================
-- 4. 恢复 CWMP 预设任务表
-- =====================================================

CREATE TABLE IF NOT EXISTS cwmp_preset_task (
    id SERIAL PRIMARY KEY,
    task_id VARCHAR(128) NOT NULL UNIQUE,
    preset_id INTEGER NOT NULL,
    preset_name VARCHAR(128),
    sn VARCHAR(128) NOT NULL,
    status VARCHAR(32) DEFAULT 'pending',
    priority INTEGER DEFAULT 5,
    scheduled_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    result TEXT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (preset_id) REFERENCES cwmp_preset(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_cwmp_preset_task_sn ON cwmp_preset_task(sn);
CREATE INDEX IF NOT EXISTS idx_cwmp_preset_task_status ON cwmp_preset_task(status);
CREATE INDEX IF NOT EXISTS idx_cwmp_preset_task_preset_id ON cwmp_preset_task(preset_id);

COMMENT ON TABLE cwmp_preset_task IS 'CWMP 预设任务表';
COMMENT ON COLUMN cwmp_preset_task.task_id IS '任务ID';
COMMENT ON COLUMN cwmp_preset_task.preset_id IS '预设ID';
COMMENT ON COLUMN cwmp_preset_task.status IS '状态: pending, running, completed, failed, cancelled';
COMMENT ON COLUMN cwmp_preset_task.priority IS '优先级 (1-10)';

RAISE NOTICE '✓ 已恢复表: cwmp_preset_task';

-- =====================================================
-- 5. 恢复 CWMP 恢复出厂设置表
-- =====================================================

CREATE TABLE IF NOT EXISTS cwmp_factory_reset (
    id SERIAL PRIMARY KEY,
    sn VARCHAR(128) NOT NULL,
    status VARCHAR(32) DEFAULT 'pending',
    initiated_by VARCHAR(64),
    initiated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    result TEXT,
    error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_cwmp_factory_reset_sn ON cwmp_factory_reset(sn);
CREATE INDEX IF NOT EXISTS idx_cwmp_factory_reset_status ON cwmp_factory_reset(status);

COMMENT ON TABLE cwmp_factory_reset IS 'CWMP 恢复出厂设置记录表';
COMMENT ON COLUMN cwmp_factory_reset.sn IS '设备序列号';
COMMENT ON COLUMN cwmp_factory_reset.status IS '状态: pending, processing, completed, failed';

RAISE NOTICE '✓ 已恢复表: cwmp_factory_reset';

-- =====================================================
-- 6. 恢复 CWMP 固件配置表
-- =====================================================

CREATE TABLE IF NOT EXISTS cwmp_firmware_config (
    id SERIAL PRIMARY KEY,
    sn VARCHAR(128) NOT NULL,
    firmware_url VARCHAR(512) NOT NULL,
    firmware_version VARCHAR(64),
    file_type VARCHAR(32) DEFAULT '1 Firmware Upgrade Image',
    target_filename VARCHAR(256),
    username VARCHAR(128),
    password VARCHAR(128),
    status VARCHAR(32) DEFAULT 'pending',
    delay_seconds INTEGER DEFAULT 0,
    initiated_by VARCHAR(64),
    initiated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    result TEXT,
    error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_cwmp_firmware_config_sn ON cwmp_firmware_config(sn);
CREATE INDEX IF NOT EXISTS idx_cwmp_firmware_config_status ON cwmp_firmware_config(status);

COMMENT ON TABLE cwmp_firmware_config IS 'CWMP 固件配置记录表';
COMMENT ON COLUMN cwmp_firmware_config.sn IS '设备序列号';
COMMENT ON COLUMN cwmp_firmware_config.firmware_url IS '固件下载URL';
COMMENT ON COLUMN cwmp_firmware_config.status IS '状态: pending, processing, completed, failed';
COMMENT ON COLUMN cwmp_firmware_config.delay_seconds IS '延迟执行时间(秒)';

RAISE NOTICE '✓ 已恢复表: cwmp_firmware_config';

-- =====================================================
-- 7. 恢复 net_cpe 表的 CWMP 字段
-- =====================================================

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'net_cpe') THEN
        RAISE NOTICE '开始恢复 net_cpe 表的 CWMP 字段...';
        
        -- 添加 cwmp_url 字段
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'net_cpe' AND column_name = 'cwmp_url'
        ) THEN
            ALTER TABLE net_cpe ADD COLUMN cwmp_url VARCHAR(256);
            RAISE NOTICE '✓ 已添加字段: net_cpe.cwmp_url';
        END IF;
        
        -- 添加 cwmp_username 字段
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'net_cpe' AND column_name = 'cwmp_username'
        ) THEN
            ALTER TABLE net_cpe ADD COLUMN cwmp_username VARCHAR(128);
            RAISE NOTICE '✓ 已添加字段: net_cpe.cwmp_username';
        END IF;
        
        -- 添加 cwmp_password 字段
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'net_cpe' AND column_name = 'cwmp_password'
        ) THEN
            ALTER TABLE net_cpe ADD COLUMN cwmp_password VARCHAR(128);
            RAISE NOTICE '✓ 已添加字段: net_cpe.cwmp_password';
        END IF;
        
        -- 添加 cwmp_status 字段
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'net_cpe' AND column_name = 'cwmp_status'
        ) THEN
            ALTER TABLE net_cpe ADD COLUMN cwmp_status VARCHAR(32) DEFAULT 'offline';
            CREATE INDEX IF NOT EXISTS idx_net_cpe_cwmp_status ON net_cpe(cwmp_status);
            RAISE NOTICE '✓ 已添加字段: net_cpe.cwmp_status';
        END IF;
        
        -- 添加 cwmp_last_inform 字段
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'net_cpe' AND column_name = 'cwmp_last_inform'
        ) THEN
            ALTER TABLE net_cpe ADD COLUMN cwmp_last_inform TIMESTAMP;
            RAISE NOTICE '✓ 已添加字段: net_cpe.cwmp_last_inform';
        END IF;
        
        RAISE NOTICE '✓ net_cpe 表字段恢复完成';
    ELSE
        RAISE NOTICE '警告: net_cpe 表不存在，跳过字段恢复';
    END IF;
END $$;

-- =====================================================
-- 8. 恢复备份数据 (如果存在)
-- =====================================================

DO $$
DECLARE
    restore_count INTEGER := 0;
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = '_backup_v8_net_cpe_cwmp') THEN
        RAISE NOTICE '检测到备份表，开始恢复数据...';
        
        -- 恢复 net_cpe 的 CWMP 字段数据
        UPDATE net_cpe AS n
        SET 
            cwmp_url = b.cwmp_url,
            cwmp_username = b.cwmp_username,
            cwmp_password = b.cwmp_password,
            cwmp_status = b.cwmp_status,
            cwmp_last_inform = b.cwmp_last_inform
        FROM _backup_v8_net_cpe_cwmp AS b
        WHERE n.id = b.id;
        
        GET DIAGNOSTICS restore_count = ROW_COUNT;
        RAISE NOTICE '✓ 已恢复 % 条 CPE CWMP 数据', restore_count;
    ELSE
        RAISE NOTICE '警告: 未找到备份表，跳过数据恢复';
        RAISE NOTICE '提示: 如需恢复数据，请从数据库备份中导入';
    END IF;
END $$;

-- =====================================================
-- 9. 更新迁移记录
-- =====================================================

INSERT INTO schema_migrations (version, description, success)
VALUES (
    'v8.0.0-restore-tr069',
    'Rollback: Restore TR069/CWMP module tables and fields',
    TRUE
)
ON CONFLICT (version) DO UPDATE
SET applied_at = CURRENT_TIMESTAMP,
    success = TRUE;

RAISE NOTICE '✓ 回滚记录已保存到 schema_migrations 表';

-- =====================================================
-- 10. 统计信息
-- =====================================================

DO $$
DECLARE
    restore_count INTEGER := 0;
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE '回滚统计信息:';
    RAISE NOTICE '========================================';
    RAISE NOTICE '已恢复表: 6 个';
    RAISE NOTICE '  - cwmp_config';
    RAISE NOTICE '  - cwmp_config_session';
    RAISE NOTICE '  - cwmp_preset';
    RAISE NOTICE '  - cwmp_preset_task';
    RAISE NOTICE '  - cwmp_factory_reset';
    RAISE NOTICE '  - cwmp_firmware_config';
    RAISE NOTICE '已修改表: 1 个 (net_cpe - 添加 5 个 CWMP 字段)';
    
    SELECT COUNT(*) INTO restore_count 
    FROM information_schema.tables 
    WHERE table_name = '_backup_v8_net_cpe_cwmp';
    
    IF restore_count > 0 THEN
        SELECT COUNT(*) INTO restore_count FROM _backup_v8_net_cpe_cwmp;
        RAISE NOTICE '恢复数据: % 条 CPE CWMP 记录', restore_count;
    ELSE
        RAISE NOTICE '恢复数据: 0 条 (未找到备份表)';
    END IF;
    
    RAISE NOTICE '========================================';
    RAISE NOTICE '回滚完成时间: %', NOW();
    RAISE NOTICE '========================================';
END $$;

-- =====================================================
-- 11. 提交事务
-- =====================================================

COMMIT;

-- 最终提示
DO $$
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '✅ ToughRadius v9 -> v8 数据库回滚成功完成！';
    RAISE NOTICE '';
    RAISE NOTICE '注意事项:';
    RAISE NOTICE '1. 表结构已恢复，但数据需要从备份中导入';
    RAISE NOTICE '2. 如果有备份表 _backup_v8_net_cpe_cwmp，数据已自动恢复';
    RAISE NOTICE '3. 其他表 (cwmp_config, cwmp_preset等) 需要手动导入数据';
    RAISE NOTICE '4. 确保 TR069 模块代码已恢复到 v8 版本';
    RAISE NOTICE '';
END $$;
