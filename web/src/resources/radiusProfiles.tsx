import {
  List,
  Datagrid,
  TextField,
  DateField,
  Edit,
  TextInput,
  Create,
  Show,
  BooleanInput,
  NumberInput,
  required,
  minLength,
  maxLength,
  useRecordContext,
  Toolbar,
  SaveButton,
  DeleteButton,
  SimpleForm,
  ToolbarProps,
  TopToolbar,
  EditButton,
  ListButton,
  FilterButton,
  CreateButton,
  ExportButton,
  SelectInput
} from 'react-admin';
import {
  Box,
  Chip,
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
  Stack
} from '@mui/material';
import { Theme } from '@mui/material/styles';
import { ReactNode } from 'react';

// 状态显示组件
const StatusField = () => {
  const record = useRecordContext();
  if (!record) return null;

  return (
    <Chip
      label={record.status === 'enabled' ? '启用' : '禁用'}
      color={record.status === 'enabled' ? 'success' : 'default'}
      size="small"
    />
  );
};

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

const controlWrapperSx = {
  border: (theme: Theme) => `1px solid ${theme.palette.divider}`,
  borderRadius: 2,
  px: 2,
  py: 1.5,
  height: '100%',
  display: 'flex',
  alignItems: 'center',
  '& .MuiFormControl-root': {
    width: '100%',
    margin: 0
  },
  '& .MuiFormControlLabel-root': {
    margin: 0,
    width: '100%'
  }
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

const ProfileFormToolbar = (props: ToolbarProps) => (
  <Toolbar {...props}>
    <SaveButton />
    <DeleteButton mutationMode="pessimistic" />
  </Toolbar>
);

// RADIUS 计费策略列表操作栏
const ProfileListActions = () => (
  <TopToolbar>
    <FilterButton />
    <CreateButton />
    <ExportButton />
  </TopToolbar>
);

// RADIUS 计费策略过滤器
const profileFilters = [
  <TextInput key="name" label="策略名称" source="name" alwaysOn />,
  <TextInput key="addr_pool" label="地址池" source="addr_pool" />,
  <TextInput key="domain" label="域" source="domain" />,
  <SelectInput
    key="status"
    label="状态"
    source="status"
    choices={[
      { id: 'enabled', name: '启用' },
      { id: 'disabled', name: '禁用' },
    ]}
  />,
];

// RADIUS 计费策略列表
export const RadiusProfileList = () => (
  <List actions={<ProfileListActions />} filters={profileFilters}>
    <Datagrid rowClick="show">
      <TextField source="name" label="策略名称" />
      <StatusField />
      <TextField source="active_num" label="并发数" />
      <TextField source="up_rate" label="上行速率(Kbps)" />
      <TextField source="down_rate" label="下行速率(Kbps)" />
      <TextField source="addr_pool" label="地址池" />
      <TextField source="domain" label="域" />
      <DateField source="created_at" label="创建时间" showTime />
    </Datagrid>
  </List>
);

// RADIUS 计费策略编辑
export const RadiusProfileEdit = () => {
  return (
    <Edit>
      <SimpleForm toolbar={<ProfileFormToolbar />} sx={formLayoutSx}>
        <FormSection
          title="基本信息"
          description="策略的基本配置信息"
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="id"
                disabled
                label="策略ID"
                helperText="系统自动生成的唯一标识"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="name"
                label="策略名称"
                validate={[required(), minLength(2), maxLength(50)]}
                helperText="2-50个字符"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="status"
                  label="启用状态"
                  helperText="是否启用此策略"
                />
              </Box>
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title="并发和速率控制"
          description="并发数和带宽速率限制"
        >
          <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
            <FieldGridItem>
              <NumberInput
                source="active_num"
                label="并发数"
                min={0}
                helperText="允许的最大并发在线数，0表示不限制"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput
                source="up_rate"
                label="上行速率(Kbps)"
                min={0}
                helperText="上传带宽限制"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput
                source="down_rate"
                label="下行速率(Kbps)"
                min={0}
                helperText="下载带宽限制"
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title="网络配置"
          description="IP地址池和IPv6配置"
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="addr_pool"
                label="地址池"
                helperText="IP地址池名称"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="ipv6_prefix"
                label="IPv6前缀"
                helperText="如 2001:db8::/64"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="domain"
                label="域"
                helperText="对应NAS设备域属性，如华为domain_code"
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title="绑定策略"
          description="MAC和VLAN绑定控制"
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="bind_mac"
                  label="绑定MAC"
                  helperText="是否启用MAC地址绑定"
                />
              </Box>
            </FieldGridItem>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="bind_vlan"
                  label="绑定VLAN"
                  helperText="是否启用VLAN绑定"
                />
              </Box>
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title="备注信息"
          description="额外的说明和备注"
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="remark"
                label="备注"
                multiline
                minRows={3}
                fullWidth
                size="small"
                helperText="可选的备注信息，最多1000个字符"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Edit>
  );
};

// RADIUS 计费策略创建
export const RadiusProfileCreate = () => (
  <Create>
    <SimpleForm sx={formLayoutSx}>
      <FormSection
        title="基本信息"
        description="策略的基本配置信息"
      >
        <FieldGrid columns={{ xs: 1, sm: 2 }}>
          <FieldGridItem>
            <TextInput
              source="name"
              label="策略名称"
              validate={[required(), minLength(2), maxLength(50)]}
              helperText="2-50个字符"
              fullWidth
              size="small"
            />
          </FieldGridItem>
          <FieldGridItem>
            <Box sx={controlWrapperSx}>
              <BooleanInput
                source="status"
                label="启用状态"
                defaultValue={true}
                helperText="是否启用此策略"
              />
            </Box>
          </FieldGridItem>
        </FieldGrid>
      </FormSection>

      <FormSection
        title="并发和速率控制"
        description="并发数和带宽速率限制"
      >
        <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
          <FieldGridItem>
            <NumberInput
              source="active_num"
              label="并发数"
              min={0}
              defaultValue={1}
              helperText="允许的最大并发在线数，0表示不限制"
              fullWidth
              size="small"
            />
          </FieldGridItem>
          <FieldGridItem>
            <NumberInput
              source="up_rate"
              label="上行速率(Kbps)"
              min={0}
              defaultValue={1024}
              helperText="上传带宽限制"
              fullWidth
              size="small"
            />
          </FieldGridItem>
          <FieldGridItem>
            <NumberInput
              source="down_rate"
              label="下行速率(Kbps)"
              min={0}
              defaultValue={1024}
              helperText="下载带宽限制"
              fullWidth
              size="small"
            />
          </FieldGridItem>
        </FieldGrid>
      </FormSection>

      <FormSection
        title="网络配置"
        description="IP地址池和IPv6配置"
      >
        <FieldGrid columns={{ xs: 1, sm: 2 }}>
          <FieldGridItem>
            <TextInput
              source="addr_pool"
              label="地址池"
              helperText="IP地址池名称"
              fullWidth
              size="small"
            />
          </FieldGridItem>
          <FieldGridItem>
            <TextInput
              source="ipv6_prefix"
              label="IPv6前缀"
              helperText="如 2001:db8::/64"
              fullWidth
              size="small"
            />
          </FieldGridItem>
          <FieldGridItem span={{ xs: 1, sm: 2 }}>
            <TextInput
              source="domain"
              label="域"
              helperText="对应NAS设备域属性，如华为domain_code"
              fullWidth
              size="small"
            />
          </FieldGridItem>
        </FieldGrid>
      </FormSection>

      <FormSection
        title="绑定策略"
        description="MAC和VLAN绑定控制"
      >
        <FieldGrid columns={{ xs: 1, sm: 2 }}>
          <FieldGridItem>
            <Box sx={controlWrapperSx}>
              <BooleanInput
                source="bind_mac"
                label="绑定MAC"
                defaultValue={false}
                helperText="是否启用MAC地址绑定"
              />
            </Box>
          </FieldGridItem>
          <FieldGridItem>
            <Box sx={controlWrapperSx}>
              <BooleanInput
                source="bind_vlan"
                label="绑定VLAN"
                defaultValue={false}
                helperText="是否启用VLAN绑定"
              />
            </Box>
          </FieldGridItem>
        </FieldGrid>
      </FormSection>

      <FormSection
        title="备注信息"
        description="额外的说明和备注"
      >
        <FieldGrid columns={{ xs: 1, sm: 2 }}>
          <FieldGridItem span={{ xs: 1, sm: 2 }}>
            <TextInput
              source="remark"
              label="备注"
              multiline
              minRows={3}
              fullWidth
              size="small"
              helperText="可选的备注信息，最多1000个字符"
            />
          </FieldGridItem>
        </FieldGrid>
      </FormSection>
    </SimpleForm>
  </Create>
);

// 详情页工具栏
const ProfileShowActions = () => (
  <TopToolbar>
    <ListButton />
    <EditButton />
    <DeleteButton mutationMode="pessimistic" />
  </TopToolbar>
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

// 状态显示组件（用于详情页）
const StatusDisplay = () => {
  const record = useRecordContext();
  if (!record) return null;

  return (
    <Chip
      label={record.status === 'enabled' ? '启用' : '禁用'}
      color={record.status === 'enabled' ? 'success' : 'default'}
      size="small"
      sx={{ fontWeight: 500 }}
    />
  );
};

// 布尔值显示组件
const BooleanDisplay = ({ source }: { source: string }) => {
  const record = useRecordContext();
  if (!record) return null;

  const value = record[source];
  return (
    <Chip
      label={value ? '是' : '否'}
      color={value ? 'success' : 'default'}
      size="small"
      variant="outlined"
    />
  );
};

// RADIUS 计费策略详情
export const RadiusProfileShow = () => {
  return (
    <Show actions={<ProfileShowActions />}>
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
                      label="策略ID"
                      value={<TextField source="id" />}
                    />
                    <DetailRow
                      label="策略名称"
                      value={<TextField source="name" />}
                    />
                    <DetailRow
                      label="启用状态"
                      value={<StatusDisplay />}
                    />
                  </TableBody>
                </Table>
              </TableContainer>
            </CardContent>
          </Card>

          {/* 并发和速率控制卡片 */}
          <Box sx={{ display: 'grid', gap: 3, gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' } }}>
            <Card elevation={2}>
              <CardContent>
                <Typography variant="h6" gutterBottom sx={{ mb: 2, fontWeight: 600, color: 'primary.main' }}>
                  并发和速率控制
                </Typography>
                <Divider sx={{ mb: 2 }} />
                <TableContainer>
                  <Table>
                    <TableBody>
                      <DetailRow
                        label="并发数"
                        value={<TextField source="active_num" />}
                      />
                      <DetailRow
                        label="上行速率"
                        value={
                          <Box>
                            <TextField source="up_rate" />
                            <Typography component="span" variant="body2" color="text.secondary" sx={{ ml: 1 }}>
                              Kbps
                            </Typography>
                          </Box>
                        }
                      />
                      <DetailRow
                        label="下行速率"
                        value={
                          <Box>
                            <TextField source="down_rate" />
                            <Typography component="span" variant="body2" color="text.secondary" sx={{ ml: 1 }}>
                              Kbps
                            </Typography>
                          </Box>
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
                        label="地址池"
                        value={<TextField source="addr_pool" emptyText="-" />}
                      />
                      <DetailRow
                        label="IPv6前缀"
                        value={<TextField source="ipv6_prefix" emptyText="-" />}
                      />
                      <DetailRow
                        label="域"
                        value={<TextField source="domain" emptyText="-" />}
                      />
                    </TableBody>
                  </Table>
                </TableContainer>
              </CardContent>
            </Card>
          </Box>

          {/* 绑定策略和时间信息 */}
          <Box sx={{ display: 'grid', gap: 3, gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' } }}>
            <Card elevation={2}>
              <CardContent>
                <Typography variant="h6" gutterBottom sx={{ mb: 2, fontWeight: 600, color: 'primary.main' }}>
                  绑定策略
                </Typography>
                <Divider sx={{ mb: 2 }} />
                <TableContainer>
                  <Table>
                    <TableBody>
                      <DetailRow
                        label="绑定MAC"
                        value={<BooleanDisplay source="bind_mac" />}
                      />
                      <DetailRow
                        label="绑定VLAN"
                        value={<BooleanDisplay source="bind_vlan" />}
                      />
                    </TableBody>
                  </Table>
                </TableContainer>
              </CardContent>
            </Card>

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
          </Box>

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
