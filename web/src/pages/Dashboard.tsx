import PeopleAltOutlinedIcon from '@mui/icons-material/PeopleAltOutlined';
import OnlinePredictionOutlinedIcon from '@mui/icons-material/OnlinePredictionOutlined';
import VerifiedUserOutlinedIcon from '@mui/icons-material/VerifiedUserOutlined';
import SwapVertOutlinedIcon from '@mui/icons-material/SwapVertOutlined';
import Grid from '@mui/material/GridLegacy';
import {
  Box,
  Card,
  CardContent,
  Chip,
  LinearProgress,
  Stack,
  Typography,
} from '@mui/material';
import { alpha, useTheme } from '@mui/material/styles';
import ReactECharts from 'echarts-for-react';
import { useEffect, useMemo, useState } from 'react';

interface DashboardStats {
  total_users: number;
  online_users: number;
  today_auth_count: number;
  today_acct_count: number;
  total_profiles: number;
  disabled_users: number;
  expired_users: number;
  today_input_gb: number;
  today_output_gb: number;
}

const weekDays = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];
const weeklyAuthData = [320, 432, 401, 534, 590, 530, 520];
const connectionSources = [
  { value: 45, name: 'PPPoE' },
  { value: 30, name: 'IPoE' },
  { value: 15, name: 'WiFi' },
  { value: 10, name: '其他' },
];
const trafficHours = Array.from({ length: 24 }, (_, hour) => `${hour}:00`);
const trafficUpload = [18, 16, 20, 24, 26, 33, 37, 40, 42, 45, 48, 52, 54, 60, 64, 71, 74, 70, 60, 52, 45, 36, 28, 22];
const trafficDownload = [42, 48, 50, 56, 64, 72, 80, 90, 102, 118, 128, 138, 146, 154, 162, 170, 168, 158, 140, 128, 118, 102, 80, 60];
const numberFormatter = new Intl.NumberFormat('zh-CN');

const defaultStats: DashboardStats = {
  total_users: 1250,
  online_users: 423,
  today_auth_count: 3784,
  today_acct_count: 2156,
  total_profiles: 8,
  disabled_users: 45,
  expired_users: 23,
  today_input_gb: 156.78,
  today_output_gb: 892.34,
};

const Dashboard = () => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const [stats, setStats] = useState<DashboardStats>(defaultStats);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const token = localStorage.getItem('token');
        const headers: Record<string, string> = {
          'Content-Type': 'application/json',
        };

        if (token) {
          headers.Authorization = `Bearer ${token}`;
        }

        const response = await fetch('/api/v1/dashboard/stats', {
          headers,
        });

        if (!response.ok) {
          // Silently handle API errors, don't log to console
          if (response.status === 401) {
            // Handle unauthorized - maybe redirect to login
            console.warn('Dashboard: Unauthorized access');
          } else if (response.status === 404) {
            console.warn('Dashboard: API endpoint not found');
          } else {
            console.warn(`Dashboard: API request failed with status ${response.status}`);
          }
          return;
        }

        const data = await response.json();
        setStats((prev) => ({ ...prev, ...data }));
      } catch (error) {
        // Silent error handling - avoid console errors
        if (error instanceof TypeError && error.message.includes('fetch')) {
          console.warn('Dashboard: Network error - using demo data');
        } else {
          console.warn('Dashboard: Failed to load stats - using demo data');
        }
      }
    };

    fetchStats();
  }, []);

  const onlineRatio =
    stats.total_users > 0 ? Math.min((stats.online_users / stats.total_users) * 100, 100) : 0;

  const statCards = [
    {
      title: '总用户数',
      value: numberFormatter.format(stats.total_users),
      icon: <PeopleAltOutlinedIcon fontSize="large" />,
      accent: theme.palette.primary.main,
      highlights: [
        { label: '禁用', value: stats.disabled_users },
        { label: '过期', value: stats.expired_users },
      ],
    },
    {
      title: '在线用户',
      value: numberFormatter.format(stats.online_users),
      icon: <OnlinePredictionOutlinedIcon fontSize="large" />,
      accent: '#34d399',
      highlights: [{ label: '策略总数', value: stats.total_profiles }],
    },
    {
      title: '今日认证',
      value: numberFormatter.format(stats.today_auth_count),
      icon: <VerifiedUserOutlinedIcon fontSize="large" />,
      accent: theme.palette.secondary.main,
      highlights: [{ label: '计费记录', value: stats.today_acct_count }],
    },
    {
      title: '今日流量',
      value: `↑ ${stats.today_input_gb.toFixed(2)} GB`,
      secondaryValue: `↓ ${stats.today_output_gb.toFixed(2)} GB`,
      icon: <SwapVertOutlinedIcon fontSize="large" />,
      accent: '#f97316',
      highlights: [{ label: '单位', value: 'GB' }],
    },
  ];

  const authTrendOption = useMemo(
    () => ({
      backgroundColor: 'transparent',
      tooltip: { trigger: 'axis' },
      textStyle: { color: alpha(theme.palette.text.primary, 0.7) },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: {
        type: 'category',
        data: weekDays,
        boundaryGap: false,
        axisLine: {
          lineStyle: { color: alpha(theme.palette.text.secondary, 0.25) },
        },
        axisLabel: {
          color: alpha(theme.palette.text.primary, 0.6),
        },
      },
      yAxis: {
        type: 'value',
        splitLine: {
          lineStyle: { color: alpha(theme.palette.text.secondary, 0.15) },
        },
        axisLabel: {
          color: alpha(theme.palette.text.secondary, 0.65),
        },
      },
      series: [
        {
          name: '认证次数',
          type: 'line',
          smooth: true,
          symbol: 'circle',
          symbolSize: 10,
          data: weeklyAuthData,
          lineStyle: { width: 4 },
          itemStyle: { color: theme.palette.primary.main },
          areaStyle: {
            color: alpha(theme.palette.primary.main, 0.18),
          },
        },
      ],
    }),
    [theme],
  );

  const onlineDistributionOption = useMemo(
    () => ({
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'item',
      },
      legend: {
        orient: 'vertical',
        left: 0,
        textStyle: { color: alpha(theme.palette.text.primary, 0.75) },
      },
      series: [
        {
          name: '在线用户',
          type: 'pie',
          radius: ['35%', '70%'],
          avoidLabelOverlap: false,
          itemStyle: {
            borderRadius: 8,
            borderColor: '#fff',
            borderWidth: 2,
          },
          label: {
            formatter: '{b}\n{d}%',
            color: theme.palette.text.primary,
          },
          labelLine: {
            smooth: true,
            length: 20,
          },
          data: connectionSources,
          color: [
            theme.palette.primary.main,
            theme.palette.secondary.main,
            '#34d399',
            '#facc15',
          ],
        },
      ],
    }),
    [theme],
  );

  const trafficOption = useMemo(
    () => ({
      backgroundColor: 'transparent',
      tooltip: { trigger: 'axis' },
      legend: {
        data: ['上传', '下载'],
        top: 0,
        textStyle: { color: alpha(theme.palette.text.primary, 0.7) },
      },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: {
        type: 'category',
        data: trafficHours,
        axisLine: {
          lineStyle: { color: alpha(theme.palette.text.secondary, 0.2) },
        },
        axisLabel: {
          color: alpha(theme.palette.text.primary, 0.6),
        },
      },
      yAxis: {
        type: 'value',
        name: 'GB',
        nameTextStyle: {
          color: alpha(theme.palette.text.secondary, 0.7),
        },
        splitLine: {
          lineStyle: { color: alpha(theme.palette.text.secondary, 0.1) },
        },
      },
      series: [
        {
          name: '上传',
          type: 'bar',
          stack: 'traffic',
          emphasis: { focus: 'series' },
          data: trafficUpload,
          color: alpha(theme.palette.secondary.main, 0.7),
        },
        {
          name: '下载',
          type: 'bar',
          stack: 'traffic',
          emphasis: { focus: 'series' },
          data: trafficDownload,
          color: theme.palette.primary.main,
        },
      ],
    }),
    [theme],
  );

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
      <Card
        sx={{
          borderRadius: 4,
          overflow: 'hidden',
          background: isDark 
            ? 'linear-gradient(135deg, #1e293b, #334155)' 
            : 'linear-gradient(135deg, #eef2ff, #fdf2f8)',
          border: `1px solid ${isDark ? 'rgba(148, 163, 184, 0.1)' : 'rgba(255, 255, 255, 0.6)'}`,
        }}
      >
        <CardContent>
          <Stack
            direction={{ xs: 'column', md: 'row' }}
            spacing={3}
            alignItems="center"
            justifyContent="space-between"
          >
            <Box>
              <Chip label="RADIUS 系统总览" color="primary" sx={{ mb: 2, fontWeight: 600 }} />
              <Typography variant="body1" sx={{ color: 'text.secondary', maxWidth: 520 }}>
                通过实时洞察快速掌握系统动态,精准控制 RADIUS 用户、会话及策略配置，保障网络稳定运行。
              </Typography>

              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={4} sx={{ mt: 3 }}>
                <Box>
                  <Typography variant="subtitle2" color="text.secondary">
                    今日认证
                  </Typography>
                  <Typography variant="h5" sx={{ fontWeight: 700, color: 'text.primary' }}>
                    {numberFormatter.format(stats.today_auth_count)}
                  </Typography>
                </Box>
                <Box>
                  <Typography variant="subtitle2" color="text.secondary">
                    今日计费记录
                  </Typography>
                  <Typography variant="h5" sx={{ fontWeight: 700, color: 'text.primary' }}>
                    {numberFormatter.format(stats.today_acct_count)}
                  </Typography>
                </Box>
              </Stack>
            </Box>

            <Box sx={{ minWidth: 260 }}>
              <Typography variant="subtitle2" color="text.secondary">
                在线用户占比
              </Typography>
              <Typography variant="h3" sx={{ fontWeight: 700, my: 1 }}>
                {onlineRatio.toFixed(1)}%
              </Typography>
              <LinearProgress
                variant="determinate"
                value={onlineRatio}
                sx={{
                  height: 10,
                  borderRadius: 999,
                  backgroundColor: alpha(theme.palette.primary.main, 0.15),
                  '& .MuiLinearProgress-bar': {
                    borderRadius: 999,
                  },
                }}
              />
              <Stack direction="row" justifyContent="space-between" sx={{ mt: 2 }}>
                <Typography variant="body2" color="text.secondary">
                  在线 {stats.online_users}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  总数 {stats.total_users}
                </Typography>
              </Stack>
            </Box>
          </Stack>
        </CardContent>
      </Card>

      <Grid container spacing={3}>
        {statCards.map((card) => (
          <Grid item xs={12} sm={6} lg={3} key={card.title}>
            <Card
              sx={{
                height: '100%',
                borderRadius: 4,
              }}
            >
              <CardContent>
                <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary">
                      {card.title}
                    </Typography>
                    <Typography variant="h4" sx={{ fontWeight: 700, my: 1 }}>
                      {card.value}
                    </Typography>
                    {card.secondaryValue && (
                      <Typography variant="h6" sx={{ color: alpha(theme.palette.text.primary, 0.65) }}>
                        {card.secondaryValue}
                      </Typography>
                    )}
                  </Box>
                  <Box
                    sx={{
                      width: 56,
                      height: 56,
                      borderRadius: 3,
                      display: 'grid',
                      placeItems: 'center',
                      backgroundColor: alpha(card.accent, 0.15),
                      color: card.accent,
                    }}
                  >
                    {card.icon}
                  </Box>
                </Stack>
                <Stack direction="row" spacing={1} sx={{ mt: 2, flexWrap: 'wrap' }}>
                  {card.highlights.map((item) => (
                    <Chip key={item.label} label={`${item.label}: ${item.value}`} size="small" />
                  ))}
                </Stack>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Card sx={{ borderRadius: 4, height: '100%' }}>
            <CardContent sx={{ height: '100%' }}>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 2 }}>
                认证趋势（近 7 天）
              </Typography>
              <ReactECharts option={authTrendOption} style={{ height: 320 }} />
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={6}>
          <Card sx={{ borderRadius: 4, height: '100%' }}>
            <CardContent sx={{ height: '100%' }}>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 2 }}>
                在线用户分布
              </Typography>
              <ReactECharts option={onlineDistributionOption} style={{ height: 320 }} />
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12}>
          <Card sx={{ borderRadius: 4 }}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 2 }}>
                流量统计（近 24 小时）
              </Typography>
              <ReactECharts option={trafficOption} style={{ height: 360 }} />
            </CardContent>
          </Card>
        </Grid>
      </Grid>

    </Box>
  );
};


export default Dashboard;
