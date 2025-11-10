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
  SelectInput,
  useTranslate
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
  const translate = useTranslate();
  if (!record) return null;

  return (
    <Chip
      label={translate(`resources.radius/profiles.status.${record.status}`)}
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
const useProfileFilters = () => {
  const translate = useTranslate();
  
  return [
    <TextInput key="name" label={translate('resources.radius/profiles.fields.name')} source="name" alwaysOn />,
    <TextInput key="addr_pool" label={translate('resources.radius/profiles.fields.addr_pool')} source="addr_pool" />,
    <TextInput key="domain" label={translate('resources.radius/profiles.fields.domain')} source="domain" />,
    <SelectInput
      key="status"
      label={translate('resources.radius/profiles.fields.status')}
      source="status"
      choices={[
        { id: 'enabled', name: translate('resources.radius/profiles.status.enabled') },
        { id: 'disabled', name: translate('resources.radius/profiles.status.disabled') },
      ]}
    />,
  ];
};

// RADIUS 计费策略列表
export const RadiusProfileList = () => {
  const translate = useTranslate();
  
  return (
    <List actions={<ProfileListActions />} filters={useProfileFilters()}>
      <Datagrid rowClick="show">
        <TextField source="name" label={translate('resources.radius/profiles.fields.name')} />
        <StatusField />
        <TextField source="active_num" label={translate('resources.radius/profiles.fields.active_num')} />
        <TextField source="up_rate" label={translate('resources.radius/profiles.fields.up_rate')} />
        <TextField source="down_rate" label={translate('resources.radius/profiles.fields.down_rate')} />
        <TextField source="addr_pool" label={translate('resources.radius/profiles.fields.addr_pool')} />
        <TextField source="domain" label={translate('resources.radius/profiles.fields.domain')} />
        <DateField source="created_at" label={translate('resources.radius/profiles.fields.created_at')} showTime />
      </Datagrid>
    </List>
  );
};

// RADIUS 计费策略编辑
export const RadiusProfileEdit = () => {
  const translate = useTranslate();
  
  return (
    <Edit>
      <SimpleForm toolbar={<ProfileFormToolbar />} sx={formLayoutSx}>
        <FormSection
          title={translate('resources.radius/profiles.sections.basic.title')}
          description={translate('resources.radius/profiles.sections.basic.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="id"
                disabled
                label={translate('resources.radius/profiles.fields.id')}
                helperText={translate('resources.radius/profiles.helpers.id')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="name"
                label={translate('resources.radius/profiles.fields.name')}
                validate={[required(), minLength(2), maxLength(50)]}
                helperText={translate('resources.radius/profiles.helpers.name')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="status"
                  label={translate('resources.radius/profiles.fields.status_enabled')}
                  helperText={translate('resources.radius/profiles.helpers.status')}
                />
              </Box>
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.rate_control.title')}
          description={translate('resources.radius/profiles.sections.rate_control.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
            <FieldGridItem>
              <NumberInput
                source="active_num"
                label={translate('resources.radius/profiles.fields.active_num')}
                min={0}
                helperText={translate('resources.radius/profiles.helpers.active_num')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput
                source="up_rate"
                label={translate('resources.radius/profiles.fields.up_rate')}
                min={0}
                helperText={translate('resources.radius/profiles.helpers.up_rate')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput
                source="down_rate"
                label={translate('resources.radius/profiles.fields.down_rate')}
                min={0}
                helperText={translate('resources.radius/profiles.helpers.down_rate')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.network.title')}
          description={translate('resources.radius/profiles.sections.network.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="addr_pool"
                label={translate('resources.radius/profiles.fields.addr_pool')}
                helperText={translate('resources.radius/profiles.helpers.addr_pool')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="ipv6_prefix"
                label={translate('resources.radius/profiles.fields.ipv6_prefix')}
                helperText={translate('resources.radius/profiles.helpers.ipv6_prefix')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="domain"
                label={translate('resources.radius/profiles.fields.domain')}
                helperText={translate('resources.radius/profiles.helpers.domain')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.binding.title')}
          description={translate('resources.radius/profiles.sections.binding.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="bind_mac"
                  label={translate('resources.radius/profiles.fields.bind_mac')}
                  helperText={translate('resources.radius/profiles.helpers.bind_mac')}
                />
              </Box>
            </FieldGridItem>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="bind_vlan"
                  label={translate('resources.radius/profiles.fields.bind_vlan')}
                  helperText={translate('resources.radius/profiles.helpers.bind_vlan')}
                />
              </Box>
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.remark.title')}
          description={translate('resources.radius/profiles.sections.remark.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="remark"
                label={translate('resources.radius/profiles.fields.remark')}
                multiline
                minRows={3}
                fullWidth
                size="small"
                helperText={translate('resources.radius/profiles.helpers.remark')}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Edit>
  );
};

// RADIUS 计费策略创建
export const RadiusProfileCreate = () => {
  const translate = useTranslate();
  
  return (
    <Create>
      <SimpleForm sx={formLayoutSx}>
        <FormSection
          title={translate('resources.radius/profiles.sections.basic.title')}
          description={translate('resources.radius/profiles.sections.basic.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="name"
                label={translate('resources.radius/profiles.fields.name')}
                validate={[required(), minLength(2), maxLength(50)]}
                helperText={translate('resources.radius/profiles.helpers.name')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="status"
                  label={translate('resources.radius/profiles.fields.status_enabled')}
                  defaultValue={true}
                  helperText={translate('resources.radius/profiles.helpers.status')}
                />
              </Box>
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.rate_control.title')}
          description={translate('resources.radius/profiles.sections.rate_control.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
            <FieldGridItem>
              <NumberInput
                source="active_num"
                label={translate('resources.radius/profiles.fields.active_num')}
                min={0}
                defaultValue={1}
                helperText={translate('resources.radius/profiles.helpers.active_num')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput
                source="up_rate"
                label={translate('resources.radius/profiles.fields.up_rate')}
                min={0}
                defaultValue={1024}
                helperText={translate('resources.radius/profiles.helpers.up_rate')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput
                source="down_rate"
                label={translate('resources.radius/profiles.fields.down_rate')}
                min={0}
                defaultValue={1024}
                helperText={translate('resources.radius/profiles.helpers.down_rate')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.network.title')}
          description={translate('resources.radius/profiles.sections.network.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="addr_pool"
                label={translate('resources.radius/profiles.fields.addr_pool')}
                helperText={translate('resources.radius/profiles.helpers.addr_pool')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="ipv6_prefix"
                label={translate('resources.radius/profiles.fields.ipv6_prefix')}
                helperText={translate('resources.radius/profiles.helpers.ipv6_prefix')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="domain"
                label={translate('resources.radius/profiles.fields.domain')}
                helperText={translate('resources.radius/profiles.helpers.domain')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.binding.title')}
          description={translate('resources.radius/profiles.sections.binding.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="bind_mac"
                  label={translate('resources.radius/profiles.fields.bind_mac')}
                  defaultValue={false}
                  helperText={translate('resources.radius/profiles.helpers.bind_mac')}
                />
              </Box>
            </FieldGridItem>
            <FieldGridItem>
              <Box sx={controlWrapperSx}>
                <BooleanInput
                  source="bind_vlan"
                  label={translate('resources.radius/profiles.fields.bind_vlan')}
                  defaultValue={false}
                  helperText={translate('resources.radius/profiles.helpers.bind_vlan')}
                />
              </Box>
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.radius/profiles.sections.remark.title')}
          description={translate('resources.radius/profiles.sections.remark.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="remark"
                label={translate('resources.radius/profiles.fields.remark')}
                multiline
                minRows={3}
                fullWidth
                size="small"
                helperText={translate('resources.radius/profiles.helpers.remark')}
              />
          </FieldGridItem>
        </FieldGrid>
      </FormSection>
    </SimpleForm>
  </Create>
);
};

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
  const translate = useTranslate();
  if (!record) return null;

  return (
    <Chip
      label={translate(`resources.radius/profiles.status.${record.status}`)}
      color={record.status === 'enabled' ? 'success' : 'default'}
      size="small"
      sx={{ fontWeight: 500 }}
    />
  );
};

// 布尔值显示组件
const BooleanDisplay = ({ source }: { source: string }) => {
  const record = useRecordContext();
  const translate = useTranslate();
  if (!record) return null;

  const value = record[source];
  return (
    <Chip
      label={translate(value ? 'common.yes' : 'common.no')}
      color={value ? 'success' : 'default'}
      size="small"
      variant="outlined"
    />
  );
};

// RADIUS 计费策略详情
export const RadiusProfileShow = () => {
  const translate = useTranslate();
  const kbpsUnit = translate('resources.radius/profiles.units.kbps');
  const noRemark = translate('resources.radius/profiles.helpers.no_remark');
  
  return (
    <Show actions={<ProfileShowActions />}>
      <Box sx={{ width: '100%', p: { xs: 2, sm: 3, md: 4 } }}>
        <Stack spacing={3}>
          {/* 基本信息卡片 */}
          <Card elevation={2}>
            <CardContent>
              <Typography variant="h6" gutterBottom sx={{ mb: 2, fontWeight: 600, color: 'primary.main' }}>
                {translate('resources.radius/profiles.sections.basic.title')}
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <TableContainer>
                <Table>
                  <TableBody>
                    <DetailRow
                      label={translate('resources.radius/profiles.fields.id')}
                      value={<TextField source="id" />}
                    />
                    <DetailRow
                      label={translate('resources.radius/profiles.fields.name')}
                      value={<TextField source="name" />}
                    />
                    <DetailRow
                      label={translate('resources.radius/profiles.fields.status')}
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
                  {translate('resources.radius/profiles.sections.rate_control.title')}
                </Typography>
                <Divider sx={{ mb: 2 }} />
                <TableContainer>
                  <Table>
                    <TableBody>
                      <DetailRow
                        label={translate('resources.radius/profiles.fields.active_num')}
                        value={<TextField source="active_num" />}
                      />
                      <DetailRow
                        label={translate('resources.radius/profiles.fields.up_rate')}
                        value={
                          <Box>
                            <TextField source="up_rate" />
                            <Typography component="span" variant="body2" color="text.secondary" sx={{ ml: 1 }}>
                              {kbpsUnit}
                            </Typography>
                          </Box>
                        }
                      />
                      <DetailRow
                        label={translate('resources.radius/profiles.fields.down_rate')}
                        value={
                          <Box>
                            <TextField source="down_rate" />
                            <Typography component="span" variant="body2" color="text.secondary" sx={{ ml: 1 }}>
                              {kbpsUnit}
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
                  {translate('resources.radius/profiles.sections.network.title')}
                </Typography>
                <Divider sx={{ mb: 2 }} />
                <TableContainer>
                  <Table>
                    <TableBody>
                      <DetailRow
                        label={translate('resources.radius/profiles.fields.addr_pool')}
                        value={<TextField source="addr_pool" emptyText="-" />}
                      />
                      <DetailRow
                        label={translate('resources.radius/profiles.fields.ipv6_prefix')}
                        value={<TextField source="ipv6_prefix" emptyText="-" />}
                      />
                      <DetailRow
                        label={translate('resources.radius/profiles.fields.domain')}
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
                  {translate('resources.radius/profiles.sections.binding.title')}
                </Typography>
                <Divider sx={{ mb: 2 }} />
                <TableContainer>
                  <Table>
                    <TableBody>
                      <DetailRow
                        label={translate('resources.radius/profiles.fields.bind_mac')}
                        value={<BooleanDisplay source="bind_mac" />}
                      />
                      <DetailRow
                        label={translate('resources.radius/profiles.fields.bind_vlan')}
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
                  {translate('resources.radius/profiles.sections.timestamps.title')}
                </Typography>
                <Divider sx={{ mb: 2 }} />
                <TableContainer>
                  <Table>
                    <TableBody>
                      <DetailRow
                        label={translate('resources.radius/profiles.fields.created_at')}
                        value={<DateField source="created_at" showTime />}
                      />
                      <DetailRow
                        label={translate('resources.radius/profiles.fields.updated_at')}
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
                {translate('resources.radius/profiles.sections.remark.title')}
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <Box sx={{ p: 2, backgroundColor: theme => theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.02)' : 'rgba(0, 0, 0, 0.01)', borderRadius: 1 }}>
                <TextField
                  source="remark"
                  emptyText={noRemark}
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
