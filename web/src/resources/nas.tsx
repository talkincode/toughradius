import {
  List,
  Datagrid,
  TextField,
  DateField,
  Edit,
  SimpleForm,
  TextInput,
  NumberInput,
  SelectInput,
  ReferenceInput,
  Create,
  Show,
  TopToolbar,
  CreateButton,
  ExportButton,
  SortButton,
  ReferenceField,
  PasswordInput,
  required,
  minLength,
  maxLength,
  number,
  minValue,
  maxValue,
  useRecordContext,
  Toolbar,
  SaveButton,
  DeleteButton,
  ToolbarProps,
  ListButton,
  useTranslate,
  useListContext,
  useRefresh,
  useNotify,
  RaRecord,
  FunctionField
} from 'react-admin';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Stack,
  Chip,
  Avatar,
  Skeleton,
  IconButton,
  Tooltip,
  useTheme,
  useMediaQuery,
  TextField as MuiTextField,
  alpha
} from '@mui/material';
import { useMemo, useCallback, useState, useEffect } from 'react';
import {
  Router as NasIcon,
  NetworkCheck as NetworkIcon,
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
  Dns as ServerIcon,
  VpnKey as SecretIcon,
  Business as VendorIcon
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

interface NASDevice extends RaRecord {
  name?: string;
  identifier?: string;
  ipaddr?: string;
  hostname?: string;
  secret?: string;
  vendor_code?: string;
  model?: string;
  coa_port?: number;
  status?: 'enabled' | 'disabled';
  node_id?: string;
  tags?: string;
  remark?: string;
  created_at?: string;
  updated_at?: string;
}

// ============ 常量定义 ============

// 厂商代码选项
const VENDOR_CHOICES = [
  { id: '9', name: 'Cisco' },
  { id: '2011', name: 'Huawei' },
  { id: '14988', name: 'Mikrotik' },
  { id: '25506', name: 'H3C' },
  { id: '3902', name: 'ZTE' },
  { id: '10055', name: 'Ikuai' },
  { id: '0', name: 'Standard' },
];

// 状态选项
const STATUS_CHOICES = [
  { id: 'enabled', name: '启用' },
  { id: 'disabled', name: '禁用' },
];

// 获取厂商名称
const getVendorName = (code?: string): string => {
  if (!code) return '-';
  const vendor = VENDOR_CHOICES.find(v => v.id === String(code));
  return vendor ? vendor.name : code;
};

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

// ============ 列表加载骨架屏 ============

const NASListSkeleton = ({ rows = 10 }: { rows?: number }) => (
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
          gridTemplateColumns: 'repeat(7, 1fr)',
          gap: 1,
          p: 2,
          bgcolor: theme =>
            theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.02)',
          borderBottom: theme => `1px solid ${theme.palette.divider}`,
        }}
      >
        {[...Array(7)].map((_, i) => (
          <Skeleton key={i} variant="text" height={20} width="80%" />
        ))}
      </Box>

      {/* 表格行 */}
      {[...Array(rows)].map((_, rowIndex) => (
        <Box
          key={rowIndex}
          sx={{
            display: 'grid',
            gridTemplateColumns: 'repeat(7, 1fr)',
            gap: 1,
            p: 2,
            borderBottom: theme => `1px solid ${theme.palette.divider}`,
          }}
        >
          {[...Array(7)].map((_, colIndex) => (
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

const NASEmptyState = () => {
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
      <NasIcon sx={{ fontSize: 64, opacity: 0.3, mb: 2 }} />
      <Typography variant="h6" sx={{ opacity: 0.6, mb: 1 }}>
        {translate('resources.network/nas.empty.title', { _: '暂无 NAS 设备' })}
      </Typography>
      <Typography variant="body2" sx={{ opacity: 0.5 }}>
        {translate('resources.network/nas.empty.description', { _: '点击"新建"按钮添加第一个 NAS 设备' })}
      </Typography>
    </Box>
  );
};

// ============ 搜索表头区块组件 ============

const NASSearchHeaderCard = () => {
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
    { key: 'name', label: translate('resources.network/nas.fields.name', { _: '设备名称' }) },
    { key: 'ipaddr', label: translate('resources.network/nas.fields.ipaddr', { _: 'IP地址' }) },
    { key: 'identifier', label: translate('resources.network/nas.fields.identifier', { _: '标识符' }) },
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
          {translate('resources.network/nas.filter.title', { _: '筛选条件' })}
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

// ============ 状态组件 ============

const StatusIndicator = ({ isEnabled }: { isEnabled: boolean }) => {
  const translate = useTranslate();
  return (
    <Chip
      icon={isEnabled ? <EnabledIcon sx={{ fontSize: '0.85rem !important' }} /> : <DisabledIcon sx={{ fontSize: '0.85rem !important' }} />}
      label={isEnabled ? translate('resources.network/nas.status.enabled', { _: '启用' }) : translate('resources.network/nas.status.disabled', { _: '禁用' })}
      size="small"
      color={isEnabled ? 'success' : 'default'}
      variant={isEnabled ? 'filled' : 'outlined'}
      sx={{ height: 22, fontWeight: 500, fontSize: '0.75rem' }}
    />
  );
};

// ============ 增强版字段组件 ============

const NASNameField = () => {
  const record = useRecordContext<NASDevice>();
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
        {record.name?.charAt(0).toUpperCase() || 'N'}
      </Avatar>
      <Box>
        <Typography
          variant="body2"
          sx={{ fontWeight: 600, color: 'text.primary', lineHeight: 1.3 }}
        >
          {record.name || '-'}
        </Typography>
        <StatusIndicator isEnabled={isEnabled} />
      </Box>
    </Box>
  );
};

const VendorField = () => {
  const record = useRecordContext<NASDevice>();
  if (!record) return null;
  
  const vendorName = getVendorName(record.vendor_code);
  return (
    <Chip
      label={vendorName}
      size="small"
      color="info"
      variant="outlined"
      sx={{ height: 22, fontSize: '0.75rem' }}
    />
  );
};

const IPAddressField = () => {
  const record = useRecordContext<NASDevice>();
  if (!record?.ipaddr) return <EmptyValue />;
  
  return (
    <Typography
      variant="body2"
      sx={{
        fontFamily: 'monospace',
        fontSize: '0.85rem',
        bgcolor: theme => alpha(theme.palette.info.main, 0.1),
        px: 1,
        py: 0.25,
        borderRadius: 1,
        display: 'inline-block',
      }}
    >
      {record.ipaddr}
    </Typography>
  );
};

// 标签显示组件
const TagsDisplay = ({ tags }: { tags?: string }) => {
  if (!tags) return <EmptyValue />;

  const tagList = tags.split(',').map((tag: string) => tag.trim()).filter((tag: string) => tag);
  
  if (tagList.length === 0) {
    return <EmptyValue />;
  }

  return (
    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
      {tagList.map((tag: string, index: number) => (
        <Chip
          key={index}
          label={tag}
          size="small"
          variant="outlined"
          color="primary"
          sx={{ height: 22, fontSize: '0.7rem' }}
        />
      ))}
    </Box>
  );
};

// ============ 表单工具栏 ============

const NASFormToolbar = (props: ToolbarProps) => (
  <Toolbar {...props}>
    <SaveButton />
    <DeleteButton mutationMode="pessimistic" />
  </Toolbar>
);

// ============ 列表操作栏组件 ============

const NASListActions = () => {
  const translate = useTranslate();
  return (
    <TopToolbar>
      <SortButton
        fields={['created_at', 'name', 'ipaddr']}
        label={translate('ra.action.sort', { _: '排序' })}
      />
      <CreateButton />
      <ExportButton />
    </TopToolbar>
  );
};

// ============ 内部列表内容组件 ============

const NASListContent = () => {
  const translate = useTranslate();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { data, isLoading, total } = useListContext<NASDevice>();

  const fieldLabels = useMemo(
    () => ({
      name: translate('resources.network/nas.fields.name', { _: '设备名称' }),
      ipaddr: translate('resources.network/nas.fields.ipaddr', { _: 'IP地址' }),
      identifier: translate('resources.network/nas.fields.identifier', { _: '标识符' }),
      status: translate('resources.network/nas.fields.status', { _: '状态' }),
    }),
    [translate],
  );

  const statusLabels = useMemo(
    () => ({
      enabled: translate('resources.network/nas.status.enabled', { _: '启用' }),
      disabled: translate('resources.network/nas.status.disabled', { _: '禁用' }),
    }),
    [translate],
  );

  if (isLoading) {
    return <NASListSkeleton />;
  }

  if (!data || data.length === 0) {
    return (
      <Box>
        <NASSearchHeaderCard />
        <Card
          elevation={0}
          sx={{
            borderRadius: 2,
            border: theme => `1px solid ${theme.palette.divider}`,
          }}
        >
          <NASEmptyState />
        </Card>
      </Box>
    );
  }

  return (
    <Box>
      {/* 搜索区块 */}
      <NASSearchHeaderCard />

      {/* 活动筛选标签 */}
      <ActiveFilters fieldLabels={fieldLabels} valueLabels={{ status: statusLabels }} />

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
            共 <strong>{total?.toLocaleString() || 0}</strong> 台 NAS 设备
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
              source="name"
              label={translate('resources.network/nas.fields.name', { _: '设备名称' })}
              render={() => <NASNameField />}
            />
            <FunctionField
              source="ipaddr"
              label={translate('resources.network/nas.fields.ipaddr', { _: 'IP地址' })}
              render={() => <IPAddressField />}
            />
            <TextField
              source="identifier"
              label={translate('resources.network/nas.fields.identifier', { _: '标识符' })}
            />
            <FunctionField
              source="vendor_code"
              label={translate('resources.network/nas.fields.vendor_code', { _: '厂商' })}
              render={() => <VendorField />}
            />
            <TextField
              source="model"
              label={translate('resources.network/nas.fields.model', { _: '型号' })}
            />
            <ReferenceField source="node_id" reference="network/nodes" label={translate('resources.network/nas.fields.node_id', { _: '所属节点' })} link="show">
              <TextField source="name" />
            </ReferenceField>
            <DateField
              source="created_at"
              label={translate('resources.network/nas.fields.created_at', { _: '创建时间' })}
              showTime
            />
          </Datagrid>
        </Box>
      </Card>
    </Box>
  );
};

// NAS 设备列表
export const NASList = () => {
  return (
    <List
      actions={<NASListActions />}
      sort={{ field: 'created_at', order: 'DESC' }}
      perPage={LARGE_LIST_PER_PAGE}
      pagination={<ServerPagination />}
      empty={false}
    >
      <NASListContent />
    </List>
  );
};

// ============ 编辑页面 ============

export const NASEdit = () => {
  const translate = useTranslate();
  
  return (
    <Edit>
      <SimpleForm toolbar={<NASFormToolbar />} sx={formLayoutSx}>
        <FormSection
          title={translate('resources.network/nas.sections.basic.title', { _: '基本信息' })}
          description={translate('resources.network/nas.sections.basic.description', { _: 'NAS 设备的基本配置' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
            <FieldGridItem>
              <TextInput
                source="id"
                disabled
                label={translate('resources.network/nas.fields.id', { _: '设备ID' })}
                helperText={translate('resources.network/nas.helpers.id', { _: '系统自动生成的唯一标识' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="name"
                label={translate('resources.network/nas.fields.name', { _: '设备名称' })}
                validate={[required(), minLength(1), maxLength(100)]}
                helperText={translate('resources.network/nas.helpers.name', { _: '1-100个字符的设备名称' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="identifier"
                label={translate('resources.network/nas.fields.identifier', { _: '标识符' })}
                validate={[required(), minLength(1), maxLength(100)]}
                helperText={translate('resources.network/nas.helpers.identifier', { _: 'NAS-Identifier属性值' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput
                source="vendor_code"
                label={translate('resources.network/nas.fields.vendor_code', { _: '厂商代码' })}
                validate={[required()]}
                choices={VENDOR_CHOICES}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="model"
                label={translate('resources.network/nas.fields.model', { _: '设备型号' })}
                validate={[maxLength(100)]}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput
                source="status"
                label={translate('resources.network/nas.fields.status', { _: '状态' })}
                validate={[required()]}
                choices={STATUS_CHOICES}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.network/nas.sections.network.title', { _: '网络配置' })}
          description={translate('resources.network/nas.sections.network.description', { _: 'IP地址和主机名配置' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
            <FieldGridItem>
              <TextInput
                source="ipaddr"
                label={translate('resources.network/nas.fields.ipaddr', { _: 'IP地址' })}
                validate={[required()]}
                helperText={translate('resources.network/nas.helpers.ipaddr', { _: 'NAS设备的IP地址' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="hostname"
                label={translate('resources.network/nas.fields.hostname', { _: '主机名' })}
                validate={[maxLength(200)]}
                helperText={translate('resources.network/nas.helpers.hostname', { _: 'NAS设备的主机名' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput
                source="coa_port"
                label={translate('resources.network/nas.fields.coa_port', { _: 'CoA端口' })}
                validate={[number(), minValue(1), maxValue(65535)]}
                helperText={translate('resources.network/nas.helpers.coa_port', { _: 'CoA/DM端口号 (1-65535)' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.network/nas.sections.radius.title', { _: 'RADIUS配置' })}
          description={translate('resources.network/nas.sections.radius.description', { _: 'RADIUS认证相关配置' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <PasswordInput
                source="secret"
                label={translate('resources.network/nas.fields.secret', { _: '共享密钥' })}
                validate={[required(), minLength(6)]}
                helperText={translate('resources.network/nas.helpers.secret', { _: 'RADIUS共享密钥，至少6位' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <ReferenceInput source="node_id" reference="network/nodes" label={translate('resources.network/nas.fields.node_id', { _: '所属节点' })}>
                <SelectInput optionText="name" fullWidth size="small" />
              </ReferenceInput>
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="tags"
                label={translate('resources.network/nas.fields.tags', { _: '标签' })}
                validate={[maxLength(200)]}
                helperText={translate('resources.network/nas.helpers.tags', { _: '多个标签用逗号分隔' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.network/nas.sections.remark.title', { _: '备注信息' })}
          description={translate('resources.network/nas.sections.remark.description', { _: '额外的说明和备注' })}
        >
          <FieldGrid columns={{ xs: 1 }}>
            <FieldGridItem>
              <TextInput
                source="remark"
                label={translate('resources.network/nas.fields.remark', { _: '备注' })}
                validate={[maxLength(500)]}
                multiline
                minRows={3}
                fullWidth
                size="small"
                helperText={translate('resources.network/nas.helpers.remark', { _: '可选的备注信息' })}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Edit>
  );
};

// ============ 创建页面 ============

export const NASCreate = () => {
  const translate = useTranslate();
  
  return (
    <Create>
      <SimpleForm sx={formLayoutSx}>
        <FormSection
          title={translate('resources.network/nas.sections.basic.title', { _: '基本信息' })}
          description={translate('resources.network/nas.sections.basic.description', { _: 'NAS 设备的基本配置' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
            <FieldGridItem>
              <TextInput
                source="name"
                label={translate('resources.network/nas.fields.name', { _: '设备名称' })}
                validate={[required(), minLength(1), maxLength(100)]}
                helperText={translate('resources.network/nas.helpers.name', { _: '1-100个字符的设备名称' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="identifier"
                label={translate('resources.network/nas.fields.identifier', { _: '标识符' })}
                validate={[required(), minLength(1), maxLength(100)]}
                helperText={translate('resources.network/nas.helpers.identifier', { _: 'NAS-Identifier属性值' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput
                source="vendor_code"
                label={translate('resources.network/nas.fields.vendor_code', { _: '厂商代码' })}
                validate={[required()]}
                choices={VENDOR_CHOICES}
                defaultValue="0"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="model"
                label={translate('resources.network/nas.fields.model', { _: '设备型号' })}
                validate={[maxLength(100)]}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput
                source="status"
                label={translate('resources.network/nas.fields.status', { _: '状态' })}
                validate={[required()]}
                choices={STATUS_CHOICES}
                defaultValue="enabled"
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.network/nas.sections.network.title', { _: '网络配置' })}
          description={translate('resources.network/nas.sections.network.description', { _: 'IP地址和主机名配置' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
            <FieldGridItem>
              <TextInput
                source="ipaddr"
                label={translate('resources.network/nas.fields.ipaddr', { _: 'IP地址' })}
                validate={[required()]}
                helperText={translate('resources.network/nas.helpers.ipaddr', { _: 'NAS设备的IP地址' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="hostname"
                label={translate('resources.network/nas.fields.hostname', { _: '主机名' })}
                validate={[maxLength(200)]}
                helperText={translate('resources.network/nas.helpers.hostname', { _: 'NAS设备的主机名' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput
                source="coa_port"
                label={translate('resources.network/nas.fields.coa_port', { _: 'CoA端口' })}
                validate={[number(), minValue(1), maxValue(65535)]}
                helperText={translate('resources.network/nas.helpers.coa_port', { _: 'CoA/DM端口号 (1-65535)' })}
                defaultValue={3799}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.network/nas.sections.radius.title', { _: 'RADIUS配置' })}
          description={translate('resources.network/nas.sections.radius.description', { _: 'RADIUS认证相关配置' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <PasswordInput
                source="secret"
                label={translate('resources.network/nas.fields.secret', { _: '共享密钥' })}
                validate={[required(), minLength(6)]}
                helperText={translate('resources.network/nas.helpers.secret', { _: 'RADIUS共享密钥，至少6位' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <ReferenceInput source="node_id" reference="network/nodes" label={translate('resources.network/nas.fields.node_id', { _: '所属节点' })}>
                <SelectInput optionText="name" fullWidth size="small" />
              </ReferenceInput>
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="tags"
                label={translate('resources.network/nas.fields.tags', { _: '标签' })}
                validate={[maxLength(200)]}
                helperText={translate('resources.network/nas.helpers.tags', { _: '多个标签用逗号分隔' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.network/nas.sections.remark.title', { _: '备注信息' })}
          description={translate('resources.network/nas.sections.remark.description', { _: '额外的说明和备注' })}
        >
          <FieldGrid columns={{ xs: 1 }}>
            <FieldGridItem>
              <TextInput
                source="remark"
                label={translate('resources.network/nas.fields.remark', { _: '备注' })}
                validate={[maxLength(500)]}
                multiline
                minRows={3}
                fullWidth
                size="small"
                helperText={translate('resources.network/nas.helpers.remark', { _: '可选的备注信息' })}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Create>
  );
};

// ============ 详情页顶部概览卡片 ============

const NASHeaderCard = () => {
  const record = useRecordContext<NASDevice>();
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
              ? `linear-gradient(135deg, ${alpha(theme.palette.primary.dark, 0.4)} 0%, ${alpha(theme.palette.success.dark, 0.3)} 100%)`
              : `linear-gradient(135deg, ${alpha(theme.palette.grey[800], 0.5)} 0%, ${alpha(theme.palette.grey[700], 0.3)} 100%)`
            : isEnabled
            ? `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.1)} 0%, ${alpha(theme.palette.success.main, 0.08)} 100%)`
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
          {/* 左侧：设备信息 */}
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
              {record.name?.charAt(0).toUpperCase() || 'N'}
            </Avatar>
            <Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
                <Typography variant="h5" sx={{ fontWeight: 700, color: 'text.primary' }}>
                  {record.name || <EmptyValue message="未知设备" />}
                </Typography>
                <StatusIndicator isEnabled={isEnabled} />
              </Box>
              {record.ipaddr && (
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 0.5 }}>
                  <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{ fontFamily: 'monospace' }}
                  >
                    {record.ipaddr}
                  </Typography>
                  <Tooltip title="复制IP地址">
                    <IconButton
                      size="small"
                      onClick={() => handleCopy(record.ipaddr!, 'IP地址')}
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
              <VendorIcon sx={{ fontSize: '1.1rem', color: 'info.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.network/nas.fields.vendor_code', { _: '厂商' })}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {getVendorName(record.vendor_code)}
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
              <ServerIcon sx={{ fontSize: '1.1rem', color: 'success.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.network/nas.fields.identifier', { _: '标识符' })}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600, fontFamily: 'monospace' }}>
              {record.identifier || '-'}
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
              <NasIcon sx={{ fontSize: '1.1rem', color: 'warning.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.network/nas.fields.model', { _: '型号' })}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {record.model || '-'}
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
              <NetworkIcon sx={{ fontSize: '1.1rem', color: 'primary.main' }} />
              <Typography variant="caption" color="text.secondary">
                {translate('resources.network/nas.fields.coa_port', { _: 'CoA端口' })}
              </Typography>
            </Box>
            <Typography variant="body2" sx={{ fontWeight: 600 }}>
              {record.coa_port || '-'}
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

// ============ NAS 详情内容 ============

const NASDetails = () => {
  const record = useRecordContext<NASDevice>();
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
          <NASHeaderCard />

          {/* 网络配置 */}
          <DetailSectionCard
            title={translate('resources.network/nas.sections.network.title', { _: '网络配置' })}
            description={translate('resources.network/nas.sections.network.description', { _: '主机名配置' })}
            icon={<NetworkIcon />}
            color="success"
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
                label={translate('resources.network/nas.fields.hostname', { _: '主机名' })}
                value={record.hostname || <EmptyValue />}
              />
            </Box>
          </DetailSectionCard>

          {/* RADIUS 配置 */}
          <DetailSectionCard
            title={translate('resources.network/nas.sections.radius.title', { _: 'RADIUS配置' })}
            description={translate('resources.network/nas.sections.radius.description', { _: 'RADIUS认证相关配置' })}
            icon={<SecretIcon />}
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
                label={translate('resources.network/nas.fields.node_id', { _: '所属节点' })}
                value={
                  record.node_id ? (
                    <ReferenceField source="node_id" reference="network/nodes" link="show">
                      <TextField source="name" />
                    </ReferenceField>
                  ) : <EmptyValue />
                }
              />
              <DetailItem
                label={translate('resources.network/nas.fields.tags', { _: '标签' })}
                value={<TagsDisplay tags={record.tags} />}
              />
            </Box>
          </DetailSectionCard>

          {/* 时间信息 */}
          <DetailSectionCard
            title={translate('resources.network/nas.sections.timestamps.title', { _: '时间信息' })}
            description={translate('resources.network/nas.sections.timestamps.description', { _: '创建和更新时间' })}
            icon={<TimeIcon />}
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
                label={translate('resources.network/nas.fields.created_at', { _: '创建时间' })}
                value={formatTimestamp(record.created_at)}
              />
              <DetailItem
                label={translate('resources.network/nas.fields.updated_at', { _: '更新时间' })}
                value={formatTimestamp(record.updated_at)}
              />
            </Box>
          </DetailSectionCard>

          {/* 备注信息 */}
          <DetailSectionCard
            title={translate('resources.network/nas.sections.remark.title', { _: '备注信息' })}
            description={translate('resources.network/nas.sections.remark.description', { _: '额外的说明和备注' })}
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
                {record.remark || translate('resources.network/nas.helpers.no_remark', { _: '无备注信息' })}
              </Typography>
            </Box>
          </DetailSectionCard>
        </Stack>
      </Box>
    </>
  );
};

// NAS 设备详情
export const NASShow = () => {
  return (
    <Show>
      <NASDetails />
    </Show>
  );
};
