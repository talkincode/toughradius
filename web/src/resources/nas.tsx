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
  FilterButton,
  TopToolbar,
  CreateButton,
  ExportButton,
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
  EditButton,
  ListButton,
} from 'react-admin';
import {
  Box,
  Typography,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableRow,
  Card,
  CardContent,
  Divider,
  Stack,
  Chip,
} from '@mui/material';
import { ReactNode } from 'react';

// ============ 共享组件 ============

interface FormSectionProps {
  title: string;
  description?: string;
  children: ReactNode;
}

const FormSection = ({ title, description, children }: FormSectionProps) => (
  <Paper
    elevation={0}
    sx={{
      p: 3,
      mb: 3,
      borderRadius: 2,
      border: theme => `1px solid ${theme.palette.divider}`,
      backgroundColor: theme => theme.palette.background.paper,
      width: '100%'
    }}
  >
    <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
      {title}
    </Typography>
    {description && (
      <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5, mb: 1 }}>
        {description}
      </Typography>
    )}
    <Box sx={{ mt: 2, width: '100%' }}>
      {children}
    </Box>
  </Paper>
);

type ColumnConfig = {
  xs?: number;
  sm?: number;
  md?: number;
  lg?: number;
  xl?: number;
};

interface FieldGridProps {
  children: ReactNode;
  columns?: ColumnConfig;
  gap?: number;
}

const defaultColumns: Required<Pick<ColumnConfig, 'xs' | 'sm' | 'md' | 'lg'>> = {
  xs: 1,
  sm: 2,
  md: 3,
  lg: 3
};

const FieldGrid = ({
  children,
  columns = {},
  gap = 2
}: FieldGridProps) => {
  const resolved = {
    xs: columns.xs ?? defaultColumns.xs,
    sm: columns.sm ?? defaultColumns.sm,
    md: columns.md ?? defaultColumns.md,
    lg: columns.lg ?? defaultColumns.lg
  };

  return (
    <Box
      sx={{
        display: 'grid',
        gap,
        width: '100%',
        alignItems: 'stretch',
        justifyItems: 'stretch',
        gridTemplateColumns: {
          xs: `repeat(${resolved.xs}, minmax(0, 1fr))`,
          sm: `repeat(${resolved.sm}, minmax(0, 1fr))`,
          md: `repeat(${resolved.md}, minmax(0, 1fr))`,
          lg: `repeat(${resolved.lg}, minmax(0, 1fr))`
        }
      }}
    >
      {children}
    </Box>
  );
};

interface FieldGridItemProps {
  children: ReactNode;
  span?: ColumnConfig;
}

const FieldGridItem = ({
  children,
  span = {}
}: FieldGridItemProps) => {
  const resolved = {
    xs: span.xs ?? 1,
    sm: span.sm ?? span.xs ?? 1,
    md: span.md ?? span.sm ?? span.xs ?? 1,
    lg: span.lg ?? span.md ?? span.sm ?? span.xs ?? 1
  };

  return (
    <Box
      sx={{
        width: '100%',
        gridColumn: {
          xs: `span ${resolved.xs}`,
          sm: `span ${resolved.sm}`,
          md: `span ${resolved.md}`,
          lg: `span ${resolved.lg}`
        }
      }}
    >
      {children}
    </Box>
  );
};

const formLayoutSx = {
  width: '100%',
  maxWidth: 'none',
  mx: 0,
  px: { xs: 2, sm: 3, md: 4 },
  '& .RaSimpleForm-main': {
    width: '100%',
    maxWidth: 'none',
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'flex-start'
  },
  '& .RaSimpleForm-content': {
    width: '100%',
    maxWidth: 'none',
    px: 0
  },
  '& .RaSimpleForm-form': {
    width: '100%',
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'flex-start'
  },
  '& .RaSimpleForm-input': {
    width: '100%'
  }
};

const NASFormToolbar = (props: ToolbarProps) => (
  <Toolbar {...props}>
    <SaveButton />
    <DeleteButton mutationMode="pessimistic" />
  </Toolbar>
);

// 详情信息行组件
const DetailRow = ({ label, value }: { label: string; value: ReactNode }) => (
  <TableRow>
    <TableCell
      component="th"
      scope="row"
      sx={{
        fontWeight: 600,
        backgroundColor: theme => theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.02)',
        width: '30%',
        py: 1.5,
        px: 2
      }}
    >
      {label}
    </TableCell>
    <TableCell sx={{ py: 1.5, px: 2 }}>
      {value}
    </TableCell>
  </TableRow>
);

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
const getVendorName = (code: string): string => {
  const vendor = VENDOR_CHOICES.find(v => v.id === String(code));
  return vendor ? vendor.name : code;
};

// ============ 列表相关 ============

// 筛选器
const nasFilters = [
  <TextInput key="name" label="设备名称" source="name" alwaysOn />,
  <TextInput key="ipaddr" label="设备IP" source="ipaddr" />,
  <SelectInput
    key="status"
    label="状态"
    source="status"
    choices={STATUS_CHOICES}
  />,
];

// NAS 列表操作栏
const NASListActions = () => (
  <TopToolbar>
    <FilterButton />
    <CreateButton />
    <ExportButton />
  </TopToolbar>
);

// 状态 Chip 字段
const StatusChipField = () => {
  const record = useRecordContext();
  if (!record) return null;
  
  return (
    <Chip
      label={record.status === 'enabled' ? '启用' : '禁用'}
      color={record.status === 'enabled' ? 'success' : 'default'}
      size="small"
      variant="outlined"
    />
  );
};

// NAS 设备列表
export const NASList = () => (
  <List actions={<NASListActions />} filters={nasFilters}>
    <Datagrid rowClick="show">
      <TextField source="name" label="设备名称" />
      <TextField source="ipaddr" label="设备IP" />
      <TextField source="identifier" label="设备标识" />
      <TextField source="model" label="设备型号" />
      <StatusChipField />
      <ReferenceField source="node_id" reference="network/nodes" label="所属节点" link="show">
        <TextField source="name" />
      </ReferenceField>
      <DateField source="created_at" label="创建时间" showTime />
    </Datagrid>
  </List>
);

// ============ 编辑页面 ============

export const NASEdit = () => {
  return (
    <Edit>
      <SimpleForm toolbar={<NASFormToolbar />} sx={formLayoutSx}>
        <FormSection
          title="基本信息"
          description="NAS设备的基本配置信息"
        >
          <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
            <FieldGridItem>
              <TextInput
                source="id"
                disabled
                label="设备ID"
                helperText="系统自动生成的唯一标识"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="name"
                label="设备名称"
                validate={[required(), minLength(1), maxLength(100)]}
                helperText="1-100个字符"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="identifier"
                label="设备标识-RADIUS"
                validate={[required(), minLength(1), maxLength(100)]}
                helperText="RADIUS认证标识"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput
                source="vendor_code"
                label="厂商代码"
                validate={[required()]}
                choices={VENDOR_CHOICES}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="model"
                label="设备型号"
                validate={[maxLength(100)]}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput
                source="status"
                label="状态"
                validate={[required()]}
                choices={STATUS_CHOICES}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title="网络配置"
          description="设备的网络地址和端口配置"
        >
          <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
            <FieldGridItem>
              <TextInput
                source="ipaddr"
                label="设备IP地址"
                validate={[required()]}
                helperText="IPv4或IPv6地址"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="hostname"
                label="设备主机地址"
                validate={[maxLength(200)]}
                helperText="可选的主机名"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput
                source="coa_port"
                label="COA端口"
                validate={[number(), minValue(1), maxValue(65535)]}
                helperText="1-65535，默认3799"
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title="RADIUS配置"
          description="RADIUS认证和计费配置"
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <PasswordInput
                source="secret"
                label="RADIUS秘钥"
                validate={[required(), minLength(6)]}
                helperText="至少6个字符"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <ReferenceInput source="node_id" reference="network/nodes" label="所属节点">
                <SelectInput optionText="name" fullWidth size="small" />
              </ReferenceInput>
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="tags"
                label="标签"
                validate={[maxLength(200)]}
                helperText="多个标签用逗号分隔，最多200个字符"
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title="备注信息"
          description="额外的说明和备注"
        >
          <FieldGrid columns={{ xs: 1 }}>
            <FieldGridItem>
              <TextInput
                source="remark"
                label="备注"
                validate={[maxLength(500)]}
                multiline
                minRows={3}
                fullWidth
                size="small"
                helperText="可选的备注信息，最多500个字符"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Edit>
  );
};

// ============ 创建页面 ============

export const NASCreate = () => (
  <Create>
    <SimpleForm sx={formLayoutSx}>
      <FormSection
        title="基本信息"
        description="NAS设备的基本配置信息"
      >
        <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
          <FieldGridItem>
            <TextInput
              source="name"
              label="设备名称"
              validate={[required(), minLength(1), maxLength(100)]}
              helperText="1-100个字符"
              fullWidth
              size="small"
            />
          </FieldGridItem>
          <FieldGridItem>
            <TextInput
              source="identifier"
              label="设备标识-RADIUS"
              validate={[required(), minLength(1), maxLength(100)]}
              helperText="RADIUS认证标识"
              fullWidth
              size="small"
            />
          </FieldGridItem>
          <FieldGridItem>
            <SelectInput
              source="vendor_code"
              label="厂商代码"
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
              label="设备型号"
              validate={[maxLength(100)]}
              fullWidth
              size="small"
            />
          </FieldGridItem>
          <FieldGridItem>
            <SelectInput
              source="status"
              label="状态"
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
        title="网络配置"
        description="设备的网络地址和端口配置"
      >
        <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
          <FieldGridItem>
            <TextInput
              source="ipaddr"
              label="设备IP地址"
              validate={[required()]}
              helperText="IPv4或IPv6地址"
              fullWidth
              size="small"
            />
          </FieldGridItem>
          <FieldGridItem>
            <TextInput
              source="hostname"
              label="设备主机地址"
              validate={[maxLength(200)]}
              helperText="可选的主机名"
              fullWidth
              size="small"
            />
          </FieldGridItem>
          <FieldGridItem>
            <NumberInput
              source="coa_port"
              label="COA端口"
              validate={[number(), minValue(1), maxValue(65535)]}
              helperText="1-65535，默认3799"
              defaultValue={3799}
              fullWidth
              size="small"
            />
          </FieldGridItem>
        </FieldGrid>
      </FormSection>

      <FormSection
        title="RADIUS配置"
        description="RADIUS认证和计费配置"
      >
        <FieldGrid columns={{ xs: 1, sm: 2 }}>
          <FieldGridItem>
            <PasswordInput
              source="secret"
              label="RADIUS秘钥"
              validate={[required(), minLength(6)]}
              helperText="至少6个字符"
              fullWidth
              size="small"
            />
          </FieldGridItem>
          <FieldGridItem>
            <ReferenceInput source="node_id" reference="network/nodes" label="所属节点">
              <SelectInput optionText="name" fullWidth size="small" />
            </ReferenceInput>
          </FieldGridItem>
          <FieldGridItem span={{ xs: 1, sm: 2 }}>
            <TextInput
              source="tags"
              label="标签"
              validate={[maxLength(200)]}
              helperText="多个标签用逗号分隔，最多200个字符"
              fullWidth
              size="small"
            />
          </FieldGridItem>
        </FieldGrid>
      </FormSection>

      <FormSection
        title="备注信息"
        description="额外的说明和备注"
      >
        <FieldGrid columns={{ xs: 1 }}>
          <FieldGridItem>
            <TextInput
              source="remark"
              label="备注"
              validate={[maxLength(500)]}
              multiline
              minRows={3}
              fullWidth
              size="small"
              helperText="可选的备注信息，最多500个字符"
            />
          </FieldGridItem>
        </FieldGrid>
      </FormSection>
    </SimpleForm>
  </Create>
);

// ============ 详情页面 ============

// 详情页工具栏
const NASShowActions = () => (
  <TopToolbar>
    <ListButton />
    <EditButton />
    <DeleteButton mutationMode="pessimistic" />
  </TopToolbar>
);

// 标签显示组件
const TagsField = () => {
  const record = useRecordContext();
  if (!record || !record.tags) return <Typography variant="body2" color="text.secondary">-</Typography>;

  const tags = record.tags.split(',').map((tag: string) => tag.trim()).filter((tag: string) => tag);
  
  if (tags.length === 0) {
    return <Typography variant="body2" color="text.secondary">-</Typography>;
  }

  return (
    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
      {tags.map((tag: string, index: number) => (
        <Chip
          key={index}
          label={tag}
          size="small"
          variant="outlined"
          color="primary"
        />
      ))}
    </Box>
  );
};

// 厂商代码显示
const VendorCodeField = () => {
  const record = useRecordContext();
  if (!record) return null;
  
  return (
    <Typography variant="body2">
      {getVendorName(record.vendor_code)} ({record.vendor_code})
    </Typography>
  );
};

export const NASShow = () => {
  return (
    <Show actions={<NASShowActions />}>
      <Box sx={{ width: '100%', p: { xs: 2, sm: 3, md: 4 } }}>
        <Stack spacing={3}>
          {/* 基本信息卡片 */}
          <Card elevation={2}>
            <CardContent>
              <Typography variant="h6" gutterBottom sx={{ mb: 2, fontWeight: 600, color: 'primary.main' }}>
                基本信息
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <TableContainer>
                <Table>
                  <TableBody>
                    <DetailRow
                      label="设备ID"
                      value={<TextField source="id" />}
                    />
                    <DetailRow
                      label="设备名称"
                      value={<TextField source="name" />}
                    />
                    <DetailRow
                      label="设备标识"
                      value={<TextField source="identifier" />}
                    />
                    <DetailRow
                      label="厂商代码"
                      value={<VendorCodeField />}
                    />
                    <DetailRow
                      label="设备型号"
                      value={<TextField source="model" emptyText="-" />}
                    />
                    <DetailRow
                      label="状态"
                      value={<StatusChipField />}
                    />
                    <DetailRow
                      label="所属节点"
                      value={
                        <ReferenceField source="node_id" reference="network/nodes" link="show">
                          <TextField source="name" />
                        </ReferenceField>
                      }
                    />
                  </TableBody>
                </Table>
              </TableContainer>
            </CardContent>
          </Card>

          {/* 网络配置卡片 */}
          <Card elevation={2}>
            <CardContent>
              <Typography variant="h6" gutterBottom sx={{ mb: 2, fontWeight: 600, color: 'primary.main' }}>
                网络配置
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <TableContainer>
                <Table>
                  <TableBody>
                    <DetailRow
                      label="设备IP地址"
                      value={<TextField source="ipaddr" />}
                    />
                    <DetailRow
                      label="设备主机地址"
                      value={<TextField source="hostname" emptyText="-" />}
                    />
                    <DetailRow
                      label="COA端口"
                      value={<TextField source="coa_port" />}
                    />
                    <DetailRow
                      label="标签"
                      value={<TagsField />}
                    />
                  </TableBody>
                </Table>
              </TableContainer>
            </CardContent>
          </Card>

          {/* 时间信息卡片 */}
          <Card elevation={2}>
            <CardContent>
              <Typography variant="h6" gutterBottom sx={{ mb: 2, fontWeight: 600, color: 'primary.main' }}>
                时间信息
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <TableContainer>
                <Table>
                  <TableBody>
                    <DetailRow
                      label="创建时间"
                      value={<DateField source="created_at" showTime />}
                    />
                    <DetailRow
                      label="更新时间"
                      value={<DateField source="updated_at" showTime />}
                    />
                  </TableBody>
                </Table>
              </TableContainer>
            </CardContent>
          </Card>

          {/* 备注信息卡片 */}
          <Card elevation={2}>
            <CardContent>
              <Typography variant="h6" gutterBottom sx={{ mb: 2, fontWeight: 600, color: 'primary.main' }}>
                备注信息
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <Box sx={{ p: 2, backgroundColor: theme => theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.02)' : 'rgba(0, 0, 0, 0.01)', borderRadius: 1 }}>
                <TextField
                  source="remark"
                  emptyText="无备注信息"
                  sx={{
                    '& .RaTextField-root': {
                      whiteSpace: 'pre-wrap',
                      wordBreak: 'break-word'
                    }
                  }}
                />
              </Box>
            </CardContent>
          </Card>
        </Stack>
      </Box>
    </Show>
  );
};
