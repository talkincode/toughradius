import {
  List,
  Datagrid,
  TextField,
  DateField,
  Edit,
  TextInput,
  SelectInput,
  Create,
  Show,
  EmailField,
  BooleanInput,
  required,
  minLength,
  maxLength,
  email,
  useRecordContext,
  Toolbar,
  SaveButton,
  DeleteButton,
  SimpleForm,
  ToolbarProps,
  ReferenceInput,
  ReferenceField,
  TopToolbar,
  EditButton,
  ListButton,
  FilterButton,
  CreateButton,
  ExportButton
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

// 删除硬编码的配置选项
// const profileChoices = [
//   { id: 'default', name: '默认配置' },
//   { id: 'premium', name: '高级配置' },
//   { id: 'business', name: '企业配置' },
// ];

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

// 简化后的自定义工具栏（仅展示保存与删除）
const UserFormToolbar = (props: ToolbarProps) => (
  <Toolbar {...props}>
    <SaveButton />
    <DeleteButton mutationMode="pessimistic" />
  </Toolbar>
);

// RADIUS 用户列表操作栏
const UserListActions = () => (
  <TopToolbar>
    <FilterButton />
    <CreateButton />
    <ExportButton />
  </TopToolbar>
);

// RADIUS 用户过滤器
const userFilters = [
  <TextInput key="username" label="用户名" source="username" alwaysOn />,
  <TextInput key="realname" label="真实姓名" source="realname" />,
  <TextInput key="email" label="邮箱" source="email" />,
  <TextInput key="mobile" label="手机号" source="mobile" />,
  <SelectInput
    key="status"
    label="状态"
    source="status"
    choices={[
      { id: 'enabled', name: '启用' },
      { id: 'disabled', name: '禁用' },
    ]}
  />,
  <ReferenceInput key="profile_id" source="profile_id" reference="radius/profiles">
    <SelectInput label="计费策略" optionText="name" />
  </ReferenceInput>,
];

// RADIUS 用户列表
export const RadiusUserList = () => (
  <List actions={<UserListActions />} filters={userFilters}>
    <Datagrid rowClick="show">
      <TextField source="username" label="用户名" />
      <TextField source="realname" label="真实姓名" />
      <EmailField source="email" label="邮箱" />
      <TextField source="mobile" label="手机号" />
      <TextField source="address" label="地址" />
      <StatusField />
      <ReferenceField source="profile_id" reference="radius/profiles" label="计费策略">
        <TextField source="name" />
      </ReferenceField>
      <DateField source="created_at" label="创建时间" showTime />
      <DateField source="expire_time" label="过期时间" showTime />
    </Datagrid>
  </List>
);

// RADIUS 用户编辑
export const RadiusUserEdit = () => {
  return (
    <Edit>
      <SimpleForm toolbar={<UserFormToolbar />} sx={formLayoutSx}>
        <FormSection
          title="身份认证"
          description="用户的基本认证信息"
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="id"
                disabled
                label="用户ID"
                helperText="系统自动生成的唯一标识"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="username"
                label="用户名"
                validate={[required(), minLength(3), maxLength(50)]}
                helperText="3-50个字符，只能包含字母、数字、下划线"
                autoComplete="username"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="password"
                label="密码"
                type="password"
                validate={[minLength(6), maxLength(128)]}
                helperText="留空则不修改密码"
                autoComplete="new-password"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="realname"
                label="真实姓名"
                validate={[maxLength(100)]}
                helperText="用户的真实姓名"
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title="联系方式"
          description="联系信息和地址"
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="email"
                label="邮箱"
                type="email"
                validate={[email(), maxLength(100)]}
                helperText="用于接收通知和找回密码"
                autoComplete="email"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="mobile"
                label="手机号"
                validate={[maxLength(20)]}
                helperText="手机号码（可选），最多20个字符"
                autoComplete="tel"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="address"
                label="地址"
                multiline
                minRows={2}
                helperText="详细地址信息"
                autoComplete="street-address"
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title="服务配置"
          description="RADIUS服务和权限设置"
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="status"
                  label="启用状态"
                  helperText="是否启用此用户的RADIUS服务"
                />
              </Box>
            </FieldGridItem>
            <FieldGridItem>
              <ReferenceInput source="profile_id" reference="radius/profiles">
                <SelectInput
                  label="计费策略"
                  optionText="name"
                  helperText="选择用户的RADIUS计费策略"
                  fullWidth
                  size="small"
                />
              </ReferenceInput>
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="expire_time"
                label="过期时间"
                type="datetime-local"
                helperText="用户服务到期时间，留空表示永不过期"
                fullWidth
                size="small"
                InputLabelProps={{ shrink: true }}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title="网络配置"
          description="IP地址分配设置"
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="ip_addr"
                label="IPv4地址"
                helperText="静态IPv4地址，如 192.168.1.100"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="ipv6_addr"
                label="IPv6地址"
                helperText="静态IPv6地址，如 2001:db8::1"
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

// RADIUS 用户创建
export const RadiusUserCreate = () => (
  <Create>
    <SimpleForm sx={formLayoutSx}>
      <FormSection
        title="身份认证"
        description="用户的基本认证信息"
      >
        <FieldGrid columns={{ xs: 1, sm: 2 }}>
          <FieldGridItem>
            <TextInput
              source="username"
              label="用户名"
              validate={[required(), minLength(3), maxLength(50)]}
              helperText="3-50个字符，只能包含字母、数字、下划线"
              autoComplete="username"
              fullWidth
              size="small"
            />
          </FieldGridItem>
          <FieldGridItem>
            <TextInput
              source="password"
              label="密码"
              type="password"
              validate={[required(), minLength(6), maxLength(128)]}
              helperText="6-128个字符的密码"
              autoComplete="new-password"
              fullWidth
              size="small"
            />
          </FieldGridItem>
          <FieldGridItem span={{ xs: 1, sm: 2 }}>
            <TextInput
              source="realname"
              label="真实姓名"
              validate={[maxLength(100)]}
              helperText="用户的真实姓名"
              autoComplete="name"
              fullWidth
              size="small"
            />
          </FieldGridItem>
        </FieldGrid>
      </FormSection>

        <FormSection
          title="联系方式"
          description="联系信息和地址"
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="email"
                label="邮箱"
                type="email"
                validate={[email(), maxLength(100)]}
                helperText="用于接收通知和找回密码"
                autoComplete="email"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="mobile"
                label="手机号"
                validate={[maxLength(20)]}
                helperText="手机号码（可选），最多20个字符"
                autoComplete="tel"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="address"
                label="地址"
                multiline
                minRows={2}
                helperText="详细地址信息"
                autoComplete="street-address"
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

      <FormSection
        title="服务配置"
        description="RADIUS服务和权限设置"
      >
        <FieldGrid columns={{ xs: 1, sm: 2 }}>
          <FieldGridItem>
            <Box sx={controlWrapperSx}>
              <BooleanInput
                source="status"
                label="启用状态"
                defaultValue={true}
                helperText="是否启用此用户的RADIUS服务"
              />
            </Box>
          </FieldGridItem>
          <FieldGridItem>
            <ReferenceInput source="profile_id" reference="radius/profiles">
              <SelectInput
                label="计费策略"
                optionText="name"
                helperText="选择用户的RADIUS计费策略"
                fullWidth
                size="small"
              />
            </ReferenceInput>
          </FieldGridItem>
          <FieldGridItem span={{ xs: 1, sm: 2 }}>
            <TextInput
              source="expire_time"
              label="过期时间"
              type="datetime-local"
              helperText="用户服务到期时间，留空表示永不过期"
              fullWidth
              size="small"
              InputLabelProps={{ shrink: true }}
            />
          </FieldGridItem>
        </FieldGrid>
      </FormSection>

      <FormSection
        title="网络配置"
        description="IP地址分配设置"
      >
        <FieldGrid columns={{ xs: 1, sm: 2 }}>
          <FieldGridItem>
            <TextInput
              source="ip_addr"
              label="IPv4地址"
              helperText="静态IPv4地址，如 192.168.1.100"
              fullWidth
              size="small"
            />
          </FieldGridItem>
          <FieldGridItem>
            <TextInput
              source="ipv6_addr"
              label="IPv6地址"
              helperText="静态IPv6地址，如 2001:db8::1"
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
const UserShowActions = () => (
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

// RADIUS 用户详情
export const RadiusUserShow = () => {
  return (
    <Show actions={<UserShowActions />}>
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
                      label="用户ID"
                      value={<TextField source="id" />}
                    />
                    <DetailRow
                      label="用户名"
                      value={<TextField source="username" />}
                    />
                    <DetailRow
                      label="真实姓名"
                      value={<TextField source="realname" emptyText="-" />}
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

          {/* 联系方式和服务配置 */}
          <Box sx={{ display: 'grid', gap: 3, gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' } }}>
            <Card elevation={2}>
              <CardContent>
                <Typography variant="h6" gutterBottom sx={{ mb: 2, fontWeight: 600, color: 'primary.main' }}>
                  联系方式
                </Typography>
                <Divider sx={{ mb: 2 }} />
                <TableContainer>
                  <Table>
                    <TableBody>
                      <DetailRow
                        label="邮箱"
                        value={<EmailField source="email" emptyText="-" />}
                      />
                      <DetailRow
                        label="手机号"
                        value={<TextField source="mobile" emptyText="-" />}
                      />
                      <DetailRow
                        label="地址"
                        value={<TextField source="address" emptyText="-" />}
                      />
                    </TableBody>
                  </Table>
                </TableContainer>
              </CardContent>
            </Card>

            <Card elevation={2}>
              <CardContent>
                <Typography variant="h6" gutterBottom sx={{ mb: 2, fontWeight: 600, color: 'primary.main' }}>
                  服务配置
                </Typography>
                <Divider sx={{ mb: 2 }} />
                <TableContainer>
                  <Table>
                    <TableBody>
                      <DetailRow
                        label="计费策略"
                        value={
                          <ReferenceField source="profile_id" reference="radius/profiles" link="show">
                            <TextField source="name" />
                          </ReferenceField>
                        }
                      />
                      <DetailRow
                        label="过期时间"
                        value={<DateField source="expire_time" showTime emptyText="永不过期" />}
                      />
                    </TableBody>
                  </Table>
                </TableContainer>
              </CardContent>
            </Card>
          </Box>

          {/* 网络配置 */}
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
                      label="IPv4地址"
                      value={<TextField source="ip_addr" emptyText="-" />}
                    />
                    <DetailRow
                      label="IPv6地址"
                      value={<TextField source="ipv6_addr" emptyText="-" />}
                    />
                  </TableBody>
                </Table>
              </TableContainer>
            </CardContent>
          </Card>

          {/* 时间信息 */}
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
