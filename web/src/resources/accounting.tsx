import {
  List,
  Datagrid,
  TextField,
  DateField,
  FunctionField,
  Show,
  useRecordContext,
  useTranslate,
  ExportButton,
  TopToolbar,
  ListButton,
  useRefresh,
  useNotify,
  useListContext,
  SortButton,
  RaRecord,
} from 'react-admin';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Chip,
  Stack,
  alpha,
  Avatar,
  IconButton,
  Tooltip,
  Skeleton,
  useTheme,
  useMediaQuery,
  TextField as MuiTextField,
} from '@mui/material';
import {
  Router as DeviceIcon,
  AccessTime as TimeIcon,
  DataUsage as TrafficIcon,
  Info as InfoIcon,
  Wifi as OnlineIcon,
  WifiOff as OfflineIcon,
  ContentCopy as CopyIcon,
  Refresh as RefreshIcon,
  ArrowBack as BackIcon,
  CloudUpload as UploadIcon,
  CloudDownload as DownloadIcon,
  Speed as SpeedIcon,
  SignalCellularAlt as SignalIcon,
  CheckCircle as SuccessIcon,
  Warning as WarningIcon,
  Print as PrintIcon,
  FilterList as FilterIcon,
  Search as SearchIcon,
  Clear as ClearIcon,
} from '@mui/icons-material';
import { ReactNode, useMemo, useCallback, useState, useEffect } from 'react';
import { ServerPagination, ActiveFilters } from '../components';

const LARGE_LIST_PER_PAGE = 50;

interface AccountingRecord extends RaRecord {
  acct_session_id?: string;
  username?: string;
  nas_id?: string;
  nas_addr?: string;
  nas_paddr?: string;
  nas_port?: string | number;
  nas_class?: string;
  nas_port_id?: string;
  nas_port_type?: number;
  service_type?: number;
  session_timeout?: number;
  framed_ipaddr?: string;
  framed_netmask?: string;
  framed_ipv6_prefix?: string;
  framed_ipv6_address?: string;
  delegated_ipv6_prefix?: string;
  mac_addr?: string;
  acct_session_time?: number;
  acct_input_total?: number;
  acct_output_total?: number;
  acct_input_packets?: number;
  acct_output_packets?: number;
  acct_start_time?: string | number;
  acct_stop_time?: string | number;
  last_update?: string | number;
  acct_terminate_cause?: string;
  acct_unique_id?: string;
}

const formatDuration = (seconds?: number): string => {
  if (seconds === undefined || seconds === null) {
    return '-';
  }
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;
  const parts = [];
  if (hours) {
    parts.push(`${hours}h`);
  }
  if (minutes) {
    parts.push(`${minutes}m`);
  }
  parts.push(`${secs}s`);
  return parts.join(' ');
};

const formatBytes = (bytes?: number): string => {
  if (bytes === undefined || bytes === null) {
    return '-';
  }
  if (bytes === 0) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = bytes;
  let index = 0;
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024;
    index += 1;
  }
  return `${parseFloat(value.toFixed(2))} ${units[index]}`;
};

const formatTimestamp = (value?: string | number): string => {
  if (!value) {
    return '-';
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return '-';
  }
  return date.toLocaleString();
};

const NasPortField = () => {
  const translate = useTranslate();
  return (
    <FunctionField
      source="nas_port"
      label={translate('resources.radius/accounting.fields.nas_port')}
      render={(record: AccountingRecord) => {
        if (!record?.nas_port) {
          return '-';
        }
        return <Chip label={String(record.nas_port)} size="small" variant="outlined" />;
      }}
    />
  );
};

const SessionDurationField = () => {
  const translate = useTranslate();
  return (
    <FunctionField
      source="acct_session_time"
      label={translate('resources.radius/accounting.fields.session_time')}
      render={(record: AccountingRecord) => formatDuration(record.acct_session_time)}
    />
  );
};

// ============ 美化的详情组件 ============

interface DetailItemProps {
  label: string;
  value?: ReactNode;
  highlight?: boolean;
}

const DetailItem = ({ label, value, highlight = false }: DetailItemProps) => (
  <Box
    sx={{
      display: 'flex',
      flexDirection: 'column',
      gap: 0.5,
      p: 1.5,
      borderRadius: 1.5,
      backgroundColor: theme =>
        highlight
          ? alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.15 : 0.06)
          : theme.palette.mode === 'dark'
          ? 'rgba(255, 255, 255, 0.02)'
          : 'rgba(0, 0, 0, 0.02)',
      border: theme =>
        highlight
          ? `1px solid ${alpha(theme.palette.primary.main, 0.3)}`
          : `1px solid ${theme.palette.divider}`,
      transition: 'all 0.2s ease',
      '&:hover': {
        backgroundColor: theme =>
          highlight
            ? alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.2 : 0.08)
            : theme.palette.mode === 'dark'
            ? 'rgba(255, 255, 255, 0.04)'
            : 'rgba(0, 0, 0, 0.03)',
      },
    }}
  >
    <Typography
      variant="caption"
      sx={{
        color: 'text.secondary',
        fontWeight: 500,
        fontSize: '0.85rem',
        textTransform: 'uppercase',
        letterSpacing: '0.5px',
      }}
    >
      {label}
    </Typography>
    <Typography
      variant="body2"
      sx={{
        fontWeight: highlight ? 600 : 500,
        color: highlight ? 'primary.main' : 'text.primary',
        wordBreak: 'break-word',
        fontSize: '1rem',
        lineHeight: 1.5,
      }}
    >
      {value ?? <span style={{ color: 'inherit', opacity: 0.4 }}>-</span>}
    </Typography>
  </Box>
);

interface DetailSectionCardProps {
  title: string;
  description?: string;
  icon: ReactNode;
  children: ReactNode;
  color?: 'primary' | 'success' | 'warning' | 'info' | 'error';
}

const DetailSectionCard = ({
  title,
  description,
  icon,
  children,
  color = 'primary',
}: DetailSectionCardProps) => (
  <Card
    elevation={0}
    sx={{
      borderRadius: 3,
      border: theme => `1px solid ${theme.palette.divider}`,
      overflow: 'hidden',
      transition: 'all 0.2s ease',
      '&:hover': {
        boxShadow: theme =>
          theme.palette.mode === 'dark'
            ? '0 4px 20px rgba(0, 0, 0, 0.3)'
            : '0 4px 20px rgba(0, 0, 0, 0.08)',
      },
    }}
  >
    <Box
      sx={{
        px: 2.5,
        py: 2,
        backgroundColor: theme =>
          alpha(
            theme.palette[color].main,
            theme.palette.mode === 'dark' ? 0.15 : 0.06
          ),
        borderBottom: theme =>
          `1px solid ${alpha(theme.palette[color].main, 0.2)}`,
      }}
    >
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: 36,
            height: 36,
            borderRadius: 2,
            backgroundColor: theme =>
              alpha(theme.palette[color].main, theme.palette.mode === 'dark' ? 0.3 : 0.15),
            color: `${color}.main`,
          }}
        >
          {icon}
        </Box>
        <Box>
          <Typography
            variant="subtitle1"
            sx={{
              fontWeight: 600,
              color: `${color}.main`,
              fontSize: '1.1rem',
            }}
          >
            {title}
          </Typography>
          {description && (
            <Typography
              variant="body2"
              sx={{
                color: 'text.secondary',
                fontSize: '0.9rem',
                mt: 0.25,
              }}
            >
              {description}
            </Typography>
          )}
        </Box>
      </Box>
    </Box>
    <CardContent sx={{ p: 2.5 }}>{children}</CardContent>
  </Card>
);

// 流量统计卡片 - 使用特殊样式展示
interface TrafficStatProps {
  label: string;
  value: string;
  icon?: ReactNode;
  color?: 'primary' | 'success' | 'warning' | 'info' | 'error';
  subValue?: string;
}

const TrafficStat = ({ label, value, icon, color = 'primary', subValue }: TrafficStatProps) => (
  <Box
    sx={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      p: 2.5,
      borderRadius: 3,
      backgroundColor: theme =>
        alpha(theme.palette[color].main, theme.palette.mode === 'dark' ? 0.15 : 0.08),
      border: theme => `1px solid ${alpha(theme.palette[color].main, 0.2)}`,
      minWidth: 140,
      flex: 1,
      transition: 'all 0.2s ease',
      '&:hover': {
        transform: 'translateY(-2px)',
        boxShadow: theme =>
          `0 8px 24px ${alpha(theme.palette[color].main, 0.2)}`,
      },
    }}
  >
    {icon && (
      <Box
        sx={{
          mb: 1,
          color: `${color}.main`,
          opacity: 0.8,
        }}
      >
        {icon}
      </Box>
    )}
    <Typography
      variant="caption"
      sx={{
        color: 'text.secondary',
        fontWeight: 500,
        fontSize: '0.8rem',
        textTransform: 'uppercase',
        letterSpacing: '0.5px',
        mb: 0.5,
        textAlign: 'center',
      }}
    >
      {label}
    </Typography>
    <Typography
      variant="h5"
      sx={{
        fontWeight: 700,
        color: `${color}.main`,
        fontSize: '1.4rem',
      }}
    >
      {value}
    </Typography>
    {subValue && (
      <Typography
        variant="caption"
        sx={{
          color: 'text.secondary',
          fontSize: '0.8rem',
          mt: 0.5,
        }}
      >
        {subValue}
      </Typography>
    )}
  </Box>
);

// 空态组件
interface EmptyStateProps {
  message?: string;
}

const EmptyValue = ({ message = '暂无数据' }: EmptyStateProps) => (
  <Box
    sx={{
      display: 'flex',
      alignItems: 'center',
      gap: 0.5,
      color: 'text.disabled',
      fontStyle: 'italic',
      fontSize: '0.85rem',
    }}
  >
    <Typography variant="body2" sx={{ opacity: 0.6 }}>
      {message}
    </Typography>
  </Box>
);

// 获取终止原因的颜色和图标
const getTerminateCauseInfo = (cause?: string) => {
  if (!cause) return { color: 'default' as const, icon: null, label: '未知' };
  
  const normalCauses = ['User-Request', 'Session-Timeout', 'Idle-Timeout'];
  const errorCauses = ['Admin-Reset', 'Lost-Carrier', 'Port-Error', 'NAS-Error'];
  
  if (normalCauses.includes(cause)) {
    return { color: 'success' as const, icon: <SuccessIcon sx={{ fontSize: '1rem' }} />, label: cause };
  }
  if (errorCauses.includes(cause)) {
    return { color: 'error' as const, icon: <WarningIcon sx={{ fontSize: '1rem' }} />, label: cause };
  }
  return { color: 'warning' as const, icon: null, label: cause };
};

// 顶部概览卡片
const AccountingHeaderCard = () => {
  const record = useRecordContext<AccountingRecord>();
  const translate = useTranslate();
  const notify = useNotify();
  const refresh = useRefresh();

  const handleCopy = useCallback((text: string, label: string) => {
    navigator.clipboard.writeText(text);
    notify(`${label} 已复制到剪贴板`, { type: 'info' });
  }, [notify]);

  const handleRefresh = useCallback(() => {
    refresh();
    notify('数据已刷新', { type: 'info' });
  }, [refresh, notify]);

  if (!record) return null;

  const isOnline = !record.acct_stop_time;
  const totalTraffic = (record.acct_input_total ?? 0) + (record.acct_output_total ?? 0);
  const terminateInfo = getTerminateCauseInfo(record.acct_terminate_cause);

  return (
    <Card
      elevation={0}
      sx={{
        borderRadius: 4,
        background: theme =>
          theme.palette.mode === 'dark'
            ? isOnline
              ? `linear-gradient(135deg, ${alpha(theme.palette.success.dark, 0.3)} 0%, ${alpha(theme.palette.info.dark, 0.2)} 100%)`
              : `linear-gradient(135deg, ${alpha(theme.palette.grey[800], 0.5)} 0%, ${alpha(theme.palette.grey[700], 0.3)} 100%)`
            : isOnline
            ? `linear-gradient(135deg, ${alpha(theme.palette.success.main, 0.1)} 0%, ${alpha(theme.palette.info.main, 0.08)} 100%)`
            : `linear-gradient(135deg, ${alpha(theme.palette.grey[400], 0.15)} 0%, ${alpha(theme.palette.grey[300], 0.1)} 100%)`,
        border: theme => `1px solid ${alpha(isOnline ? theme.palette.success.main : theme.palette.grey[500], 0.2)}`,
        overflow: 'hidden',
        position: 'relative',
      }}
    >
      {/* 装饰背景 */}
      <Box
        sx={{
          position: 'absolute',
          top: -50,
          right: -50,
          width: 200,
          height: 200,
          borderRadius: '50%',
          background: theme => alpha(isOnline ? theme.palette.success.main : theme.palette.grey[500], 0.1),
          pointerEvents: 'none',
        }}
      />

      <CardContent sx={{ p: 3, position: 'relative', zIndex: 1 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 3 }}>
          {/* 左侧：用户信息 */}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <Avatar
              sx={{
                width: 64,
                height: 64,
                bgcolor: isOnline ? 'success.main' : 'grey.500',
                fontSize: '1.5rem',
                fontWeight: 700,
                boxShadow: theme => `0 4px 14px ${alpha(isOnline ? theme.palette.success.main : theme.palette.grey[500], 0.4)}`,
              }}
            >
              {record.username?.charAt(0).toUpperCase() || 'U'}
            </Avatar>
            <Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
                <Typography variant="h5" sx={{ fontWeight: 700, color: 'text.primary' }}>
                  {record.username || <EmptyValue message="未知用户" />}
                </Typography>
                {isOnline ? (
                  <Chip
                    icon={<OnlineIcon sx={{ fontSize: '1rem !important' }} />}
                    label="在线"
                    size="small"
                    color="success"
                    sx={{ fontWeight: 600, height: 24 }}
                  />
                ) : (
                  <Chip
                    icon={<OfflineIcon sx={{ fontSize: '1rem !important' }} />}
                    label="已结束"
                    size="small"
                    color="default"
                    variant="outlined"
                    sx={{ fontWeight: 600, height: 24 }}
                  />
                )}
              </Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <Typography variant="body2" color="text.secondary">
                  {record.framed_ipaddr || '未分配 IP'}
                </Typography>
                {record.framed_ipaddr && (
                  <Tooltip title="复制 IP 地址">
                    <IconButton
                      size="small"
                      onClick={() => handleCopy(record.framed_ipaddr!, 'IP 地址')}
                      sx={{ p: 0.5 }}
                    >
                      <CopyIcon sx={{ fontSize: '0.9rem' }} />
                    </IconButton>
                  </Tooltip>
                )}
              </Box>
              {record.acct_session_id && (
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 0.5 }}>
                  <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace' }}>
                    会话: {record.acct_session_id}
                  </Typography>
                  <Tooltip title="复制会话 ID">
                    <IconButton
                      size="small"
                      onClick={() => handleCopy(record.acct_session_id!, '会话 ID')}
                      sx={{ p: 0.5 }}
                    >
                      <CopyIcon sx={{ fontSize: '0.75rem' }} />
                    </IconButton>
                  </Tooltip>
                </Box>
              )}
            </Box>
          </Box>

          {/* 右侧：操作按钮 */}
          <Box className="no-print" sx={{ display: 'flex', gap: 1 }}>
            <Tooltip title="打印详情">
              <IconButton
                onClick={() => window.print()}
                sx={{
                  bgcolor: theme => alpha(theme.palette.info.main, 0.1),
                  '&:hover': {
                    bgcolor: theme => alpha(theme.palette.info.main, 0.2),
                  },
                }}
              >
                <PrintIcon />
              </IconButton>
            </Tooltip>
            <Tooltip title="刷新数据">
              <IconButton
                onClick={handleRefresh}
                sx={{
                  bgcolor: theme => alpha(theme.palette.primary.main, 0.1),
                  '&:hover': {
                    bgcolor: theme => alpha(theme.palette.primary.main, 0.2),
                  },
                }}
              >
                <RefreshIcon />
              </IconButton>
            </Tooltip>
            <ListButton
              label=""
              icon={<BackIcon />}
              sx={{
                minWidth: 'auto',
                bgcolor: theme => alpha(theme.palette.grey[500], 0.1),
                '&:hover': {
                  bgcolor: theme => alpha(theme.palette.grey[500], 0.2),
                },
              }}
            />
          </Box>
        </Box>

        {/* 快速统计 */}
        <Box
          sx={{
            display: 'grid',
            gap: 2,
            gridTemplateColumns: {
              xs: 'repeat(2, 1fr)',
              sm: 'repeat(4, 1fr)',
              lg: 'repeat(5, 1fr)',
            },
          }}
        >
          <Box
            sx={{
              p: 2,
              borderRadius: 2,
              bgcolor: theme => alpha(theme.palette.background.paper, 0.8),
              backdropFilter: 'blur(8px)',
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
              <TimeIcon sx={{ fontSize: '1.1rem', color: 'warning.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.radius/accounting.fields.session_time')}
              </Typography>
            </Box>
            <Typography variant="h6" sx={{ fontWeight: 700 }}>
              {formatDuration(record.acct_session_time)}
            </Typography>
          </Box>

          <Box
            sx={{
              p: 2,
              borderRadius: 2,
              bgcolor: theme => alpha(theme.palette.background.paper, 0.8),
              backdropFilter: 'blur(8px)',
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
              <UploadIcon sx={{ fontSize: '1.1rem', color: 'info.main' }} />
              <Typography variant="caption" color="text.secondary">
                上传流量
              </Typography>
            </Box>
            <Typography variant="h6" sx={{ fontWeight: 700, color: 'info.main' }}>
              {formatBytes(record.acct_input_total)}
            </Typography>
          </Box>

          <Box
            sx={{
              p: 2,
              borderRadius: 2,
              bgcolor: theme => alpha(theme.palette.background.paper, 0.8),
              backdropFilter: 'blur(8px)',
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
              <DownloadIcon sx={{ fontSize: '1.1rem', color: 'warning.main' }} />
              <Typography variant="caption" color="text.secondary">
                下载流量
              </Typography>
            </Box>
            <Typography variant="h6" sx={{ fontWeight: 700, color: 'warning.main' }}>
              {formatBytes(record.acct_output_total)}
            </Typography>
          </Box>

          <Box
            sx={{
              p: 2,
              borderRadius: 2,
              bgcolor: theme => alpha(theme.palette.background.paper, 0.8),
              backdropFilter: 'blur(8px)',
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
              <SpeedIcon sx={{ fontSize: '1.1rem', color: 'success.main' }} />
              <Typography variant="caption" color="text.secondary">
                总流量
              </Typography>
            </Box>
            <Typography variant="h6" sx={{ fontWeight: 700, color: 'success.main' }}>
              {formatBytes(totalTraffic)}
            </Typography>
          </Box>

          {!isOnline && record.acct_terminate_cause && (
            <Box
              sx={{
                p: 2,
                borderRadius: 2,
                bgcolor: theme => alpha(theme.palette.background.paper, 0.8),
                backdropFilter: 'blur(8px)',
              }}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                {terminateInfo.icon || <InfoIcon sx={{ fontSize: '1.1rem', color: `${terminateInfo.color}.main` }} />}
                <Typography variant="caption" color="text.secondary">
                  终止原因
                </Typography>
              </Box>
              <Chip
                label={terminateInfo.label}
                size="small"
                color={terminateInfo.color}
                sx={{ fontWeight: 600 }}
              />
            </Box>
          )}
        </Box>
      </CardContent>
    </Card>
  );
};

// 打印样式
const printStyles = `
  @media print {
    body * {
      visibility: hidden;
    }
    .printable-content, .printable-content * {
      visibility: visible;
    }
    .printable-content {
      position: absolute;
      left: 0;
      top: 0;
      width: 100%;
      padding: 20px !important;
    }
    .no-print {
      display: none !important;
    }
  }
`;

const AccountingDetails = () => {
  const record = useRecordContext<AccountingRecord>();
  const translate = useTranslate();
  if (!record) {
    return null;
  }

  const totalTraffic =
    (record.acct_input_total ?? 0) + (record.acct_output_total ?? 0);
  const isOnline = !record.acct_stop_time;

  return (
    <>
      <style>{printStyles}</style>
      <Box className="printable-content" sx={{ width: '100%', p: { xs: 2, sm: 3, md: 4 } }}>
      <Stack spacing={3}>
        {/* 顶部概览卡片 */}
        <AccountingHeaderCard />

        {/* 设备信息 */}
        <DetailSectionCard
          title={translate('resources.radius/accounting.sections.device')}
          description={translate('resources.radius/accounting.sections.device_desc')}
          icon={<DeviceIcon />}
          color="info"
        >
          <Box
            sx={{
              display: 'grid',
              gap: 2,
              gridTemplateColumns: {
                xs: 'repeat(1, 1fr)',
                sm: 'repeat(2, 1fr)',
                md: 'repeat(3, 1fr)',
              },
            }}
          >
            <DetailItem
              label={translate('resources.radius/accounting.fields.framed_ipaddr')}
              value={
                record.framed_ipaddr ? (
                  <Chip
                    label={record.framed_ipaddr}
                    size="small"
                    color="info"
                    variant="outlined"
                    sx={{ fontFamily: 'monospace' }}
                  />
                ) : (
                  <EmptyValue message="未分配" />
                )
              }
              highlight
            />
            <DetailItem
              label={translate('resources.radius/accounting.fields.framed_netmask')}
              value={record.framed_netmask || <EmptyValue />}
            />
            <DetailItem
              label={translate('resources.radius/accounting.fields.mac_addr')}
              value={
                record.mac_addr ? (
                  <Chip
                    label={record.mac_addr}
                    size="small"
                    variant="outlined"
                    sx={{ fontFamily: 'monospace', fontSize: '0.8rem' }}
                  />
                ) : (
                  <EmptyValue message="未获取" />
                )
              }
            />
            <DetailItem
              label={translate('resources.radius/accounting.fields.framed_ipv6_address')}
              value={record.framed_ipv6_address || <EmptyValue message="未配置" />}
            />
            <DetailItem
              label={translate('resources.radius/accounting.fields.framed_ipv6_prefix')}
              value={record.framed_ipv6_prefix || <EmptyValue message="未配置" />}
            />
          </Box>
        </DetailSectionCard>

        {/* NAS 设备信息 */}
        <DetailSectionCard
          title={translate('resources.radius/accounting.sections.overview')}
          description={translate('resources.radius/accounting.sections.overview_desc')}
          icon={<SignalIcon />}
          color="primary"
        >
          <Box
            sx={{
              display: 'grid',
              gap: 2,
              gridTemplateColumns: {
                xs: 'repeat(1, 1fr)',
                sm: 'repeat(2, 1fr)',
                md: 'repeat(3, 1fr)',
              },
            }}
          >
            <DetailItem
              label={translate('resources.radius/accounting.fields.nas_addr')}
              value={
                record.nas_addr ? (
                  <Chip
                    label={record.nas_addr}
                    size="small"
                    color="primary"
                    variant="outlined"
                    sx={{ fontFamily: 'monospace' }}
                  />
                ) : (
                  <EmptyValue />
                )
              }
              highlight
            />
            <DetailItem
              label={translate('resources.radius/accounting.fields.nas_id')}
              value={record.nas_id || <EmptyValue />}
            />
            <DetailItem
              label={translate('resources.radius/accounting.fields.nas_port')}
              value={record.nas_port?.toString() || <EmptyValue />}
            />
            <DetailItem
              label={translate('resources.radius/accounting.fields.service_type')}
              value={record.service_type?.toString() || <EmptyValue />}
            />
          </Box>
        </DetailSectionCard>

        {/* 会话时间 */}
        <DetailSectionCard
          title={translate('resources.radius/accounting.sections.timing')}
          description={translate('resources.radius/accounting.sections.timing_desc')}
          icon={<TimeIcon />}
          color="warning"
        >
          <Box
            sx={{
              display: 'grid',
              gap: 2,
              gridTemplateColumns: {
                xs: 'repeat(1, 1fr)',
                sm: 'repeat(2, 1fr)',
                md: 'repeat(3, 1fr)',
              },
            }}
          >
            <DetailItem
              label={translate('resources.radius/accounting.fields.acct_start_time')}
              value={formatTimestamp(record.acct_start_time)}
              highlight
            />
            <DetailItem
              label={translate('resources.radius/accounting.fields.acct_stop_time')}
              value={
                record.acct_stop_time ? (
                  <Chip
                    label={formatTimestamp(record.acct_stop_time)}
                    size="small"
                    color="error"
                    variant="outlined"
                    sx={{ fontSize: '0.75rem' }}
                  />
                ) : (
                  <Chip
                    icon={<OnlineIcon sx={{ fontSize: '0.9rem !important' }} />}
                    label="在线中"
                    size="small"
                    color="success"
                    sx={{ fontWeight: 600 }}
                  />
                )
              }
              highlight
            />
            <DetailItem
              label={translate('resources.radius/accounting.fields.session_time')}
              value={
                <Chip
                  label={formatDuration(record.acct_session_time)}
                  size="small"
                  color="warning"
                  sx={{ fontWeight: 600 }}
                />
              }
              highlight
            />
            <DetailItem
              label={translate('resources.radius/accounting.fields.session_timeout')}
              value={
                record.session_timeout !== undefined && record.session_timeout !== null
                  ? `${record.session_timeout}s`
                  : <EmptyValue message="无限制" />
              }
            />
            <DetailItem
              label={translate('resources.radius/accounting.fields.last_update')}
              value={formatTimestamp(record.last_update)}
            />
          </Box>
        </DetailSectionCard>

        {/* 流量统计 - 特殊展示 */}
        <DetailSectionCard
          title={translate('resources.radius/accounting.sections.traffic')}
          description={translate('resources.radius/accounting.sections.traffic_desc')}
          icon={<TrafficIcon />}
          color="success"
        >
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            {/* 流量统计卡片 */}
            <Box
              sx={{
                display: 'flex',
                flexWrap: 'wrap',
                gap: 2,
                justifyContent: 'stretch',
              }}
            >
              <TrafficStat
                label={translate('resources.radius/accounting.fields.acct_input_total')}
                value={formatBytes(record.acct_input_total)}
                icon={<UploadIcon />}
                color="info"
                subValue={`${record.acct_input_packets?.toLocaleString() ?? 0} 包`}
              />
              <TrafficStat
                label={translate('resources.radius/accounting.fields.acct_output_total')}
                value={formatBytes(record.acct_output_total)}
                icon={<DownloadIcon />}
                color="warning"
                subValue={`${record.acct_output_packets?.toLocaleString() ?? 0} 包`}
              />
              <TrafficStat
                label={translate('resources.radius/accounting.fields.total_traffic')}
                value={formatBytes(totalTraffic)}
                icon={<SpeedIcon />}
                color="success"
                subValue={`${((record.acct_input_packets ?? 0) + (record.acct_output_packets ?? 0)).toLocaleString()} 包`}
              />
            </Box>
          </Box>
        </DetailSectionCard>

        {/* 会话详情 */}
        <DetailSectionCard
          title={translate('resources.radius/accounting.sections.session_details')}
          description={translate('resources.radius/accounting.sections.session_details_desc')}
          icon={<InfoIcon />}
          color="primary"
        >
          <Box
            sx={{
              display: 'grid',
              gap: 2,
              gridTemplateColumns: {
                xs: 'repeat(1, 1fr)',
                sm: 'repeat(2, 1fr)',
                md: 'repeat(4, 1fr)',
              },
            }}
          >
            <DetailItem
              label={translate('resources.radius/accounting.fields.nas_class')}
              value={record.nas_class || <EmptyValue />}
            />
            <DetailItem
              label={translate('resources.radius/accounting.fields.nas_port_id')}
              value={record.nas_port_id || <EmptyValue />}
            />
            <DetailItem
              label={translate('resources.radius/accounting.fields.nas_port_type')}
              value={record.nas_port_type?.toString() || <EmptyValue />}
            />
            <DetailItem
              label={translate('resources.radius/accounting.fields.acct_terminate_cause')}
              value={
                record.acct_terminate_cause ? (
                  (() => {
                    const info = getTerminateCauseInfo(record.acct_terminate_cause);
                    return (
                      <Chip
                        icon={info.icon || undefined}
                        label={info.label}
                        size="small"
                        color={info.color}
                        variant="outlined"
                      />
                    );
                  })()
                ) : (
                  isOnline ? (
                    <Chip
                      icon={<OnlineIcon sx={{ fontSize: '0.9rem !important' }} />}
                      label="会话进行中"
                      size="small"
                      color="success"
                      variant="outlined"
                    />
                  ) : (
                    <EmptyValue message="未记录" />
                  )
                )
              }
            />
          </Box>
        </DetailSectionCard>
      </Stack>
    </Box>
    </>
  );
};

// ============ 列表加载骨架屏 ============

const AccountingListSkeleton = ({ rows = 10 }: { rows?: number }) => (
  <Box sx={{ width: '100%' }}>
    {/* 搜索区域骨架屏 */}
    <Card
      elevation={0}
      sx={{
        mb: 2,
        borderRadius: 2,
        border: theme => `1px solid ${theme.palette.divider}`,
      }}
    >
      <CardContent sx={{ p: 2 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
          <Skeleton variant="rectangular" width={24} height={24} />
          <Skeleton variant="text" width={100} height={24} />
        </Box>
        <Box
          sx={{
            display: 'grid',
            gap: 2,
            gridTemplateColumns: {
              xs: '1fr',
              sm: 'repeat(2, 1fr)',
              md: 'repeat(3, 1fr)',
              lg: 'repeat(6, 1fr)',
            },
          }}
        >
          {[...Array(6)].map((_, i) => (
            <Skeleton key={i} variant="rectangular" height={40} sx={{ borderRadius: 1 }} />
          ))}
        </Box>
      </CardContent>
    </Card>

    {/* 表格骨架屏 */}
    <Card
      elevation={0}
      sx={{
        borderRadius: 2,
        border: theme => `1px solid ${theme.palette.divider}`,
        overflow: 'hidden',
      }}
    >
      {/* 表头 */}
      <Box
        sx={{
          display: 'grid',
          gridTemplateColumns: 'repeat(11, 1fr) 80px',
          gap: 1,
          p: 2,
          bgcolor: theme =>
            theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.02)',
          borderBottom: theme => `1px solid ${theme.palette.divider}`,
        }}
      >
        {[...Array(12)].map((_, i) => (
          <Skeleton key={i} variant="text" height={20} width={i === 11 ? 60 : '80%'} />
        ))}
      </Box>

      {/* 表格行 */}
      {[...Array(rows)].map((_, rowIndex) => (
        <Box
          key={rowIndex}
          sx={{
            display: 'grid',
            gridTemplateColumns: 'repeat(11, 1fr) 80px',
            gap: 1,
            p: 2,
            borderBottom: theme => `1px solid ${theme.palette.divider}`,
          }}
        >
          {[...Array(12)].map((_, colIndex) => (
            <Skeleton
              key={colIndex}
              variant="text"
              height={18}
              width={colIndex === 11 ? 40 : `${60 + Math.random() * 30}%`}
            />
          ))}
        </Box>
      ))}

      {/* 分页骨架屏 */}
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'flex-end',
          alignItems: 'center',
          gap: 2,
          p: 2,
        }}
      >
        <Skeleton variant="text" width={100} />
        <Box sx={{ display: 'flex', gap: 1 }}>
          <Skeleton variant="circular" width={32} height={32} />
          <Skeleton variant="circular" width={32} height={32} />
        </Box>
      </Box>
    </Card>
  </Box>
);

// ============ 空状态组件 ============

const EmptyListState = () => {
  const translate = useTranslate();
  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        py: 8,
        color: 'text.secondary',
      }}
    >
      <TrafficIcon sx={{ fontSize: 64, opacity: 0.3, mb: 2 }} />
      <Typography variant="h6" sx={{ opacity: 0.6, mb: 1 }}>
        {translate('resources.radius/accounting.empty.title', { _: '暂无计费记录' })}
      </Typography>
      <Typography variant="body2" sx={{ opacity: 0.5 }}>
        {translate('resources.radius/accounting.empty.description', { _: '尝试调整筛选条件或等待新的会话记录' })}
      </Typography>
    </Box>
  );
};
// ============ 搜索表头区块组件 ============

const SearchHeaderCard = () => {
  const translate = useTranslate();
  const { filterValues, setFilters, displayedFilters } = useListContext();
  const [localFilters, setLocalFilters] = useState<Record<string, string>>({});

  // 同步外部筛选值到本地状态
  useEffect(() => {
    const newLocalFilters: Record<string, string> = {};
    if (filterValues) {
      Object.entries(filterValues).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          newLocalFilters[key] = String(value);
        }
      });
    }
    setLocalFilters(newLocalFilters);
  }, [filterValues]);

  const handleFilterChange = useCallback(
    (field: string, value: string) => {
      setLocalFilters(prev => ({ ...prev, [field]: value }));
    },
    [],
  );

  const handleSearch = useCallback(() => {
    const newFilters: Record<string, string> = {};
    Object.entries(localFilters).forEach(([key, value]) => {
      if (value.trim()) {
        newFilters[key] = value.trim();
      }
    });
    setFilters(newFilters, displayedFilters);
  }, [localFilters, setFilters, displayedFilters]);

  const handleClear = useCallback(() => {
    setLocalFilters({});
    setFilters({}, displayedFilters);
  }, [setFilters, displayedFilters]);

  const handleKeyPress = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') {
        handleSearch();
      }
    },
    [handleSearch],
  );

  const filterFields = [
    { key: 'username', label: translate('resources.radius/accounting.fields.username', { _: '用户名' }) },
    { key: 'acct_session_id', label: translate('resources.radius/accounting.fields.acct_session_id', { _: '会话ID' }) },
    { key: 'framed_ipaddr', label: translate('resources.radius/accounting.fields.framed_ipaddr', { _: '用户IP' }) },
    { key: 'framed_ipv6addr', label: translate('resources.radius/accounting.fields.framed_ipv6addr', { _: 'IPv6地址' }) },
    { key: 'nas_addr', label: translate('resources.radius/accounting.fields.nas_addr', { _: 'NAS地址' }) },
    { key: 'mac_addr', label: translate('resources.radius/accounting.fields.mac_addr', { _: 'MAC地址' }) },
  ];

  // 只保留开始时间范围筛选
  const dateFields = [
    { key: 'acct_start_time_gte', label: translate('resources.radius/accounting.filter.start_time_from', { _: '开始时间从' }) },
    { key: 'acct_start_time_lte', label: translate('resources.radius/accounting.filter.start_time_to', { _: '开始时间至' }) },
  ];

  return (
    <Card
      elevation={0}
      sx={{
        mb: 2,
        borderRadius: 2,
        border: theme => `1px solid ${theme.palette.divider}`,
        overflow: 'hidden',
      }}
    >
      <Box
        sx={{
          px: 2.5,
          py: 1.5,
          bgcolor: theme =>
            theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.03)' : 'rgba(0,0,0,0.02)',
          borderBottom: theme => `1px solid ${theme.palette.divider}`,
          display: 'flex',
          alignItems: 'center',
          gap: 1.5,
        }}
      >
        <FilterIcon sx={{ color: 'primary.main', fontSize: 20 }} />
        <Typography variant="subtitle2" sx={{ fontWeight: 600, color: 'text.primary' }}>
          {translate('resources.radius/accounting.filter.title', { _: '筛选条件' })}
        </Typography>
      </Box>

      <CardContent sx={{ p: 2 }}>
        {/* 使用网格布局确保响应式 */}
        <Box
          sx={{
            display: 'grid',
            gap: 1.5,
            gridTemplateColumns: {
              xs: 'repeat(2, 1fr)',           // 手机：2列
              sm: 'repeat(3, 1fr)',           // 平板：3列
              md: 'repeat(4, 1fr)',           // 中屏：4列
              lg: 'repeat(5, 1fr)',           // 大屏：5列
              xl: 'repeat(8, 1fr) auto',      // 超大屏：8列 + 按钮
            },
            alignItems: 'end',
          }}
        >
          {/* 文本筛选字段 */}
          {filterFields.map(field => (
            <MuiTextField
              key={field.key}
              label={field.label}
              value={localFilters[field.key] || ''}
              onChange={e => handleFilterChange(field.key, e.target.value)}
              onKeyPress={handleKeyPress}
              size="small"
              fullWidth
              sx={{
                '& .MuiInputBase-root': {
                  borderRadius: 1.5,
                },
              }}
            />
          ))}

          {/* 时间范围筛选字段 */}
          {dateFields.map(field => (
            <MuiTextField
              key={field.key}
              label={field.label}
              type="datetime-local"
              value={localFilters[field.key] || ''}
              onChange={e => handleFilterChange(field.key, e.target.value)}
              size="small"
              fullWidth
              InputLabelProps={{
                shrink: true,
              }}
              sx={{
                '& .MuiInputBase-root': {
                  borderRadius: 1.5,
                },
              }}
            />
          ))}

          {/* 操作按钮 */}
          <Box sx={{ display: 'flex', gap: 0.5, justifyContent: 'flex-end' }}>
            <Tooltip title={translate('ra.action.clear_filters', { _: '清除筛选' })}>
              <IconButton
                onClick={handleClear}
                size="small"
                sx={{
                  bgcolor: theme => alpha(theme.palette.grey[500], 0.1),
                  '&:hover': {
                    bgcolor: theme => alpha(theme.palette.grey[500], 0.2),
                  },
                }}
              >
                <ClearIcon />
              </IconButton>
            </Tooltip>
            <Tooltip title={translate('ra.action.search', { _: '搜索' })}>
              <IconButton
                onClick={handleSearch}
                color="primary"
                sx={{
                  bgcolor: theme => alpha(theme.palette.primary.main, 0.1),
                  '&:hover': {
                    bgcolor: theme => alpha(theme.palette.primary.main, 0.2),
                  },
                }}
              >
                <SearchIcon />
              </IconButton>
            </Tooltip>
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

// ============ 状态指示器组件 ============

const StatusIndicator = ({ isOnline }: { isOnline: boolean }) => (
  <Chip
    icon={isOnline ? <OnlineIcon sx={{ fontSize: '0.85rem !important' }} /> : <OfflineIcon sx={{ fontSize: '0.85rem !important' }} />}
    label={isOnline ? '在线' : '离线'}
    size="small"
    color={isOnline ? 'success' : 'default'}
    variant={isOnline ? 'filled' : 'outlined'}
    sx={{ height: 22, fontWeight: 500, fontSize: '0.75rem' }}
  />
);

// ============ 增强版 Datagrid 字段组件 ============

const UsernameField = () => {
  const record = useRecordContext<AccountingRecord>();
  if (!record) return null;

  const isOnline = !record.acct_stop_time;

  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
      <Avatar
        sx={{
          width: 28,
          height: 28,
          fontSize: '0.75rem',
          fontWeight: 600,
          bgcolor: isOnline ? 'success.main' : 'grey.400',
        }}
      >
        {record.username?.charAt(0).toUpperCase() || 'U'}
      </Avatar>
      <Box>
        <Typography
          variant="body2"
          sx={{ fontWeight: 600, color: 'text.primary', lineHeight: 1.3 }}
        >
          {record.username || '-'}
        </Typography>
        <StatusIndicator isOnline={isOnline} />
      </Box>
    </Box>
  );
};

const IpAddressField = () => {
  const record = useRecordContext<AccountingRecord>();
  if (!record?.framed_ipaddr) return <Typography variant="body2" color="text.secondary">-</Typography>;

  return (
    <Chip
      label={record.framed_ipaddr}
      size="small"
      color="info"
      variant="outlined"
      sx={{ fontFamily: 'monospace', fontSize: '0.8rem', height: 24 }}
    />
  );
};

const TrafficFieldCompact = ({ type }: { type: 'upload' | 'download' }) => {
  const record = useRecordContext<AccountingRecord>();
  if (!record) return null;

  const value = type === 'upload' ? record.acct_input_total : record.acct_output_total;
  const color = type === 'upload' ? 'info.main' : 'warning.main';
  const Icon = type === 'upload' ? UploadIcon : DownloadIcon;

  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
      <Icon sx={{ fontSize: 16, color }} />
      <Typography variant="body2" sx={{ fontWeight: 500, color }}>
        {formatBytes(value)}
      </Typography>
    </Box>
  );
};

const SessionIdField = () => {
  const record = useRecordContext<AccountingRecord>();
  if (!record?.acct_session_id) return <Typography variant="body2" color="text.secondary">-</Typography>;

  return (
    <Tooltip title={record.acct_session_id}>
      <Typography
        variant="body2"
        sx={{
          fontFamily: 'monospace',
          fontSize: '0.8rem',
          maxWidth: 120,
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
        }}
      >
        {record.acct_session_id}
      </Typography>
    </Tooltip>
  );
};

// ============ 列表操作栏组件 ============

const AccountingListActions = () => {
  const translate = useTranslate();
  return (
    <TopToolbar>
      <SortButton
        fields={['acct_start_time', 'acct_stop_time', 'acct_session_time', 'username']}
        label={translate('ra.action.sort', { _: '排序' })}
      />
      <ExportButton />
    </TopToolbar>
  );
};

// ============ 内部列表内容组件 ============

const AccountingListContent = () => {
  const translate = useTranslate();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { data, isLoading, total } = useListContext<AccountingRecord>();

  // 活动筛选器标签配置
  const fieldLabels = useMemo(
    () => ({
      username: translate('resources.radius/accounting.fields.username'),
      acct_session_id: translate('resources.radius/accounting.fields.acct_session_id'),
      framed_ipaddr: translate('resources.radius/accounting.fields.framed_ipaddr'),
      nas_addr: translate('resources.radius/accounting.fields.nas_addr'),
      mac_addr: translate('resources.radius/accounting.fields.mac_addr'),
      acct_start_time_gte: translate('resources.radius/accounting.fields.acct_start_time_gte'),
      acct_start_time_lte: translate('resources.radius/accounting.fields.acct_start_time_lte'),
    }),
    [translate],
  );

  if (isLoading) {
    return <AccountingListSkeleton />;
  }

  if (!data || data.length === 0) {
    return (
      <Box>
        <SearchHeaderCard />
        <Card
          elevation={0}
          sx={{
            borderRadius: 2,
            border: theme => `1px solid ${theme.palette.divider}`,
          }}
        >
          <EmptyListState />
        </Card>
      </Box>
    );
  }

  return (
    <Box>
      {/* 搜索区块 */}
      <SearchHeaderCard />

      {/* 活动筛选标签 */}
      <ActiveFilters fieldLabels={fieldLabels} />

      {/* 表格容器 */}
      <Card
        elevation={0}
        sx={{
          borderRadius: 2,
          border: theme => `1px solid ${theme.palette.divider}`,
          overflow: 'hidden',
        }}
      >
        {/* 表格统计信息 */}
        <Box
          sx={{
            px: 2,
            py: 1,
            bgcolor: theme =>
              theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.02)' : 'rgba(0,0,0,0.01)',
            borderBottom: theme => `1px solid ${theme.palette.divider}`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <Typography variant="body2" color="text.secondary">
            共 <strong>{total?.toLocaleString() || 0}</strong> 条记录
          </Typography>
        </Box>

        {/* 响应式表格 */}
        <Box
          sx={{
            overflowX: 'auto',
            '& .RaDatagrid-root': {
              minWidth: isMobile ? 1000 : 'auto',
            },
            '& .RaDatagrid-thead': {
              position: 'sticky',
              top: 0,
              zIndex: 1,
              bgcolor: theme =>
                theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.02)',
              '& th': {
                fontWeight: 600,
                fontSize: '0.8rem',
                color: 'text.secondary',
                textTransform: 'uppercase',
                letterSpacing: '0.5px',
                py: 1.5,
                borderBottom: theme => `2px solid ${theme.palette.divider}`,
              },
            },
            '& .RaDatagrid-tbody': {
              '& tr': {
                transition: 'background-color 0.15s ease',
                cursor: 'pointer',
                '&:hover': {
                  bgcolor: theme =>
                    theme.palette.mode === 'dark'
                      ? 'rgba(255,255,255,0.05)'
                      : 'rgba(25, 118, 210, 0.04)',
                  '& .row-actions': {
                    opacity: 1,
                  },
                },
                '&:nth-of-type(odd)': {
                  bgcolor: theme =>
                    theme.palette.mode === 'dark'
                      ? 'rgba(255,255,255,0.01)'
                      : 'rgba(0,0,0,0.01)',
                },
              },
              '& td': {
                py: 1.5,
                fontSize: '0.875rem',
                borderBottom: theme => `1px solid ${alpha(theme.palette.divider, 0.5)}`,
              },
            },
          }}
        >
          <Datagrid
            rowClick="show"
            bulkActionButtons={false}
          >
            <FunctionField
              source="username"
              label={translate('resources.radius/accounting.fields.username')}
              render={() => <UsernameField />}
            />
            <FunctionField
              source="acct_session_id"
              label={translate('resources.radius/accounting.fields.acct_session_id')}
              render={() => <SessionIdField />}
            />
            <FunctionField
              source="framed_ipaddr"
              label={translate('resources.radius/accounting.fields.framed_ipaddr')}
              render={() => <IpAddressField />}
            />
            <TextField
              source="nas_addr"
              label={translate('resources.radius/accounting.fields.nas_addr')}
            />
            <NasPortField />
            <DateField
              source="acct_start_time"
              label={translate('resources.radius/accounting.fields.acct_start_time')}
              showTime
            />
            <DateField
              source="acct_stop_time"
              label={translate('resources.radius/accounting.fields.acct_stop_time')}
              showTime
              emptyText="-"
            />
            <SessionDurationField />
            <FunctionField
              source="acct_input_total"
              label={translate('resources.radius/accounting.fields.acct_input_total')}
              render={() => <TrafficFieldCompact type="upload" />}
            />
            <FunctionField
              source="acct_output_total"
              label={translate('resources.radius/accounting.fields.acct_output_total')}
              render={() => <TrafficFieldCompact type="download" />}
            />
          </Datagrid>
        </Box>
      </Card>
    </Box>
  );
};

// ============ 主列表组件 ============

export const AccountingList = () => {
  return (
    <List
      actions={<AccountingListActions />}
      sort={{ field: 'acct_start_time', order: 'DESC' }}
      perPage={LARGE_LIST_PER_PAGE}
      pagination={<ServerPagination />}
      empty={false}
    >
      <AccountingListContent />
    </List>
  );
};

export const AccountingShow = () => (
  <Show>
    <AccountingDetails />
  </Show>
);
