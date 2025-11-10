#!/usr/bin/env python3
"""
插入测试在线会话数据到 ToughRADIUS 数据库
"""
import sqlite3
from datetime import datetime, timedelta
import os
import sys

# 数据库文件路径
db_path = "rundata/data/toughradius.db"


def main():
    if not os.path.exists(db_path):
        print(f"❌ 数据库文件不存在: {db_path}")
        print("提示: 请先运行 ToughRADIUS 以创建数据库")
        sys.exit(1)

    # 连接数据库
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    # 检查表是否存在
    cursor.execute(
        "SELECT name FROM sqlite_master WHERE type='table' AND name='radius_online'"
    )
    if not cursor.fetchone():
        print("❌ radius_online 表不存在")
        sys.exit(1)

    # 清空现有测试数据
    cursor.execute("DELETE FROM radius_online")
    print("✓ 已清空现有在线会话数据")

    # 测试数据: (username, nas_id, nas_addr, nas_paddr, session_timeout, framed_ipaddr,
    #            framed_netmask, mac_addr, nas_port, nas_class, nas_port_id, nas_port_type,
    #            service_type, acct_session_id, acct_session_time, acct_input_total,
    #            acct_output_total, acct_input_packets, acct_output_packets,
    #            start_offset_minutes, update_offset_minutes)
    test_sessions = [
        (
            "alice@test.com",
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
            "sess-alice-001",
            1800,
            1024000000,
            2048000000,
            50000,
            100000,
            -30,
            0,
        ),
        (
            "bob@test.com",
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
            "sess-bob-001",
            3600,
            2048000000,
            4096000000,
            100000,
            200000,
            -60,
            0,
        ),
        (
            "charlie@test.com",
            "nas-002",
            "192.168.1.2",
            "10.0.0.2",
            3600,
            "172.16.2.10",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:22",
            1,
            "premium",
            "eth0/1",
            15,
            2,
            "sess-charlie-001",
            900,
            512000000,
            1024000000,
            25000,
            50000,
            -15,
            0,
        ),
        (
            "david@test.com",
            "nas-002",
            "192.168.1.2",
            "10.0.0.2",
            1800,
            "172.16.2.11",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:33",
            2,
            "standard",
            "eth0/2",
            15,
            2,
            "sess-david-001",
            600,
            256000000,
            512000000,
            12500,
            25000,
            -10,
            0,
        ),
        (
            "alice@test.com",
            "nas-003",
            "192.168.1.3",
            "10.0.0.3",
            3600,
            "172.16.3.10",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:44",
            1,
            "premium",
            "eth0/1",
            15,
            2,
            "sess-alice-002",
            2700,
            1536000000,
            3072000000,
            75000,
            150000,
            -45,
            0,
        ),
        (
            "eve@test.com",
            "nas-003",
            "192.168.1.3",
            "10.0.0.3",
            3600,
            "172.16.3.11",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:55",
            2,
            "premium",
            "eth0/2",
            15,
            2,
            "sess-eve-001",
            1200,
            768000000,
            1536000000,
            37500,
            75000,
            -20,
            0,
        ),
        (
            "frank@test.com",
            "nas-001",
            "192.168.1.1",
            "10.0.0.1",
            1800,
            "172.16.1.12",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:66",
            3,
            "basic",
            "eth0/3",
            15,
            2,
            "sess-frank-001",
            300,
            128000000,
            256000000,
            6250,
            12500,
            -5,
            0,
        ),
        (
            "grace@test.com",
            "nas-002",
            "192.168.1.2",
            "10.0.0.2",
            7200,
            "172.16.2.12",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:77",
            3,
            "premium",
            "eth0/3",
            15,
            2,
            "sess-grace-001",
            4500,
            2560000000,
            5120000000,
            125000,
            250000,
            -75,
            0,
        ),
        (
            "henry@test.com",
            "nas-004",
            "192.168.1.4",
            "10.0.0.4",
            3600,
            "172.16.4.10",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:88",
            1,
            "standard",
            "eth0/1",
            15,
            2,
            "sess-henry-001",
            1500,
            896000000,
            1792000000,
            43750,
            87500,
            -25,
            0,
        ),
        (
            "iris@test.com",
            "nas-004",
            "192.168.1.4",
            "10.0.0.4",
            1800,
            "172.16.4.11",
            "255.255.255.0",
            "AA:BB:CC:DD:EE:99",
            2,
            "basic",
            "eth0/2",
            15,
            2,
            "sess-iris-001",
            450,
            192000000,
            384000000,
            9375,
            18750,
            -8,
            0,
        ),
    ]

    now = datetime.now()
    inserted_count = 0

    for session in test_sessions:
        (
            username,
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
            acct_session_id,
            acct_session_time,
            acct_input_total,
            acct_output_total,
            acct_input_packets,
            acct_output_packets,
            start_offset,
            update_offset,
        ) = session

        acct_start_time = (now + timedelta(minutes=start_offset)).strftime(
            "%Y-%m-%d %H:%M:%S"
        )
        last_update = (now + timedelta(minutes=update_offset)).strftime(
            "%Y-%m-%d %H:%M:%S"
        )

        cursor.execute(
            """
            INSERT INTO radius_online (
                username, nas_id, nas_addr, nas_paddr, session_timeout, 
                framed_ipaddr, framed_netmask, mac_addr, nas_port, nas_class,
                nas_port_id, nas_port_type, service_type, acct_session_id,
                acct_session_time, acct_input_total, acct_output_total,
                acct_input_packets, acct_output_packets, acct_start_time, last_update
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        """,
            (
                username,
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
                acct_session_id,
                acct_session_time,
                acct_input_total,
                acct_output_total,
                acct_input_packets,
                acct_output_packets,
                acct_start_time,
                last_update,
            ),
        )
        inserted_count += 1

    conn.commit()
    print(f"✓ 成功插入 {inserted_count} 条在线会话记录")

    # 显示插入的数据
    print("\n📊 当前在线会话:")
    cursor.execute(
        """
        SELECT id, username, nas_addr, framed_ipaddr, acct_session_time, 
               acct_input_total/1024/1024 as input_mb, 
               acct_output_total/1024/1024 as output_mb
        FROM radius_online 
        ORDER BY acct_start_time DESC
    """
    )

    print(
        f"{'ID':<4} {'用户名':<20} {'NAS地址':<15} {'分配IP':<15} {'在线时长(s)':<12} {'上传(MB)':<10} {'下载(MB)':<10}"
    )
    print("-" * 100)
    for row in cursor.fetchall():
        print(
            f"{row[0]:<4} {row[1]:<20} {row[2]:<15} {row[3]:<15} {row[4]:<12} {row[5]:<10.2f} {row[6]:<10.2f}"
        )

    conn.close()
    print(f"\n✓ 测试数据已准备完成!")
    print(f"\n💡 测试 API 命令:")
    print(f"   1. 获取所有在线会话: curl http://localhost:1816/api/v1/sessions")
    print(
        f"   2. 分页查询: curl 'http://localhost:1816/api/v1/sessions?page=1&perPage=5'"
    )
    print(
        f"   3. 按用户搜索: curl 'http://localhost:1816/api/v1/sessions?username=alice'"
    )
    print(
        f"   4. 按NAS过滤: curl 'http://localhost:1816/api/v1/sessions?nas_addr=192.168.1.1'"
    )


if __name__ == "__main__":
    main()
