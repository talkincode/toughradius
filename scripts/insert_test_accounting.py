#!/usr/bin/env python3
"""
插入测试计费记录数据到 ToughRADIUS 数据库
"""
import sqlite3
from datetime import datetime, timedelta
import os
import sys

# Database file path
db_path = "rundata/data/toughradius.db"


def main():
    if not os.path.exists(db_path):
        print(f"❌ 数据库文件不存在: {db_path}")
        print("提示: 请先运行 ToughRADIUS 以创建数据库")
        sys.exit(1)

# Connect to the database
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

# Check if the tables exist
    cursor.execute(
        "SELECT name FROM sqlite_master WHERE type='table' AND name='radius_accounting'"
    )
    if not cursor.fetchone():
        print("❌ radius_accounting 表不存在")
        sys.exit(1)

# Clear existing test data
    cursor.execute("DELETE FROM radius_accounting")
    print("✓ 已清空现有计费记录数据")

# Test data
    now = datetime.now()
    test_records = [
        # Completed session
        (
            "alice@test.com",
            "sess-alice-complete-001",
            "nas-001",
            "192.168.1.1",
            "10.0.0.1",
            3600,
            "172.16.1.10",
            "255.255.255.0",
            "00:11:22:33:44:55",
            1,
            "premium",
            "eth0/1",
            15,
            2,
            3600,
            1024000000,
            2048000000,
            50000,
            100000,
            -120,  # Start 2 hours ago
            -60,  # End 1 hour ago
        ),
        (
            "bob@test.com",
            "sess-bob-complete-001",
            "nas-001",
            "192.168.1.1",
            "10.0.0.1",
            7200,
            "172.16.1.11",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:11",
            2,
            "standard",
            "eth0/2",
            15,
            2,
            7200,
            2048000000,
            4096000000,
            100000,
            200000,
            -180,  # Start 3 hours ago
            -60,  # End 1 hour ago
        ),
        (
            "charlie@test.com",
            "sess-charlie-complete-001",
            "nas-002",
            "192.168.1.2",
            "10.0.0.2",
            1800,
            "172.16.2.10",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:22",
            1,
            "premium",
            "eth0/1",
            15,
            2,
            1800,
            512000000,
            1024000000,
            25000,
            50000,
            -90,  # Start 90 minutes ago
            -30,  # End 30 minutes ago
        ),
        (
            "david@test.com",
            "sess-david-complete-001",
            "nas-002",
            "192.168.1.2",
            "10.0.0.2",
            600,
            "172.16.2.11",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:33",
            2,
            "basic",
            "eth0/2",
            15,
            2,
            600,
            256000000,
            512000000,
            12500,
            25000,
            -45,  # Start 45 minutes ago
            -35,  # End 35 minutes ago
        ),
        (
            "eve@test.com",
            "sess-eve-complete-001",
            "nas-003",
            "192.168.1.3",
            "10.0.0.3",
            5400,
            "172.16.3.10",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:44",
            1,
            "premium",
            "eth0/1",
            15,
            2,
            5400,
            3072000000,
            6144000000,
            150000,
            300000,
            -240,  # Start 4 hours ago
            -150,  # End 2.5 hours ago
        ),
        (
            "frank@test.com",
            "sess-frank-complete-001",
            "nas-003",
            "192.168.1.3",
            "10.0.0.3",
            900,
            "172.16.3.11",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:55",
            2,
            "standard",
            "eth0/2",
            15,
            2,
            900,
            384000000,
            768000000,
            18750,
            37500,
            -60,  # Start 1 hour ago
            -45,  # End 45 minutes ago
        ),
        (
            "grace@test.com",
            "sess-grace-complete-001",
            "nas-001",
            "192.168.1.1",
            "10.0.0.1",
            10800,
            "172.16.1.12",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:66",
            3,
            "premium",
            "eth0/3",
            15,
            2,
            10800,
            5120000000,
            10240000000,
            250000,
            500000,
            -360,  # Start 6 hours ago
            -180,  # End 3 hours ago
        ),
        (
            "henry@test.com",
            "sess-henry-complete-001",
            "nas-004",
            "192.168.1.4",
            "10.0.0.4",
            2700,
            "172.16.4.10",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:77",
            1,
            "standard",
            "eth0/1",
            15,
            2,
            2700,
            1536000000,
            3072000000,
            75000,
            150000,
            -150,  # Start 2.5 hours ago
            -105,  # End 1 hour 45 minutes ago
        ),
        (
            "iris@test.com",
            "sess-iris-complete-001",
            "nas-004",
            "192.168.1.4",
            "10.0.0.4",
            1200,
            "172.16.4.11",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:88",
            2,
            "basic",
            "eth0/2",
            15,
            2,
            1200,
            640000000,
            1280000000,
            31250,
            62500,
            -100,  # Start 100 minutes ago
            -80,  # End 80 minutes ago
        ),
        (
            "jack@test.com",
            "sess-jack-complete-001",
            "nas-002",
            "192.168.1.2",
            "10.0.0.2",
            4500,
            "172.16.2.12",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:99",
            3,
            "premium",
            "eth0/3",
            15,
            2,
            4500,
            2560000000,
            5120000000,
            125000,
            250000,
            -300,  # Start 5 hours ago
            -225,  # End 3 hours 45 minutes ago
        ),
        # Most recent records
        (
            "alice@test.com",
            "sess-alice-recent-001",
            "nas-001",
            "192.168.1.1",
            "10.0.0.1",
            1800,
            "172.16.1.10",
            "255.255.255.0",
            "00:11:22:33:44:55",
            1,
            "premium",
            "eth0/1",
            15,
            2,
            1800,
            896000000,
            1792000000,
            43750,
            87500,
            -30,  # Start 30 minutes ago
            -0,  # Just ended
        ),
        (
            "bob@test.com",
            "sess-bob-recent-001",
            "nas-003",
            "192.168.1.3",
            "10.0.0.3",
            3000,
            "172.16.3.12",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:AA",
            1,
            "standard",
            "eth0/1",
            15,
            2,
            3000,
            1792000000,
            3584000000,
            87500,
            175000,
            -50,  # Start 50 minutes ago
            -0,  # Just ended
        ),
    ]

    inserted_count = 0

    for record in test_records:
        (
            username,
            acct_session_id,
            nas_id,
            nas_addr,
            nas_paddr,
            session_timeout,
            framed_ipaddr,
            framed_netmask,
            mac_addr,
            nas_port,
            nas_class,
            nas_port_id,
            nas_port_type,
            service_type,
            acct_session_time,
            acct_input_total,
            acct_output_total,
            acct_input_packets,
            acct_output_packets,
            start_offset,
            stop_offset,
        ) = record

        acct_start_time = (now + timedelta(minutes=start_offset)).strftime(
            "%Y-%m-%d %H:%M:%S"
        )
        acct_stop_time = (now + timedelta(minutes=stop_offset)).strftime(
            "%Y-%m-%d %H:%M:%S"
        )
        last_update = acct_stop_time

        cursor.execute(
            """
            INSERT INTO radius_accounting (
                username, acct_session_id, nas_id, nas_addr, nas_paddr, 
                session_timeout, framed_ipaddr, framed_netmask, mac_addr, 
                nas_port, nas_class, nas_port_id, nas_port_type, service_type,
                acct_session_time, acct_input_total, acct_output_total,
                acct_input_packets, acct_output_packets, 
                acct_start_time, acct_stop_time, last_update
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        """,
            (
                username,
                acct_session_id,
                nas_id,
                nas_addr,
                nas_paddr,
                session_timeout,
                framed_ipaddr,
                framed_netmask,
                mac_addr,
                nas_port,
                nas_class,
                nas_port_id,
                nas_port_type,
                service_type,
                acct_session_time,
                acct_input_total,
                acct_output_total,
                acct_input_packets,
                acct_output_packets,
                acct_start_time,
                acct_stop_time,
                last_update,
            ),
        )
        inserted_count += 1

    conn.commit()
    print(f"✓ 成功插入 {inserted_count} 条计费记录")

    # Display statistics of the inserted records
    print("\n📊 计费记录统计:")
    cursor.execute(
        """
        SELECT 
            COUNT(*) as total,
            SUM(acct_session_time) as total_time,
            SUM(acct_input_total)/1024/1024/1024 as total_upload_gb,
            SUM(acct_output_total)/1024/1024/1024 as total_download_gb
        FROM radius_accounting
    """
    )
    row = cursor.fetchone()
    print(f"总记录数: {row[0]}")
    print(f"总会话时长: {row[1]} 秒 ({row[1]//3600} 小时)")
    print(f"总上传流量: {row[2]:.2f} GB")
    print(f"总下载流量: {row[3]:.2f} GB")

    # Display recent records
    print("\n📋 最近的5条计费记录:")
    cursor.execute(
        """
        SELECT 
            id, username, nas_addr, framed_ipaddr, 
            acct_session_time,
            acct_input_total/1024/1024 as input_mb, 
            acct_output_total/1024/1024 as output_mb,
            acct_start_time, acct_stop_time
        FROM radius_accounting 
        ORDER BY acct_stop_time DESC
        LIMIT 5
    """
    )

    print(
        f"{'ID':<4} {'用户名':<20} {'NAS':<15} {'IP':<15} {'时长(s)':<10} {'上传MB':<10} {'下载MB':<10}"
    )
    print("-" * 100)
    for row in cursor.fetchall():
        print(
            f"{row[0]:<4} {row[1]:<20} {row[2]:<15} {row[3]:<15} {row[4]:<10} {row[5]:<10.2f} {row[6]:<10.2f}"
        )

    # Show statistics per user
    print("\n👥 用户流量统计 (TOP 5):")
    cursor.execute(
        """
        SELECT 
            username,
            COUNT(*) as session_count,
            SUM(acct_session_time) as total_time,
            SUM(acct_input_total)/1024/1024/1024 as upload_gb,
            SUM(acct_output_total)/1024/1024/1024 as download_gb
        FROM radius_accounting
        GROUP BY username
        ORDER BY (acct_input_total + acct_output_total) DESC
        LIMIT 5
    """
    )

    print(
        f"{'用户名':<20} {'会话数':<10} {'总时长(h)':<12} {'上传GB':<12} {'下载GB':<12}"
    )
    print("-" * 70)
    for row in cursor.fetchall():
        print(
            f"{row[0]:<20} {row[1]:<10} {row[2]//3600:<12} {row[3]:<12.2f} {row[4]:<12.2f}"
        )

    conn.close()
    print(f"\n✓ 测试数据已准备完成!")
    print(f"\n💡 测试 API 命令:")
    print(f"   1. 获取所有计费记录: curl http://localhost:1816/api/v1/accounting")
    print(
        f"   2. 分页查询: curl 'http://localhost:1816/api/v1/accounting?page=1&perPage=5'"
    )
    print(
        f"   3. 按用户搜索: curl 'http://localhost:1816/api/v1/accounting?username=alice'"
    )
    print(
        f"   4. 按会话ID: curl 'http://localhost:1816/api/v1/accounting?acct_session_id=sess-alice-complete-001'"
    )
    print(
        f"   5. 时间范围: curl 'http://localhost:1816/api/v1/accounting?start_time=2025-01-01T00:00:00Z'"
    )


if __name__ == "__main__":
    main()
