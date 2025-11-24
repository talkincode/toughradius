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
import { useMemo } from 'react';
import { useTranslate } from 'react-admin';
import { useApiQuery } from '../hooks/useApiQuery';

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
  auth_trend: DashboardAuthTrendPoint[];
  traffic_24h: DashboardTrafficPoint[];
  profile_distribution: DashboardProfileSlice[];
}

interface DashboardAuthTrendPoint {
  date: string;
  count: number;
}

interface DashboardTrafficPoint {
  hour: string;
  upload_gb: number;
  download_gb: number;
}

interface DashboardProfileSlice {
  profile_id: number;
  profile_name: string;
  value: number;
}

const emptyStats: DashboardStats = {
  total_users: 0,
  online_users: 0,
  today_auth_count: 0,
  today_acct_count: 0,
  total_profiles: 0,
  disabled_users: 0,
  expired_users: 0,
  today_input_gb: 0,
  today_output_gb: 0,
  auth_trend: [],
  traffic_24h: [],
  profile_distribution: [],
};

const Dashboard = () => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const translate = useTranslate();
  const { data: statsPayload, isFetching } = useApiQuery<DashboardStats>({
    path: '/dashboard/stats',
    queryKey: ['dashboard', 'stats'],
    staleTime: 60 * 1000,
    refetchInterval: 60 * 1000,
    retry: 1,
  });

  const stats = statsPayload ?? emptyStats;

  const dateFormatter = useMemo(() => new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric' }), []);
  const hourFormatter = useMemo(
    () => new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit', hour12: false }),
    [],
  );

  const numberFormatter = useMemo(() => new Intl.NumberFormat(), []);

  const onlineRatio =
    stats.total_users > 0 ? Math.min((stats.online_users / stats.total_users) * 100, 100) : 0;

  const authTrendLabels = useMemo(
    () => (stats.auth_trend ?? []).map((point) => {
      const normalized = `${point.date}T00:00:00`;
      const parsed = new Date(normalized);
      return Number.isNaN(parsed.getTime()) ? point.date : dateFormatter.format(parsed);
    }),
    [stats.auth_trend, dateFormatter],
  );

  const authTrendSeries = useMemo(() => (stats.auth_trend ?? []).map((point) => point.count), [stats.auth_trend]);

  const profileSlices = useMemo(() => (
    stats.profile_distribution ?? []
  ).map((item) => ({
    value: item.value,
    name: item.profile_name?.trim() || translate('dashboard.profile_unassigned'),
  })), [stats.profile_distribution, translate]);

  const trafficData = useMemo(() => {
    const formatHourLabel = (value: string) => {
      const normalized = value.replace(' ', 'T');
      const parsed = new Date(normalized);
      return Number.isNaN(parsed.getTime()) ? value : hourFormatter.format(parsed);
    };

    return {
      labels: (stats.traffic_24h ?? []).map((item) => formatHourLabel(item.hour)),
      upload: (stats.traffic_24h ?? []).map((item) => item.upload_gb),
      download: (stats.traffic_24h ?? []).map((item) => item.download_gb),
    };
  }, [stats.traffic_24h, hourFormatter]);

  const statCards = [
    {
      title: translate('dashboard.total_users'),
      value: numberFormatter.format(stats.total_users),
      icon: <PeopleAltOutlinedIcon fontSize="large" />,
      accent: theme.palette.primary.main,
      highlights: [
        { label: translate('dashboard.disabled'), value: stats.disabled_users },
        { label: translate('dashboard.expired'), value: stats.expired_users },
      ],
    },
    {
      title: translate('dashboard.online_users'),
      value: numberFormatter.format(stats.online_users),
      icon: <OnlinePredictionOutlinedIcon fontSize="large" />,
      accent: '#34d399',
      highlights: [{ label: translate('dashboard.total_profiles'), value: stats.total_profiles }],
    },
    {
      title: translate('dashboard.today_auth'),
      value: numberFormatter.format(stats.today_auth_count),
      icon: <VerifiedUserOutlinedIcon fontSize="large" />,
      accent: theme.palette.secondary.main,
      highlights: [{ label: translate('dashboard.acct_records'), value: stats.today_acct_count }],
    },
    {
      title: translate('dashboard.today_traffic'),
      value: `↑ ${stats.today_input_gb.toFixed(2)} GB`,
      secondaryValue: `↓ ${stats.today_output_gb.toFixed(2)} GB`,
      icon: <SwapVertOutlinedIcon fontSize="large" />,
      accent: '#f97316',
      highlights: [{ label: translate('dashboard.unit_gb'), value: 'GB' }],
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
        data: authTrendLabels,
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
          name: translate('dashboard.auth_trend').replace(/（.*?）/, '').replace(/ \(.*?\)/, ''),
          type: 'line',
          smooth: true,
          symbol: 'circle',
          symbolSize: 10,
          data: authTrendSeries,
          lineStyle: { width: 4 },
          itemStyle: { color: theme.palette.primary.main },
          areaStyle: {
            color: alpha(theme.palette.primary.main, 0.18),
          },
        },
      ],
    }),
    [theme, authTrendLabels, authTrendSeries, translate],
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
          name: translate('dashboard.online_users'),
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
          data: profileSlices,
          color: [
            theme.palette.primary.main,
            theme.palette.secondary.main,
            '#34d399',
            '#facc15',
          ],
        },
      ],
    }),
    [theme, profileSlices, translate],
  );

  const trafficOption = useMemo(
    () => ({
      backgroundColor: 'transparent',
      tooltip: { trigger: 'axis' },
      legend: {
        data: [translate('dashboard.upload'), translate('dashboard.download')],
        top: 0,
        textStyle: { color: alpha(theme.palette.text.primary, 0.7) },
      },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: {
        type: 'category',
        data: trafficData.labels,
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
          name: translate('dashboard.upload'),
          type: 'bar',
          stack: 'traffic',
          emphasis: { focus: 'series' },
          data: trafficData.upload,
          color: alpha(theme.palette.secondary.main, 0.7),
        },
        {
          name: translate('dashboard.download'),
          type: 'bar',
          stack: 'traffic',
          emphasis: { focus: 'series' },
          data: trafficData.download,
          color: theme.palette.primary.main,
        },
      ],
    }),
    [theme, trafficData, translate],
  );

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
      {isFetching && (
        <LinearProgress
          sx={{
            position: 'sticky',
            top: 0,
            left: 0,
            right: 0,
            borderRadius: 2,
          }}
        />
      )}
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
              <Chip label={translate('dashboard.title')} color="primary" sx={{ mb: 2, fontWeight: 600 }} />
              <Typography variant="body1" sx={{ color: 'text.secondary', maxWidth: 520 }}>
                {translate('dashboard.subtitle')}
              </Typography>

              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={4} sx={{ mt: 3 }}>
                <Box>
                  <Typography variant="subtitle2" color="text.secondary">
                    {translate('dashboard.today_auth')}
                  </Typography>
                  <Typography variant="h5" sx={{ fontWeight: 700, color: 'text.primary' }}>
                    {numberFormatter.format(stats.today_auth_count)}
                  </Typography>
                </Box>
                <Box>
                  <Typography variant="subtitle2" color="text.secondary">
                    {translate('dashboard.today_acct')}
                  </Typography>
                  <Typography variant="h5" sx={{ fontWeight: 700, color: 'text.primary' }}>
                    {numberFormatter.format(stats.today_acct_count)}
                  </Typography>
                </Box>
              </Stack>
            </Box>

            <Box sx={{ minWidth: 260 }}>
              <Typography variant="subtitle2" color="text.secondary">
                {translate('dashboard.online_ratio')}
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
                  {translate('dashboard.online_count')} {stats.online_users}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {translate('dashboard.total_count')} {stats.total_users}
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
                {translate('dashboard.auth_trend')}
              </Typography>
              <ReactECharts option={authTrendOption} style={{ height: 320 }} />
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={6}>
          <Card sx={{ borderRadius: 4, height: '100%' }}>
            <CardContent sx={{ height: '100%' }}>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 2 }}>
                {translate('dashboard.online_distribution')}
              </Typography>
              <ReactECharts option={onlineDistributionOption} style={{ height: 320 }} />
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12}>
          <Card sx={{ borderRadius: 4 }}>
            <CardContent>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 2 }}>
                {translate('dashboard.traffic_stats')}
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
