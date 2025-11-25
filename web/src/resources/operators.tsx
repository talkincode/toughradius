import {
  List,
  Datagrid,
  TextField,
  EmailField,
  DateField,
  Edit,
  SimpleForm,
  TextInput,
  SelectInput,
  PasswordInput,
  Create,
  Show,
  TopToolbar,
  CreateButton,
  ExportButton,
  ListButton,
  SortButton,
  required,
  minLength,
  maxLength,
  email,
  regex,
  useRecordContext,
  useGetIdentity,
  useTranslate,
  useRefresh,
  useNotify,
  useListContext,
  RaRecord,
  FunctionField
} from 'react-admin';
import {
  Box,
  Chip,
  Typography,
  Card,
  CardContent,
  Stack,
  Avatar,
  IconButton,
  Tooltip,
  Skeleton,
  useTheme,
  useMediaQuery,
  TextField as MuiTextField,
  alpha
} from '@mui/material';
import { ReactNode, useMemo, useCallback, useState, useEffect } from 'react';
import {
  Person as PersonIcon,
  Security as SecurityIcon,
  Schedule as TimeIcon,
  Note as NoteIcon,
  ContentCopy as CopyIcon,
  Refresh as RefreshIcon,
  ArrowBack as BackIcon,
  Print as PrintIcon,
  FilterList as FilterIcon,
  Search as SearchIcon,
  Clear as ClearIcon,
  CheckCircle as EnabledIcon,
  Cancel as DisabledIcon,
  Email as EmailIcon,
  Phone as PhoneIcon,
  AdminPanelSettings as AdminIcon
} from '@mui/icons-material';
import {
  ServerPagination,
  ActiveFilters,
  FormSection,
  FieldGrid,
  FieldGridItem,
  formLayoutSx,
  DetailItem,
  DetailSectionCard,
  EmptyValue
} from '../components';

const LARGE_LIST_PER_PAGE = 50;

// ============ 类型定义 ============

interface Operator extends RaRecord {
  username?: string;
  realname?: string;
  email?: string;
  mobile?: string;
  level?: 'super' | 'admin' | 'operator';
  status?: 'enabled' | 'disabled';
  remark?: string;
  last_login?: string;
  created_at?: string;
  updated_at?: string;
}

// ============ 工具函数 ============

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

// ============ 验证规则 ============

const useValidationRules = () => {
  const translate = useTranslate();

  return {
    validateUsername: [
      required(translate('resources.system/operators.validation.username_required', { _: '用户名必填' })),
      minLength(3, translate('resources.system/operators.validation.username_min', { _: '用户名至少3个字符' })),
      maxLength(30, translate('resources.system/operators.validation.username_max', { _: '用户名最多30个字符' })),
      regex(/^[a-zA-Z0-9_]+$/, translate('resources.system/operators.validation.username_format', { _: '用户名只能包含字母、数字和下划线' })),
    ],
    validatePassword: [
      required(translate('resources.system/operators.validation.password_required', { _: '密码必填' })),
      minLength(6, translate('resources.system/operators.validation.password_min', { _: '密码至少6个字符' })),
      maxLength(50, translate('resources.system/operators.validation.password_max', { _: '密码最多50个字符' })),
      regex(/^(?=.*[A-Za-z])(?=.*\d).+$/, translate('resources.system/operators.validation.password_format', { _: '密码必须包含字母和数字' })),
    ],
    validatePasswordOptional: [
      minLength(6, translate('resources.system/operators.validation.password_min', { _: '密码至少6个字符' })),
      maxLength(50, translate('resources.system/operators.validation.password_max', { _: '密码最多50个字符' })),
      regex(/^(?=.*[A-Za-z])(?=.*\d).+$/, translate('resources.system/operators.validation.password_format', { _: '密码必须包含字母和数字' })),
    ],
    validateEmail: [email(translate('resources.system/operators.validation.email_invalid', { _: '邮箱格式不正确' }))],
    validateMobile: [
      regex(
        /^(0|\+?86)?(13[0-9]|14[57]|15[0-35-9]|17[0678]|18[0-9])[0-9]{8}$/,
        translate('resources.system/operators.validation.mobile_invalid', { _: '手机号格式不正确' })
      ),
    ],
    validateRealname: [required(translate('resources.system/operators.validation.realname_required', { _: '真实姓名必填' }))],
    validateLevel: [required(translate('resources.system/operators.validation.level_required', { _: '权限级别必填' }))],
    validateStatus: [required(translate('resources.system/operators.validation.status_required', { _: '状态必填' }))],
  };
};

// ============ 列表加载骨架屏 ============

const OperatorListSkeleton = ({ rows = 10 }: { rows?: number }) => (
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
              md: 'repeat(4, 1fr)',
            },
          }}
        >
          {[...Array(4)].map((_, i) => (
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
          gridTemplateColumns: 'repeat(8, 1fr)',
          gap: 1,
          p: 2,
          bgcolor: theme =>
            theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.02)',
          borderBottom: theme => `1px solid ${theme.palette.divider}`,
        }}
      >
        {[...Array(8)].map((_, i) => (
          <Skeleton key={i} variant="text" height={20} width="80%" />
        ))}
      </Box>

      {/* 表格行 */}
      {[...Array(rows)].map((_, rowIndex) => (
        <Box
          key={rowIndex}
          sx={{
            display: 'grid',
            gridTemplateColumns: 'repeat(8, 1fr)',
            gap: 1,
            p: 2,
            borderBottom: theme => `1px solid ${theme.palette.divider}`,
          }}
        >
          {[...Array(8)].map((_, colIndex) => (
            <Skeleton
              key={colIndex}
              variant="text"
              height={18}
              width={`${60 + Math.random() * 30}%`}
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

const OperatorEmptyState = () => {
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
      <AdminIcon sx={{ fontSize: 64, opacity: 0.3, mb: 2 }} />
      <Typography variant="h6" sx={{ opacity: 0.6, mb: 1 }}>
        {translate('resources.system/operators.empty.title', { _: '暂无操作员' })}
      </Typography>
      <Typography variant="body2" sx={{ opacity: 0.5 }}>
        {translate('resources.system/operators.empty.description', { _: '点击"新建"按钮添加第一个操作员' })}
      </Typography>
    </Box>
  );
};

// ============ 搜索表头区块组件 ============

const OperatorSearchHeaderCard = () => {
  const translate = useTranslate();
  const { filterValues, setFilters, displayedFilters } = useListContext();
  const [localFilters, setLocalFilters] = useState<Record<string, string>>({});

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
    { key: 'username', label: translate('resources.system/operators.fields.username', { _: '用户名' }) },
    { key: 'realname', label: translate('resources.system/operators.fields.realname', { _: '真实姓名' }) },
    { key: 'email', label: translate('resources.system/operators.fields.email', { _: '邮箱' }) },
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
          {translate('resources.system/operators.filter.title', { _: '筛选条件' })}
        </Typography>
      </Box>

      <CardContent sx={{ p: 2 }}>
        <Box
          sx={{
            display: 'grid',
            gap: 1.5,
            gridTemplateColumns: {
              xs: 'repeat(1, 1fr)',
              sm: 'repeat(2, 1fr)',
              md: 'repeat(4, 1fr)',
            },
            alignItems: 'end',
          }}
        >
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

// ============ 状态和级别组件 ============

const StatusIndicator = ({ isEnabled }: { isEnabled: boolean }) => {
  const translate = useTranslate();
  return (
    <Chip
      icon={isEnabled ? <EnabledIcon sx={{ fontSize: '0.85rem !important' }} /> : <DisabledIcon sx={{ fontSize: '0.85rem !important' }} />}
      label={isEnabled ? translate('resources.system/operators.status.enabled', { _: '启用' }) : translate('resources.system/operators.status.disabled', { _: '禁用' })}
      size="small"
      color={isEnabled ? 'success' : 'default'}
      variant={isEnabled ? 'filled' : 'outlined'}
      sx={{ height: 22, fontWeight: 500, fontSize: '0.75rem' }}
    />
  );
};

const LevelChip = ({ level }: { level?: string }) => {
  const translate = useTranslate();
  
  const levelConfig: Record<string, { color: 'error' | 'warning' | 'info'; label: string }> = {
    super: { color: 'error', label: translate('resources.system/operators.levels.super', { _: '超级管理员' }) },
    admin: { color: 'warning', label: translate('resources.system/operators.levels.admin', { _: '管理员' }) },
    operator: { color: 'info', label: translate('resources.system/operators.levels.operator', { _: '操作员' }) },
  };

  const config = levelConfig[level || ''] || { color: 'info', label: level || '-' };

  return (
    <Chip
      label={config.label}
      size="small"
      color={config.color}
      sx={{ height: 22, fontWeight: 500, fontSize: '0.75rem' }}
    />
  );
};

// ============ 增强版字段组件 ============

const OperatorNameField = () => {
  const record = useRecordContext<Operator>();
  if (!record) return null;

  const isEnabled = record.status === 'enabled';

  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
      <Avatar
        sx={{
          width: 32,
          height: 32,
          fontSize: '0.85rem',
          fontWeight: 600,
          bgcolor: isEnabled ? 'primary.main' : 'grey.400',
        }}
      >
        {record.username?.charAt(0).toUpperCase() || 'O'}
      </Avatar>
      <Box>
        <Typography
          variant="body2"
          sx={{ fontWeight: 600, color: 'text.primary', lineHeight: 1.3 }}
        >
          {record.username || '-'}
        </Typography>
        <StatusIndicator isEnabled={isEnabled} />
      </Box>
    </Box>
  );
};

const StatusField = () => {
  const record = useRecordContext<Operator>();
  if (!record) return null;
  return <StatusIndicator isEnabled={record.status === 'enabled'} />;
};

const LevelField = () => {
  const record = useRecordContext<Operator>();
  if (!record) return null;
  return <LevelChip level={record.level} />;
};

// ============ 列表操作栏组件 ============

const OperatorListActions = () => {
  const translate = useTranslate();
  return (
    <TopToolbar>
      <SortButton
        fields={['created_at', 'username', 'last_login']}
        label={translate('ra.action.sort', { _: '排序' })}
      />
      <CreateButton />
      <ExportButton />
    </TopToolbar>
  );
};

// ============ 内部列表内容组件 ============

const OperatorListContent = () => {
  const translate = useTranslate();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { data, isLoading, total } = useListContext<Operator>();

  const fieldLabels = useMemo(
    () => ({
      username: translate('resources.system/operators.fields.username', { _: '用户名' }),
      realname: translate('resources.system/operators.fields.realname', { _: '真实姓名' }),
      email: translate('resources.system/operators.fields.email', { _: '邮箱' }),
      status: translate('resources.system/operators.fields.status', { _: '状态' }),
      level: translate('resources.system/operators.fields.level', { _: '权限级别' }),
    }),
    [translate],
  );

  const statusLabels = useMemo(
    () => ({
      enabled: translate('resources.system/operators.status.enabled', { _: '启用' }),
      disabled: translate('resources.system/operators.status.disabled', { _: '禁用' }),
    }),
    [translate],
  );

  const levelLabels = useMemo(
    () => ({
      super: translate('resources.system/operators.levels.super', { _: '超级管理员' }),
      admin: translate('resources.system/operators.levels.admin', { _: '管理员' }),
      operator: translate('resources.system/operators.levels.operator', { _: '操作员' }),
    }),
    [translate],
  );

  if (isLoading) {
    return <OperatorListSkeleton />;
  }

  if (!data || data.length === 0) {
    return (
      <Box>
        <OperatorSearchHeaderCard />
        <Card
          elevation={0}
          sx={{
            borderRadius: 2,
            border: theme => `1px solid ${theme.palette.divider}`,
          }}
        >
          <OperatorEmptyState />
        </Card>
      </Box>
    );
  }

  return (
    <Box>
      {/* 搜索区块 */}
      <OperatorSearchHeaderCard />

      {/* 活动筛选标签 */}
      <ActiveFilters fieldLabels={fieldLabels} valueLabels={{ status: statusLabels, level: levelLabels }} />

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
            共 <strong>{total?.toLocaleString() || 0}</strong> 个操作员
          </Typography>
        </Box>

        {/* 响应式表格 */}
        <Box
          sx={{
            overflowX: 'auto',
            '& .RaDatagrid-root': {
              minWidth: isMobile ? 900 : 'auto',
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
          <Datagrid rowClick="show" bulkActionButtons={false}>
            <FunctionField
              source="username"
              label={translate('resources.system/operators.fields.username', { _: '用户名' })}
              render={() => <OperatorNameField />}
            />
            <TextField
              source="realname"
              label={translate('resources.system/operators.fields.realname', { _: '真实姓名' })}
            />
            <EmailField
              source="email"
              label={translate('resources.system/operators.fields.email', { _: '邮箱' })}
            />
            <TextField
              source="mobile"
              label={translate('resources.system/operators.fields.mobile', { _: '手机号' })}
            />
            <FunctionField
              source="level"
              label={translate('resources.system/operators.fields.level', { _: '权限级别' })}
              render={() => <LevelField />}
            />
            <DateField
              source="last_login"
              label={translate('resources.system/operators.fields.last_login', { _: '最后登录' })}
              showTime
            />
            <DateField
              source="created_at"
              label={translate('resources.system/operators.fields.created_at', { _: '创建时间' })}
              showTime
            />
          </Datagrid>
        </Box>
      </Card>
    </Box>
  );
};

// 操作员列表
export const OperatorList = () => {
  return (
    <List
      actions={<OperatorListActions />}
      sort={{ field: 'created_at', order: 'DESC' }}
      perPage={LARGE_LIST_PER_PAGE}
      pagination={<ServerPagination />}
      empty={false}
    >
      <OperatorListContent />
    </List>
  );
};

// ============ 密码输入框组件 ============

const PasswordInputWithRecord = () => {
  const record = useRecordContext<Operator>();
  const translate = useTranslate();
  const validation = useValidationRules();
  
  if (record?.level === 'super') {
    return null;
  }
  
  return (
    <PasswordInput 
      source="password" 
      label={translate('resources.system/operators.fields.password', { _: '密码' })} 
      validate={validation.validatePasswordOptional}
      helperText={translate('resources.system/operators.helpers.password_optional', { _: '留空则不修改密码' })} 
      fullWidth
      size="small"
    />
  );
};

// ============ 编辑页面 ============

export const OperatorEdit = () => {
  const { identity } = useGetIdentity();
  const record = useRecordContext<Operator>();
  const translate = useTranslate();
  const validation = useValidationRules();
  
  const isEditingSelf = identity && record && String(identity.id) === String(record.id);
  const canManagePermissions = identity?.level === 'super' || identity?.level === 'admin';
  
  return (
    <Edit>
      <SimpleForm sx={formLayoutSx}>
        <FormSection 
          title={translate('resources.system/operators.sections.basic.title', { _: '账号信息' })} 
          description={translate('resources.system/operators.sections.basic.description', { _: '操作员的登录账号和密码' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput 
                source="id" 
                label={translate('resources.system/operators.fields.id', { _: '操作员ID' })} 
                disabled 
                fullWidth 
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput 
                source="username" 
                label={translate('resources.system/operators.fields.username', { _: '用户名' })} 
                validate={validation.validateUsername}
                helperText={translate('resources.system/operators.helpers.username', { _: '3-30个字符，只能包含字母、数字和下划线' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <PasswordInputWithRecord />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection 
          title={translate('resources.system/operators.sections.personal.title', { _: '个人信息' })} 
          description={translate('resources.system/operators.sections.personal.description', { _: '联系方式和个人资料' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput 
                source="realname" 
                label={translate('resources.system/operators.fields.realname', { _: '真实姓名' })} 
                validate={validation.validateRealname}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput 
                source="email" 
                label={translate('resources.system/operators.fields.email', { _: '邮箱' })} 
                type="email" 
                validate={validation.validateEmail}
                helperText={translate('resources.system/operators.helpers.email', { _: '用于接收系统通知' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput 
                source="mobile" 
                label={translate('resources.system/operators.fields.mobile', { _: '手机号' })} 
                validate={validation.validateMobile}
                helperText={translate('resources.system/operators.helpers.mobile', { _: '中国大陆手机号' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        {canManagePermissions && (
          <FormSection 
            title={translate('resources.system/operators.sections.permissions.title', { _: '权限设置' })} 
            description={translate('resources.system/operators.sections.permissions.description', { _: '账号权限和状态配置' })}
          >
            <FieldGrid columns={{ xs: 1, sm: 2 }}>
              <FieldGridItem>
                <SelectInput
                  source="level"
                  label={translate('resources.system/operators.fields.level', { _: '权限级别' })}
                  validate={validation.validateLevel}
                  disabled={isEditingSelf}
                  choices={[
                    { id: 'super', name: translate('resources.system/operators.levels.super', { _: '超级管理员' }) },
                    { id: 'admin', name: translate('resources.system/operators.levels.admin', { _: '管理员' }) },
                    { id: 'operator', name: translate('resources.system/operators.levels.operator', { _: '操作员' }) },
                  ]}
                  helperText={isEditingSelf ? translate('resources.system/operators.helpers.cannot_change_own_level', { _: '不能修改自己的权限级别' }) : translate('resources.system/operators.helpers.level', { _: '选择操作员的权限级别' })}
                  fullWidth
                  size="small"
                />
              </FieldGridItem>
              <FieldGridItem>
                <SelectInput
                  source="status"
                  label={translate('resources.system/operators.fields.status', { _: '状态' })}
                  validate={validation.validateStatus}
                  disabled={isEditingSelf}
                  choices={[
                    { id: 'enabled', name: translate('resources.system/operators.status.enabled', { _: '启用' }) },
                    { id: 'disabled', name: translate('resources.system/operators.status.disabled', { _: '禁用' }) },
                  ]}
                  helperText={isEditingSelf ? translate('resources.system/operators.helpers.cannot_change_own_status', { _: '不能修改自己的状态' }) : translate('resources.system/operators.helpers.status', { _: '禁用后无法登录系统' })}
                  fullWidth
                  size="small"
                />
              </FieldGridItem>
            </FieldGrid>
          </FormSection>
        )}

        <FormSection 
          title={translate('resources.system/operators.sections.remark.title', { _: '备注信息' })}
        >
          <FieldGrid columns={{ xs: 1 }}>
            <FieldGridItem>
              <TextInput 
                source="remark" 
                label={translate('resources.system/operators.fields.remark', { _: '备注' })} 
                multiline 
                minRows={3} 
                fullWidth
                size="small"
                helperText={translate('resources.system/operators.helpers.remark', { _: '可选的备注信息' })}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Edit>
  );
};

// ============ 创建页面 ============

export const OperatorCreate = () => {
  const translate = useTranslate();
  const validation = useValidationRules();
  
  return (
    <Create>
      <SimpleForm sx={formLayoutSx}>
        <FormSection 
          title={translate('resources.system/operators.sections.basic.title', { _: '账号信息' })} 
          description={translate('resources.system/operators.sections.basic.description', { _: '操作员的登录账号和密码' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput 
                source="username" 
                label={translate('resources.system/operators.fields.username', { _: '用户名' })} 
                validate={validation.validateUsername}
                helperText={translate('resources.system/operators.helpers.username', { _: '3-30个字符，只能包含字母、数字和下划线' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <PasswordInput 
                source="password" 
                label={translate('resources.system/operators.fields.password', { _: '密码' })} 
                validate={validation.validatePassword}
                helperText={translate('resources.system/operators.helpers.password', { _: '6-50个字符，必须包含字母和数字' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection 
          title={translate('resources.system/operators.sections.personal.title', { _: '个人信息' })} 
          description={translate('resources.system/operators.sections.personal.description', { _: '联系方式和个人资料' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput 
                source="realname" 
                label={translate('resources.system/operators.fields.realname', { _: '真实姓名' })} 
                validate={validation.validateRealname}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput 
                source="email" 
                label={translate('resources.system/operators.fields.email', { _: '邮箱' })} 
                type="email" 
                validate={validation.validateEmail}
                helperText={translate('resources.system/operators.helpers.email', { _: '用于接收系统通知' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput 
                source="mobile" 
                label={translate('resources.system/operators.fields.mobile', { _: '手机号' })} 
                validate={validation.validateMobile}
                helperText={translate('resources.system/operators.helpers.mobile', { _: '中国大陆手机号' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection 
          title={translate('resources.system/operators.sections.permissions.title', { _: '权限设置' })} 
          description={translate('resources.system/operators.sections.permissions.description', { _: '账号权限和状态配置' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <SelectInput
                source="level"
                label={translate('resources.system/operators.fields.level', { _: '权限级别' })}
                validate={validation.validateLevel}
                defaultValue="operator"
                choices={[
                  { id: 'super', name: translate('resources.system/operators.levels.super', { _: '超级管理员' }) },
                  { id: 'admin', name: translate('resources.system/operators.levels.admin', { _: '管理员' }) },
                  { id: 'operator', name: translate('resources.system/operators.levels.operator', { _: '操作员' }) },
                ]}
                helperText={translate('resources.system/operators.helpers.level', { _: '选择操作员的权限级别' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput
                source="status"
                label={translate('resources.system/operators.fields.status', { _: '状态' })}
                validate={validation.validateStatus}
                defaultValue="enabled"
                choices={[
                  { id: 'enabled', name: translate('resources.system/operators.status.enabled', { _: '启用' }) },
                  { id: 'disabled', name: translate('resources.system/operators.status.disabled', { _: '禁用' }) },
                ]}
                helperText={translate('resources.system/operators.helpers.status', { _: '禁用后无法登录系统' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection 
          title={translate('resources.system/operators.sections.remark.title', { _: '备注信息' })}
        >
          <FieldGrid columns={{ xs: 1 }}>
            <FieldGridItem>
              <TextInput 
                source="remark" 
                label={translate('resources.system/operators.fields.remark', { _: '备注' })} 
                multiline 
                minRows={3} 
                fullWidth
                size="small"
                helperText={translate('resources.system/operators.helpers.remark', { _: '可选的备注信息' })}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Create>
  );
};

// ============ 详情页顶部概览卡片 ============

const OperatorHeaderCard = () => {
  const record = useRecordContext<Operator>();
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

  const isEnabled = record.status === 'enabled';

  return (
    <Card
      elevation={0}
      sx={{
        borderRadius: 4,
        background: theme =>
          theme.palette.mode === 'dark'
            ? isEnabled
              ? `linear-gradient(135deg, ${alpha(theme.palette.primary.dark, 0.4)} 0%, ${alpha(theme.palette.info.dark, 0.3)} 100%)`
              : `linear-gradient(135deg, ${alpha(theme.palette.grey[800], 0.5)} 0%, ${alpha(theme.palette.grey[700], 0.3)} 100%)`
            : isEnabled
            ? `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.1)} 0%, ${alpha(theme.palette.info.main, 0.08)} 100%)`
            : `linear-gradient(135deg, ${alpha(theme.palette.grey[400], 0.15)} 0%, ${alpha(theme.palette.grey[300], 0.1)} 100%)`,
        border: theme => `1px solid ${alpha(isEnabled ? theme.palette.primary.main : theme.palette.grey[500], 0.2)}`,
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
          background: theme => alpha(isEnabled ? theme.palette.primary.main : theme.palette.grey[500], 0.1),
          pointerEvents: 'none',
        }}
      />

      <CardContent sx={{ p: 3, position: 'relative', zIndex: 1 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 3 }}>
          {/* 左侧：操作员信息 */}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <Avatar
              sx={{
                width: 64,
                height: 64,
                bgcolor: isEnabled ? 'primary.main' : 'grey.500',
                fontSize: '1.5rem',
                fontWeight: 700,
                boxShadow: theme => `0 4px 14px ${alpha(isEnabled ? theme.palette.primary.main : theme.palette.grey[500], 0.4)}`,
              }}
            >
              {record.username?.charAt(0).toUpperCase() || 'O'}
            </Avatar>
            <Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
                <Typography variant="h5" sx={{ fontWeight: 700, color: 'text.primary' }}>
                  {record.username || <EmptyValue message="未知用户" />}
                </Typography>
                <StatusIndicator isEnabled={isEnabled} />
                <LevelChip level={record.level} />
              </Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                {record.realname && (
                  <Typography variant="body2" color="text.secondary">
                    {record.realname}
                  </Typography>
                )}
              </Box>
              {record.username && (
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 0.5 }}>
                  <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace' }}>
                    ID: {record.id}
                  </Typography>
                  <Tooltip title="复制用户名">
                    <IconButton
                      size="small"
                      onClick={() => handleCopy(record.username!, '用户名')}
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
              <EmailIcon sx={{ fontSize: '1.1rem', color: 'info.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.system/operators.fields.email', { _: '邮箱' })}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600, wordBreak: 'break-all' }}>
              {record.email || '-'}
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
              <PhoneIcon sx={{ fontSize: '1.1rem', color: 'success.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.system/operators.fields.mobile', { _: '手机号' })}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {record.mobile || '-'}
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
              <SecurityIcon sx={{ fontSize: '1.1rem', color: 'warning.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.system/operators.fields.level', { _: '权限级别' })}
              </Typography>
            </Box>
            <LevelChip level={record.level} />
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
              <TimeIcon sx={{ fontSize: '1.1rem', color: 'primary.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.system/operators.fields.last_login', { _: '最后登录' })}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {formatTimestamp(record.last_login)}
            </Typography>
          </Box>
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

// ============ 操作员详情内容 ============

const OperatorDetails = () => {
  const record = useRecordContext<Operator>();
  const translate = useTranslate();
  
  if (!record) {
    return null;
  }

  return (
    <>
      <style>{printStyles}</style>
      <Box className="printable-content" sx={{ width: '100%', p: { xs: 2, sm: 3, md: 4 } }}>
        <Stack spacing={3}>
          {/* 顶部概览卡片 */}
          <OperatorHeaderCard />

          {/* 基本信息 */}
          <DetailSectionCard
            title={translate('resources.system/operators.sections.basic.title', { _: '账号信息' })}
            description={translate('resources.system/operators.sections.basic.description', { _: '操作员的登录账号信息' })}
            icon={<PersonIcon />}
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
                label={translate('resources.system/operators.fields.id', { _: '操作员ID' })}
                value={record.id}
              />
              <DetailItem
                label={translate('resources.system/operators.fields.username', { _: '用户名' })}
                value={record.username}
                highlight
              />
              <DetailItem
                label={translate('resources.system/operators.fields.realname', { _: '真实姓名' })}
                value={record.realname || <EmptyValue />}
              />
            </Box>
          </DetailSectionCard>

          {/* 联系方式 */}
          <DetailSectionCard
            title={translate('resources.system/operators.sections.personal.title', { _: '联系方式' })}
            description={translate('resources.system/operators.sections.personal.description', { _: '联系信息' })}
            icon={<EmailIcon />}
            color="info"
          >
            <Box
              sx={{
                display: 'grid',
                gap: 2,
                gridTemplateColumns: {
                  xs: 'repeat(1, 1fr)',
                  sm: 'repeat(2, 1fr)',
                },
              }}
            >
              <DetailItem
                label={translate('resources.system/operators.fields.email', { _: '邮箱' })}
                value={record.email || <EmptyValue />}
              />
              <DetailItem
                label={translate('resources.system/operators.fields.mobile', { _: '手机号' })}
                value={record.mobile || <EmptyValue />}
              />
            </Box>
          </DetailSectionCard>

          {/* 权限设置 */}
          <DetailSectionCard
            title={translate('resources.system/operators.sections.permissions.title', { _: '权限设置' })}
            description={translate('resources.system/operators.sections.permissions.description', { _: '账号权限和状态' })}
            icon={<SecurityIcon />}
            color="warning"
          >
            <Box
              sx={{
                display: 'grid',
                gap: 2,
                gridTemplateColumns: {
                  xs: 'repeat(1, 1fr)',
                  sm: 'repeat(2, 1fr)',
                },
              }}
            >
              <DetailItem
                label={translate('resources.system/operators.fields.level', { _: '权限级别' })}
                value={<LevelChip level={record.level} />}
                highlight
              />
              <DetailItem
                label={translate('resources.system/operators.fields.status', { _: '状态' })}
                value={<StatusIndicator isEnabled={record.status === 'enabled'} />}
                highlight
              />
            </Box>
          </DetailSectionCard>

          {/* 时间信息 */}
          <DetailSectionCard
            title={translate('resources.system/operators.sections.other.title', { _: '时间信息' })}
            description={translate('resources.system/operators.sections.other.description', { _: '登录和创建时间' })}
            icon={<TimeIcon />}
            color="info"
          >
            <Box
              sx={{
                display: 'grid',
                gap: 2,
                gridTemplateColumns: {
                  xs: 'repeat(1, 1fr)',
                  sm: 'repeat(3, 1fr)',
                },
              }}
            >
              <DetailItem
                label={translate('resources.system/operators.fields.last_login', { _: '最后登录' })}
                value={formatTimestamp(record.last_login)}
              />
              <DetailItem
                label={translate('resources.system/operators.fields.created_at', { _: '创建时间' })}
                value={formatTimestamp(record.created_at)}
              />
              <DetailItem
                label={translate('resources.system/operators.fields.updated_at', { _: '更新时间' })}
                value={formatTimestamp(record.updated_at)}
              />
            </Box>
          </DetailSectionCard>

          {/* 备注信息 */}
          <DetailSectionCard
            title={translate('resources.system/operators.sections.remark.title', { _: '备注信息' })}
            icon={<NoteIcon />}
            color="primary"
          >
            <Box
              sx={{
                p: 2,
                borderRadius: 2,
                bgcolor: theme =>
                  theme.palette.mode === 'dark'
                    ? 'rgba(255, 255, 255, 0.02)'
                    : 'rgba(0, 0, 0, 0.02)',
                border: theme => `1px solid ${theme.palette.divider}`,
                minHeight: 80,
              }}
            >
              <Typography
                variant="body2"
                sx={{
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                  color: record.remark ? 'text.primary' : 'text.disabled',
                  fontStyle: record.remark ? 'normal' : 'italic',
                }}
              >
                {record.remark || translate('resources.system/operators.empty.no_remark', { _: '无备注信息' })}
              </Typography>
            </Box>
          </DetailSectionCard>
        </Stack>
      </Box>
    </>
  );
};

// 操作员详情
export const OperatorShow = () => {
  return (
    <Show>
      <OperatorDetails />
    </Show>
  );
};
