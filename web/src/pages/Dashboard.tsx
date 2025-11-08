import { useEffect, useState } from 'react';
import ReactECharts from 'echarts-for-react';

interface DashboardStats {
  totalUsers: number;
  onlineUsers: number;
  todayAuth: number;
  todayAcct: number;
}

const Dashboard = () => {
  const [stats, setStats] = useState<DashboardStats>({
    totalUsers: 0,
    onlineUsers: 0,
    todayAuth: 0,
    todayAcct: 0,
  });

  useEffect(() => {
    // 获取统计数据
    // TODO: 从 API 获取实际数据
    setStats({
      totalUsers: 1250,
      onlineUsers: 86,
      todayAuth: 542,
      todayAcct: 2380,
    });
  }, []);

  // 认证趋势图配置
  const authTrendOption = {
    title: {
      text: '认证趋势（近7天）',
      left: 'center',
    },
    tooltip: {
      trigger: 'axis',
    },
    xAxis: {
      type: 'category',
      data: ['周一', '周二', '周三', '周四', '周五', '周六', '周日'],
    },
    yAxis: {
      type: 'value',
    },
    series: [
      {
        name: '认证次数',
        type: 'line',
        smooth: true,
        data: [320, 432, 401, 534, 590, 530, 520],
      },
    ],
  };

  // 在线用户分布图配置
  const onlineDistributionOption = {
    title: {
      text: '在线用户分布',
      left: 'center',
    },
    tooltip: {
      trigger: 'item',
    },
    legend: {
      orient: 'vertical',
      left: 'left',
    },
    series: [
      {
        name: '在线用户',
        type: 'pie',
        radius: '50%',
        data: [
          { value: 35, name: 'PPPoE' },
          { value: 28, name: 'IPoE' },
          { value: 15, name: 'WiFi' },
          { value: 8, name: '其他' },
        ],
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)',
          },
        },
      },
    ],
  };

  // 流量统计图配置
  const trafficOption = {
    title: {
      text: '流量统计（近24小时）',
      left: 'center',
    },
    tooltip: {
      trigger: 'axis',
    },
    legend: {
      data: ['上传', '下载'],
      top: 30,
    },
    xAxis: {
      type: 'category',
      data: Array.from({ length: 24 }, (_, i) => `${i}:00`),
    },
    yAxis: {
      type: 'value',
      name: 'GB',
    },
    series: [
      {
        name: '上传',
        type: 'bar',
        data: Array.from({ length: 24 }, () => Math.random() * 100),
      },
      {
        name: '下载',
        type: 'bar',
        data: Array.from({ length: 24 }, () => Math.random() * 200),
      },
    ],
  };

  return (
    <div style={{ padding: '20px' }}>
      <h1>欢迎使用 ToughRADIUS v9</h1>

      {/* 统计卡片 */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '20px', marginBottom: '20px' }}>
        <div style={{ padding: '20px', backgroundColor: '#fff', borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '14px', color: '#999' }}>总用户数</div>
            <div style={{ fontSize: '32px', fontWeight: 'bold', marginTop: '10px' }}>{stats.totalUsers}</div>
          </div>
        </div>

        <div style={{ padding: '20px', backgroundColor: '#fff', borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '14px', color: '#999' }}>在线用户</div>
            <div style={{ fontSize: '32px', fontWeight: 'bold', marginTop: '10px', color: '#52c41a' }}>
              {stats.onlineUsers}
            </div>
          </div>
        </div>

        <div style={{ padding: '20px', backgroundColor: '#fff', borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '14px', color: '#999' }}>今日认证</div>
            <div style={{ fontSize: '32px', fontWeight: 'bold', marginTop: '10px', color: '#1890ff' }}>
              {stats.todayAuth}
            </div>
          </div>
        </div>

        <div style={{ padding: '20px', backgroundColor: '#fff', borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: '14px', color: '#999' }}>今日计费</div>
            <div style={{ fontSize: '32px', fontWeight: 'bold', marginTop: '10px', color: '#722ed1' }}>
              {stats.todayAcct}
            </div>
          </div>
        </div>
      </div>

      {/* 图表 */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '20px' }}>
        <div style={{ padding: '20px', backgroundColor: '#fff', borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
          <ReactECharts option={authTrendOption} style={{ height: '400px' }} />
        </div>

        <div style={{ padding: '20px', backgroundColor: '#fff', borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
          <ReactECharts option={onlineDistributionOption} style={{ height: '400px' }} />
        </div>

        <div style={{ gridColumn: '1 / -1', padding: '20px', backgroundColor: '#fff', borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
          <ReactECharts option={trafficOption} style={{ height: '400px' }} />
        </div>
      </div>

      {/* 系统信息 */}
      <div style={{ marginTop: '20px', padding: '20px', backgroundColor: '#fff', borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
        <h2>系统信息</h2>
        <div style={{ padding: '10px' }}>
          <p><strong>版本:</strong> ToughRADIUS v9.0.0</p>
          <p><strong>架构:</strong> React Admin + Go Backend</p>
          <p><strong>主要功能:</strong></p>
          <ul>
            <li>RADIUS 用户管理 - 支持用户增删改查、批量操作</li>
            <li>在线会话监控 - 实时查看在线用户和会话状态</li>
            <li>计费记录查询 - 支持多条件筛选和统计分析</li>
            <li>RADIUS 配置管理 - 灵活的配置文件和策略管理</li>
            <li>数据可视化 - 丰富的图表和统计报表</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
