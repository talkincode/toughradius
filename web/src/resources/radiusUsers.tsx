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
  ExportButton,
  useTranslate,
  FilterLiveSearch,
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
import { ServerPagination } from '../components';

const DEFAULT_USER_PER_PAGE = 25;

// 状态显示组件
const StatusField = () => {
  const record = useRecordContext();
  const translate = useTranslate();
  if (!record) return null;

  return (
    <Chip
      label={translate(`resources.radius/users.status.${record.status === 'enabled' ? 'enabled' : 'disabled'}`)}
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

// 简化后的自定义工具栏
const UserFormToolbar = (props: ToolbarProps) => (
  <Toolbar {...props}>
    <SaveButton />
    <DeleteButton mutationMode="pessimistic" />
  </Toolbar>
);

// RADIUS 用户列表操作栏
const UserListActions = () => (
  <TopToolbar>
    <FilterLiveSearch source="q" />
    <FilterButton />
    <CreateButton />
    <ExportButton />
  </TopToolbar>
);

// RADIUS 用户过滤器
const useUserFilters = () => {
  const translate = useTranslate();
  return [
    <TextInput key="username" label={translate('resources.radius/users.fields.username')} source="username" alwaysOn />,
    <TextInput key="realname" label={translate('resources.radius/users.fields.realname')} source="realname" />,
    <TextInput key="email" label={translate('resources.radius/users.fields.email')} source="email" />,
    <TextInput key="mobile" label={translate('resources.radius/users.fields.mobile')} source="mobile" />,
    <SelectInput
      key="status"
      label={translate('resources.radius/users.fields.status')}
      source="status"
      choices={[
        { id: 'enabled', name: translate('resources.radius/users.status.enabled') },
        { id: 'disabled', name: translate('resources.radius/users.status.disabled') },
      ]}
    />,
    <ReferenceInput key="profile_id" source="profile_id" reference="radius/profiles">
      <SelectInput label={translate('resources.radius/users.fields.profile_id')} optionText="name" />
    </ReferenceInput>,
  ];
};

// RADIUS 用户列表
export const RadiusUserList = () => {
  const translate = useTranslate();
  const userFilters = useUserFilters();
  
  return (
    <List
      actions={<UserListActions />}
      filters={userFilters}
      perPage={DEFAULT_USER_PER_PAGE}
      pagination={<ServerPagination />}
    >
      <Datagrid rowClick="show">
        <TextField source="username" label={translate('resources.radius/users.fields.username')} />
        <TextField source="realname" label={translate('resources.radius/users.fields.realname')} />
        <EmailField source="email" label={translate('resources.radius/users.fields.email')} />
        <TextField source="mobile" label={translate('resources.radius/users.fields.mobile')} />
        <TextField source="address" label={translate('resources.radius/users.fields.address')} />
        <StatusField />
        <ReferenceField source="profile_id" reference="radius/profiles" label={translate('resources.radius/users.fields.profile_id')}>
          <TextField source="name" />
        </ReferenceField>
        <DateField source="created_at" label={translate('resources.radius/users.fields.created_at')} showTime />
        <DateField source="expire_time" label={translate('resources.radius/users.fields.expire_time')} showTime />
      </Datagrid>
    </List>
  );
};

// RADIUS 用户编辑
export const RadiusUserEdit = () => {
  const translate = useTranslate();
  
  return (
    <Edit>
      <SimpleForm toolbar={<UserFormToolbar />} sx={formLayoutSx}>
        <FormSection
          title={translate('resources.radius/users.sections.authentication')}
          description={translate('resources.radius/users.sections.authentication_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="id"
                disabled
                label={translate('resources.radius/users.fields.id')}
                helperText={translate('resources.radius/users.helpers.id')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="username"
                label={translate('resources.radius/users.fields.username')}
                validate={[required(), minLength(3), maxLength(50)]}
                helperText={translate('resources.radius/users.helpers.username')}
                autoComplete="username"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="password"
                label={translate('resources.radius/users.fields.password')}
                type="password"
                validate={[minLength(6), maxLength(128)]}
                helperText={translate('resources.radius/users.helpers.password')}
                autoComplete="new-password"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="realname"
                label={translate('resources.radius/users.fields.realname')}
                validate={[maxLength(100)]}
                helperText={translate('resources.radius/users.helpers.realname')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.contact')}
          description={translate('resources.radius/users.sections.contact_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="email"
                label={translate('resources.radius/users.fields.email')}
                type="email"
                validate={[email(), maxLength(100)]}
                helperText={translate('resources.radius/users.helpers.email')}
                autoComplete="email"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="mobile"
                label={translate('resources.radius/users.fields.mobile')}
                validate={[maxLength(20)]}
                helperText={translate('resources.radius/users.helpers.mobile')}
                autoComplete="tel"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="address"
                label={translate('resources.radius/users.fields.address')}
                multiline
                minRows={2}
                helperText={translate('resources.radius/users.helpers.address')}
                autoComplete="street-address"
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.service')}
          description={translate('resources.radius/users.sections.service_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="status"
                  label={translate('resources.radius/users.fields.status')}
                  helperText={translate('resources.radius/users.helpers.status')}
                />
              </Box>
            </FieldGridItem>
            <FieldGridItem>
              <ReferenceInput source="profile_id" reference="radius/profiles">
                <SelectInput
                  label={translate('resources.radius/users.fields.profile_id')}
                  optionText="name"
                  helperText={translate('resources.radius/users.helpers.profile_id')}
                  fullWidth
                  size="small"
                />
              </ReferenceInput>
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="expire_time"
                label={translate('resources.radius/users.fields.expire_time')}
                type="datetime-local"
                helperText={translate('resources.radius/users.helpers.expire_time')}
                fullWidth
                size="small"
                InputLabelProps={{ shrink: true }}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.network')}
          description={translate('resources.radius/users.sections.network_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="ip_addr"
                label={translate('resources.radius/users.fields.ip_addr')}
                helperText={translate('resources.radius/users.helpers.ip_addr')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="ipv6_addr"
                label={translate('resources.radius/users.fields.ipv6_addr')}
                helperText={translate('resources.radius/users.helpers.ipv6_addr')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="domain"
                label={translate('resources.radius/users.fields.domain')}
                helperText={translate('resources.radius/users.helpers.domain')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.remark')}
          description={translate('resources.radius/users.sections.remark_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="remark"
                label={translate('resources.radius/users.fields.remark')}
                multiline
                minRows={3}
                fullWidth
                size="small"
                helperText={translate('resources.radius/users.helpers.remark')}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Edit>
  );
};

// RADIUS 用户创建
export const RadiusUserCreate = () => {
  const translate = useTranslate();
  
  return (
    <Create>
      <SimpleForm sx={formLayoutSx}>
        <FormSection
          title={translate('resources.radius/users.sections.authentication')}
          description={translate('resources.radius/users.sections.authentication_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="username"
                label={translate('resources.radius/users.fields.username')}
                validate={[required(), minLength(3), maxLength(50)]}
                helperText={translate('resources.radius/users.helpers.username')}
                autoComplete="username"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="password"
                label={translate('resources.radius/users.fields.password')}
                type="password"
                validate={[required(), minLength(6), maxLength(128)]}
                helperText={translate('resources.radius/users.helpers.password_create')}
                autoComplete="new-password"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="realname"
                label={translate('resources.radius/users.fields.realname')}
                validate={[maxLength(100)]}
                helperText={translate('resources.radius/users.helpers.realname')}
                autoComplete="name"
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.contact')}
          description={translate('resources.radius/users.sections.contact_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="email"
                label={translate('resources.radius/users.fields.email')}
                type="email"
                validate={[email(), maxLength(100)]}
                helperText={translate('resources.radius/users.helpers.email')}
                autoComplete="email"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="mobile"
                label={translate('resources.radius/users.fields.mobile')}
                validate={[maxLength(20)]}
                helperText={translate('resources.radius/users.helpers.mobile')}
                autoComplete="tel"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="address"
                label={translate('resources.radius/users.fields.address')}
                multiline
                minRows={2}
                helperText={translate('resources.radius/users.helpers.address')}
                autoComplete="street-address"
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.service')}
          description={translate('resources.radius/users.sections.service_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="status"
                  label={translate('resources.radius/users.fields.status')}
                  defaultValue={true}
                  helperText={translate('resources.radius/users.helpers.status')}
                />
              </Box>
            </FieldGridItem>
            <FieldGridItem>
              <ReferenceInput source="profile_id" reference="radius/profiles">
                <SelectInput
                  label={translate('resources.radius/users.fields.profile_id')}
                  optionText="name"
                  helperText={translate('resources.radius/users.helpers.profile_id')}
                  fullWidth
                  size="small"
                />
              </ReferenceInput>
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="expire_time"
                label={translate('resources.radius/users.fields.expire_time')}
                type="datetime-local"
                helperText={translate('resources.radius/users.helpers.expire_time')}
                fullWidth
                size="small"
                InputLabelProps={{ shrink: true }}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.network')}
          description={translate('resources.radius/users.sections.network_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="ip_addr"
                label={translate('resources.radius/users.fields.ip_addr')}
                helperText={translate('resources.radius/users.helpers.ip_addr')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="ipv6_addr"
                label={translate('resources.radius/users.fields.ipv6_addr')}
                helperText={translate('resources.radius/users.helpers.ipv6_addr')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="domain"
                label={translate('resources.radius/users.fields.domain')}
                helperText={translate('resources.radius/users.helpers.domain')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/users.sections.remark')}
          description={translate('resources.radius/users.sections.remark_desc')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="remark"
                label={translate('resources.radius/users.fields.remark')}
                multiline
                minRows={3}
                fullWidth
                size="small"
                helperText={translate('resources.radius/users.helpers.remark')}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Create>
  );
};

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
  const translate = useTranslate();
  if (!record) return null;

  return (
    <Chip
      label={translate(`resources.radius/users.status.${record.status === 'enabled' ? 'enabled' : 'disabled'}`)}
      color={record.status === 'enabled' ? 'success' : 'default'}
      size="small"
      sx={{ fontWeight: 500 }}
    />
  );
};

// RADIUS 用户详情
export const RadiusUserShow = () => {
  const translate = useTranslate();
  
  return (
    <Show actions={<UserShowActions />}>
      <Box sx={{ width: '100%', p: { xs: 2, sm: 3, md: 4 } }}>
        <Stack spacing={3}>
          {/* 基本信息卡片 */}
          <Card elevation={2}>
            <CardContent>
              <Typography variant="h6" gutterBottom sx={{ mb: 2, fontWeight: 600, color: 'primary.main' }}>
                {translate('resources.radius/users.sections.basic')}
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <TableContainer>
                <Table>
                  <TableBody>
                    <DetailRow
                      label={translate('resources.radius/users.fields.id')}
                      value={<TextField source="id" />}
                    />
                    <DetailRow
                      label={translate('resources.radius/users.fields.username')}
                      value={<TextField source="username" />}
                    />
                    <DetailRow
                      label={translate('resources.radius/users.fields.realname')}
                      value={<TextField source="realname" emptyText="-" />}
                    />
                    <DetailRow
                      label={translate('resources.radius/users.fields.status')}
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
                  {translate('resources.radius/users.sections.contact')}
                </Typography>
                <Divider sx={{ mb: 2 }} />
                <TableContainer>
                  <Table>
                    <TableBody>
                      <DetailRow
                        label={translate('resources.radius/users.fields.email')}
                        value={<EmailField source="email" emptyText="-" />}
                      />
                      <DetailRow
                        label={translate('resources.radius/users.fields.mobile')}
                        value={<TextField source="mobile" emptyText="-" />}
                      />
                      <DetailRow
                        label={translate('resources.radius/users.fields.address')}
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
                  {translate('resources.radius/users.sections.service')}
                </Typography>
                <Divider sx={{ mb: 2 }} />
                <TableContainer>
                  <Table>
                    <TableBody>
                      <DetailRow
                        label={translate('resources.radius/users.fields.profile_id')}
                        value={
                          <ReferenceField source="profile_id" reference="radius/profiles" link="show">
                            <TextField source="name" />
                          </ReferenceField>
                        }
                      />
                      <DetailRow
                        label={translate('resources.radius/users.fields.expire_time')}
                        value={<DateField source="expire_time" showTime emptyText={translate('resources.radius/users.empty_text.expire_time')} />}
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
                {translate('resources.radius/users.sections.network')}
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <TableContainer>
                <Table>
                  <TableBody>
                    <DetailRow
                      label={translate('resources.radius/users.fields.ip_addr')}
                      value={<TextField source="ip_addr" emptyText="-" />}
                    />
                    <DetailRow
                      label={translate('resources.radius/users.fields.ipv6_addr')}
                      value={<TextField source="ipv6_addr" emptyText="-" />}
                    />
                    <DetailRow
                      label={translate('resources.radius/users.fields.domain')}
                      value={<TextField source="domain" emptyText="-" />}
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
                {translate('resources.radius/users.sections.time')}
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <TableContainer>
                <Table>
                  <TableBody>
                    <DetailRow
                      label={translate('resources.radius/users.fields.created_at')}
                      value={<DateField source="created_at" showTime />}
                    />
                    <DetailRow
                      label={translate('resources.radius/users.fields.updated_at')}
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
                {translate('resources.radius/users.sections.remark')}
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <Box sx={{ p: 2, backgroundColor: theme => theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.02)' : 'rgba(0, 0, 0, 0.01)', borderRadius: 1 }}>
                <TextField
                  source="remark"
                  emptyText={translate('resources.radius/users.empty_text.no_remark')}
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
