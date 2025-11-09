import {
  List,
  Datagrid,
  TextField,
  DateField,
  Edit,
  TextInput,
  Create,
  Show,
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
  Chip
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

const NodeFormToolbar = (props: ToolbarProps) => (
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

// ============ 列表相关 ============

// 筛选器
const nodeFilters = [
  <TextInput key="name" label="节点名称" source="name" alwaysOn />,
];

// 网络节点列表操作栏
const NodesListActions = () => (
  <TopToolbar>
    <FilterButton />
    <CreateButton />
    <ExportButton />
  </TopToolbar>
);

// 网络节点列表
export const NodeList = () => (
  <List actions={<NodesListActions />} filters={nodeFilters}>
    <Datagrid rowClick="show">
      <TextField source="name" label="节点名称" />
      <TextField source="tags" label="标签" />
      <TextField source="remark" label="备注" />
      <DateField source="created_at" label="创建时间" showTime />
      <DateField source="updated_at" label="更新时间" showTime />
    </Datagrid>
  </List>
);

// ============ 编辑页面 ============

export const NodeEdit = () => {
  return (
    <Edit>
      <SimpleForm toolbar={<NodeFormToolbar />} sx={formLayoutSx}>
        <FormSection
          title="基本信息"
          description="节点的基本配置信息"
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="id"
                disabled
                label="节点ID"
                helperText="系统自动生成的唯一标识"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="name"
                label="节点名称"
                validate={[required(), minLength(1), maxLength(100)]}
                helperText="1-100个字符"
                fullWidth
                size="small"
              />
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

export const NodeCreate = () => (
  <Create>
    <SimpleForm sx={formLayoutSx}>
      <FormSection
        title="基本信息"
        description="节点的基本配置信息"
      >
        <FieldGrid columns={{ xs: 1, sm: 2 }}>
          <FieldGridItem>
            <TextInput
              source="name"
              label="节点名称"
              validate={[required(), minLength(1), maxLength(100)]}
              helperText="1-100个字符"
              fullWidth
              size="small"
            />
          </FieldGridItem>
          <FieldGridItem>
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
const NodeShowActions = () => (
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

export const NodeShow = () => {
  return (
    <Show actions={<NodeShowActions />}>
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
                      label="节点ID"
                      value={<TextField source="id" />}
                    />
                    <DetailRow
                      label="节点名称"
                      value={<TextField source="name" />}
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
