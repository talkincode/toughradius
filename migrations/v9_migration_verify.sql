-- =====================================================
-- ToughRadius v9 数据库迁移验证脚本
-- 功能: 验证 TR069/CWMP 模块移除是否完整
-- 版本: v9
-- 作者: ToughRADIUS Team
-- 日期: 2025-11-08
-- =====================================================

-- 设置输出格式
\set QUIET on
\pset border 2
\pset format wrapped

\echo ''
\echo '=========================================='
\echo 'ToughRadius v9 迁移验证报告'
\echo '=========================================='
\echo ''

-- =====================================================
-- 1. 检查 CWMP 表是否已删除
-- =====================================================

\echo '1. 检查 CWMP 相关表...'
\echo '------------------------------------------'

SELECT 
    CASE 
        WHEN COUNT(*) = 0 THEN '✅ 通过'
        ELSE '❌ 失败: 发现 ' || COUNT(*) || ' 个 CWMP 表'
    END AS "检查结果",
    COALESCE(STRING_AGG(table_name, ', '), '无') AS "残留表"
FROM information_schema.tables
WHERE table_schema = 'public'
  AND (
    table_name LIKE 'cwmp%' 
    OR table_name IN ('cwmp_config', 'cwmp_config_session', 'cwmp_preset', 
                      'cwmp_preset_task', 'cwmp_factory_reset', 'cwmp_firmware_config')
  );

\echo ''

-- =====================================================
-- 2. 检查 net_cpe 表的 CWMP 字段
-- =====================================================

\echo '2. 检查 net_cpe 表的 CWMP 字段...'
\echo '------------------------------------------'

SELECT 
    CASE 
        WHEN COUNT(*) = 0 THEN '✅ 通过'
        ELSE '❌ 失败: 发现 ' || COUNT(*) || ' 个 CWMP 字段'
    END AS "检查结果",
    COALESCE(STRING_AGG(column_name, ', '), '无') AS "残留字段"
FROM information_schema.columns
WHERE table_name = 'net_cpe'
  AND column_name LIKE 'cwmp%';

\echo ''

-- =====================================================
-- 3. 检查 CWMP 相关索引
-- =====================================================

\echo '3. 检查 CWMP 相关索引...'
\echo '------------------------------------------'

SELECT 
    CASE 
        WHEN COUNT(*) = 0 THEN '✅ 通过'
        ELSE '⚠️  警告: 发现 ' || COUNT(*) || ' 个 CWMP 索引'
    END AS "检查结果",
    COALESCE(STRING_AGG(indexname, ', '), '无') AS "残留索引"
FROM pg_indexes
WHERE schemaname = 'public'
  AND indexname LIKE '%cwmp%';

\echo ''

-- =====================================================
-- 4. 检查 CWMP 相关序列
-- =====================================================

\echo '4. 检查 CWMP 相关序列...'
\echo '------------------------------------------'

SELECT 
    CASE 
        WHEN COUNT(*) = 0 THEN '✅ 通过'
        ELSE '⚠️  警告: 发现 ' || COUNT(*) || ' 个 CWMP 序列'
    END AS "检查结果",
    COALESCE(STRING_AGG(sequencename, ', '), '无') AS "残留序列"
FROM pg_sequences
WHERE schemaname = 'public'
  AND sequencename LIKE '%cwmp%';

\echo ''

-- =====================================================
-- 5. 检查备份表
-- =====================================================

\echo '5. 检查备份表状态...'
\echo '------------------------------------------'

SELECT 
    CASE 
        WHEN COUNT(*) > 0 THEN '✅ 备份表存在'
        ELSE '⚠️  警告: 备份表不存在'
    END AS "备份状态",
    COALESCE(
        (SELECT COUNT(*)::TEXT || ' 条记录' 
         FROM _backup_v8_net_cpe_cwmp 
         WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = '_backup_v8_net_cpe_cwmp')),
        '0 条记录'
    ) AS "备份数据量"
FROM information_schema.tables
WHERE table_name = '_backup_v8_net_cpe_cwmp';

\echo ''

-- =====================================================
-- 6. 检查迁移记录
-- =====================================================

\echo '6. 检查迁移记录...'
\echo '------------------------------------------'

SELECT 
    version AS "迁移版本",
    description AS "描述",
    applied_at AS "执行时间",
    CASE 
        WHEN success THEN '✅ 成功'
        ELSE '❌ 失败'
    END AS "状态"
FROM schema_migrations
WHERE version LIKE '%tr069%' OR version LIKE '%cwmp%'
ORDER BY applied_at DESC
LIMIT 5;

\echo ''

-- =====================================================
-- 7. 数据库完整性检查
-- =====================================================

\echo '7. 数据库完整性检查...'
\echo '------------------------------------------'

-- 检查表数量
SELECT 
    COUNT(*) AS "当前表数量"
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_type = 'BASE TABLE';

-- 检查 net_cpe 表结构
SELECT 
    COUNT(*) AS "net_cpe 字段数量"
FROM information_schema.columns
WHERE table_name = 'net_cpe';

\echo ''

-- =====================================================
-- 8. 详细验证 (可选)
-- =====================================================

\echo '8. 详细验证 - 所有公共表列表...'
\echo '------------------------------------------'

SELECT 
    table_name AS "表名",
    (SELECT COUNT(*) FROM information_schema.columns WHERE columns.table_name = tables.table_name) AS "字段数"
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_type = 'BASE TABLE'
  AND table_name NOT LIKE 'pg_%'
  AND table_name NOT LIKE '_backup%'
ORDER BY table_name;

\echo ''

-- =====================================================
-- 9. net_cpe 表结构详情
-- =====================================================

\echo '9. net_cpe 表当前结构...'
\echo '------------------------------------------'

SELECT 
    column_name AS "字段名",
    data_type AS "数据类型",
    CASE 
        WHEN is_nullable = 'YES' THEN 'NULL'
        ELSE 'NOT NULL'
    END AS "是否允许空",
    column_default AS "默认值"
FROM information_schema.columns
WHERE table_name = 'net_cpe'
ORDER BY ordinal_position;

\echo ''

-- =====================================================
-- 10. 总结报告
-- =====================================================

\echo '=========================================='
\echo '验证总结'
\echo '=========================================='

DO $$
DECLARE
    cwmp_tables INTEGER;
    cwmp_fields INTEGER;
    cwmp_indexes INTEGER;
    backup_exists BOOLEAN;
    migration_exists BOOLEAN;
BEGIN
    -- 检查表
    SELECT COUNT(*) INTO cwmp_tables
    FROM information_schema.tables
    WHERE table_schema = 'public'
      AND table_name LIKE 'cwmp%';
    
    -- 检查字段
    SELECT COUNT(*) INTO cwmp_fields
    FROM information_schema.columns
    WHERE table_name = 'net_cpe'
      AND column_name LIKE 'cwmp%';
    
    -- 检查索引
    SELECT COUNT(*) INTO cwmp_indexes
    FROM pg_indexes
    WHERE schemaname = 'public'
      AND indexname LIKE '%cwmp%';
    
    -- 检查备份
    SELECT EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = '_backup_v8_net_cpe_cwmp'
    ) INTO backup_exists;
    
    -- 检查迁移记录
    SELECT EXISTS (
        SELECT 1 FROM schema_migrations 
        WHERE version = 'v9.0.0-remove-tr069' AND success = TRUE
    ) INTO migration_exists;
    
    -- 输出总结
    RAISE NOTICE '';
    RAISE NOTICE '检查项目总结:';
    RAISE NOTICE '----------------------------------------';
    RAISE NOTICE 'CWMP 表: % (应为 0)', cwmp_tables;
    RAISE NOTICE 'CWMP 字段: % (应为 0)', cwmp_fields;
    RAISE NOTICE 'CWMP 索引: % (应为 0)', cwmp_indexes;
    RAISE NOTICE '备份表: % (建议保留)', CASE WHEN backup_exists THEN '存在' ELSE '不存在' END;
    RAISE NOTICE '迁移记录: % (应为存在)', CASE WHEN migration_exists THEN '存在' ELSE '不存在' END;
    RAISE NOTICE '----------------------------------------';
    RAISE NOTICE '';
    
    -- 最终判定
    IF cwmp_tables = 0 AND cwmp_fields = 0 AND migration_exists THEN
        RAISE NOTICE '✅ 迁移验证通过！';
        RAISE NOTICE '数据库已成功移除 TR069/CWMP 模块';
    ELSE
        RAISE NOTICE '❌ 迁移验证失败！';
        RAISE NOTICE '请检查上述报告，手动清理残留项';
    END IF;
    
    RAISE NOTICE '';
END $$;

\echo ''
\echo '=========================================='
\echo '验证完成'
\echo '=========================================='
\echo ''
\echo '后续建议:'
\echo '1. 如果验证通过，可以继续代码清理'
\echo '2. 备份表 _backup_v8_net_cpe_cwmp 建议保留一段时间'
\echo '3. 定期检查应用日志，确认无 CWMP 相关错误'
\echo '4. 如需回滚，运行 v9_migration_down.sql'
\echo ''

-- 设置退出状态
\set QUIET off
