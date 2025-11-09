import {
  List,
  Datagrid,
  TextField,
  DateField,
  Edit,
  TextInput,
  Create,
  Show,
  SimpleShowLayout,
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
  ToolbarProps
} from 'react-admin';
import {
  Box,
  Chip,
  Typography,
  Paper
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

// RADIUS 计费策略列表
export const RadiusProfileList = () => (
  <List>
    <Datagrid rowClick="edit">
      <TextField source="id" label="ID" />
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

// RADIUS 计费策略详情
export const RadiusProfileShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="id" label="ID" />
      <TextField source="name" label="策略名称" />
      <StatusField />
      <TextField source="active_num" label="并发数" />
      <TextField source="up_rate" label="上行速率(Kbps)" />
      <TextField source="down_rate" label="下行速率(Kbps)" />
      <TextField source="addr_pool" label="地址池" />
      <TextField source="ipv6_prefix" label="IPv6前缀" />
      <TextField source="domain" label="域" />
      <TextField source="bind_mac" label="绑定MAC" />
      <TextField source="bind_vlan" label="绑定VLAN" />
      <TextField source="remark" label="备注" />
      <DateField source="created_at" label="创建时间" showTime />
      <DateField source="updated_at" label="更新时间" showTime />
    </SimpleShowLayout>
  </Show>
);
