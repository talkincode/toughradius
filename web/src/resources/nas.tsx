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
  useTranslate
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
const useNasFilters = () => {
  const translate = useTranslate();
  return [
    <TextInput key="name" label={translate('resources.network/nas.fields.name')} source="name" alwaysOn />,
    <TextInput key="ipaddr" label={translate('resources.network/nas.fields.ip_addr')} source="ipaddr" />,
    <SelectInput
      key="status"
      label={translate('resources.network/nas.fields.status')}
      source="status"
      choices={STATUS_CHOICES}
    />,
  ];
};

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
  const translate = useTranslate();
  if (!record) return null;
  
  return (
    <Chip
      label={translate(`resources.network/nas.status.${record.status}`)}
      color={record.status === 'enabled' ? 'success' : 'default'}
      size="small"
      variant="outlined"
    />
  );
};

// NAS 设备列表
export const NASList = () => {
  const translate = useTranslate();
  
  return (
    <List actions={<NASListActions />} filters={useNasFilters()}>
      <Datagrid rowClick="show">
        <TextField source="name" label={translate('resources.network/nas.fields.name')} />
        <TextField source="ipaddr" label={translate('resources.network/nas.fields.ipaddr')} />
        <TextField source="identifier" label={translate('resources.network/nas.fields.identifier')} />
        <TextField source="model" label={translate('resources.network/nas.fields.model')} />
        <StatusChipField />
        <ReferenceField source="node_id" reference="network/nodes" label={translate('resources.network/nas.fields.node_id')} link="show">
          <TextField source="name" />
        </ReferenceField>
        <DateField source="created_at" label={translate('resources.network/nas.fields.created_at')} showTime />
      </Datagrid>
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
          title={translate('resources.network/nas.sections.basic.title')}
          description={translate('resources.network/nas.sections.basic.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
            <FieldGridItem>
              <TextInput
                source="id"
                disabled
                label={translate('resources.network/nas.fields.id')}
                helperText={translate('resources.network/nas.helpers.id')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="name"
                label={translate('resources.network/nas.fields.name')}
                validate={[required(), minLength(1), maxLength(100)]}
                helperText={translate('resources.network/nas.helpers.name')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="identifier"
                label={translate('resources.network/nas.fields.identifier')}
                validate={[required(), minLength(1), maxLength(100)]}
                helperText={translate('resources.network/nas.helpers.identifier')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput
                source="vendor_code"
                label={translate('resources.network/nas.fields.vendor_code')}
                validate={[required()]}
                choices={VENDOR_CHOICES}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="model"
                label={translate('resources.network/nas.fields.model')}
                validate={[maxLength(100)]}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput
                source="status"
                label={translate('resources.network/nas.fields.status')}
                validate={[required()]}
                choices={STATUS_CHOICES}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.network/nas.sections.network.title')}
          description={translate('resources.network/nas.sections.network.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
            <FieldGridItem>
              <TextInput
                source="ipaddr"
                label={translate('resources.network/nas.fields.ipaddr')}
                validate={[required()]}
                helperText={translate('resources.network/nas.helpers.ipaddr')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="hostname"
                label={translate('resources.network/nas.fields.hostname')}
                validate={[maxLength(200)]}
                helperText={translate('resources.network/nas.helpers.hostname')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput
                source="coa_port"
                label={translate('resources.network/nas.fields.coa_port')}
                validate={[number(), minValue(1), maxValue(65535)]}
                helperText={translate('resources.network/nas.helpers.coa_port')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.network/nas.sections.radius.title')}
          description={translate('resources.network/nas.sections.radius.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <PasswordInput
                source="secret"
                label={translate('resources.network/nas.fields.secret')}
                validate={[required(), minLength(6)]}
                helperText={translate('resources.network/nas.helpers.secret')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <ReferenceInput source="node_id" reference="network/nodes" label={translate('resources.network/nas.fields.node_id')}>
                <SelectInput optionText="name" fullWidth size="small" />
              </ReferenceInput>
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="tags"
                label={translate('resources.network/nas.fields.tags')}
                validate={[maxLength(200)]}
                helperText={translate('resources.network/nas.helpers.tags')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.network/nas.sections.remark.title')}
          description={translate('resources.network/nas.sections.remark.description')}
        >
          <FieldGrid columns={{ xs: 1 }}>
            <FieldGridItem>
              <TextInput
                source="remark"
                label={translate('resources.network/nas.fields.remark')}
                validate={[maxLength(500)]}
                multiline
                minRows={3}
                fullWidth
                size="small"
                helperText={translate('resources.network/nas.helpers.remark')}
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
          title={translate('resources.network/nas.sections.basic.title')}
          description={translate('resources.network/nas.sections.basic.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
            <FieldGridItem>
              <TextInput
                source="name"
                label={translate('resources.network/nas.fields.name')}
                validate={[required(), minLength(1), maxLength(100)]}
                helperText={translate('resources.network/nas.helpers.name')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="identifier"
                label={translate('resources.network/nas.fields.identifier')}
                validate={[required(), minLength(1), maxLength(100)]}
                helperText={translate('resources.network/nas.helpers.identifier')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput
                source="vendor_code"
                label={translate('resources.network/nas.fields.vendor_code')}
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
                label={translate('resources.network/nas.fields.model')}
                validate={[maxLength(100)]}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput
                source="status"
                label={translate('resources.network/nas.fields.status')}
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
          title={translate('resources.network/nas.sections.network.title')}
          description={translate('resources.network/nas.sections.network.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2, md: 3 }}>
            <FieldGridItem>
              <TextInput
                source="ipaddr"
                label={translate('resources.network/nas.fields.ipaddr')}
                validate={[required()]}
                helperText={translate('resources.network/nas.helpers.ipaddr')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="hostname"
                label={translate('resources.network/nas.fields.hostname')}
                validate={[maxLength(200)]}
                helperText={translate('resources.network/nas.helpers.hostname')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <NumberInput
                source="coa_port"
                label={translate('resources.network/nas.fields.coa_port')}
                validate={[number(), minValue(1), maxValue(65535)]}
                helperText={translate('resources.network/nas.helpers.coa_port')}
                defaultValue={3799}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.network/nas.sections.radius.title')}
          description={translate('resources.network/nas.sections.radius.description')}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <PasswordInput
                source="secret"
                label={translate('resources.network/nas.fields.secret')}
                validate={[required(), minLength(6)]}
                helperText={translate('resources.network/nas.helpers.secret')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <ReferenceInput source="node_id" reference="network/nodes" label={translate('resources.network/nas.fields.node_id')}>
                <SelectInput optionText="name" fullWidth size="small" />
              </ReferenceInput>
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="tags"
                label={translate('resources.network/nas.fields.tags')}
                validate={[maxLength(200)]}
                helperText={translate('resources.network/nas.helpers.tags')}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.network/nas.sections.remark.title')}
          description={translate('resources.network/nas.sections.remark.description')}
        >
          <FieldGrid columns={{ xs: 1 }}>
            <FieldGridItem>
              <TextInput
                source="remark"
                label={translate('resources.network/nas.fields.remark')}
                validate={[maxLength(500)]}
                multiline
                minRows={3}
                fullWidth
                size="small"
                helperText={translate('resources.network/nas.helpers.remark')}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Create>
  );
};

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
  const translate = useTranslate();
  
  return (
    <Show actions={<NASShowActions />}>
      <Box sx={{ width: '100%', p: { xs: 2, sm: 3, md: 4 } }}>
        <Stack spacing={3}>
          {/* 基本信息卡片 */}
          <Card elevation={2}>
            <CardContent>
              <Typography variant="h6" gutterBottom sx={{ mb: 2, fontWeight: 600, color: 'primary.main' }}>
                {translate('resources.network/nas.sections.basic.title')}
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <TableContainer>
                <Table>
                  <TableBody>
                    <DetailRow
                      label={translate('resources.network/nas.fields.id')}
                      value={<TextField source="id" />}
                    />
                    <DetailRow
                      label={translate('resources.network/nas.fields.name')}
                      value={<TextField source="name" />}
                    />
                    <DetailRow
                      label={translate('resources.network/nas.fields.identifier')}
                      value={<TextField source="identifier" />}
                    />
                    <DetailRow
                      label={translate('resources.network/nas.fields.vendor_code')}
                      value={<VendorCodeField />}
                    />
                    <DetailRow
                      label={translate('resources.network/nas.fields.model')}
                      value={<TextField source="model" emptyText="-" />}
                    />
                    <DetailRow
                      label={translate('resources.network/nas.fields.status')}
                      value={<StatusChipField />}
                    />
                    <DetailRow
                      label={translate('resources.network/nas.fields.node_id')}
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
                {translate('resources.network/nas.sections.network.title')}
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <TableContainer>
                <Table>
                  <TableBody>
                    <DetailRow
                      label={translate('resources.network/nas.fields.ipaddr')}
                      value={<TextField source="ipaddr" />}
                    />
                    <DetailRow
                      label={translate('resources.network/nas.fields.hostname')}
                      value={<TextField source="hostname" emptyText="-" />}
                    />
                    <DetailRow
                      label={translate('resources.network/nas.fields.coa_port')}
                      value={<TextField source="coa_port" />}
                    />
                    <DetailRow
                      label={translate('resources.network/nas.fields.tags')}
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
                {translate('resources.network/nas.sections.timestamps.title')}
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <TableContainer>
                <Table>
                  <TableBody>
                    <DetailRow
                      label={translate('resources.network/nas.fields.created_at')}
                      value={<DateField source="created_at" showTime />}
                    />
                    <DetailRow
                      label={translate('resources.network/nas.fields.updated_at')}
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
                {translate('resources.network/nas.sections.remark.title')}
              </Typography>
              <Divider sx={{ mb: 2 }} />
              <Box sx={{ p: 2, backgroundColor: theme => theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.02)' : 'rgba(0, 0, 0, 0.01)', borderRadius: 1 }}>
                <TextField
                  source="remark"
                  emptyText={translate('resources.network/nas.helpers.no_remark')}
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
