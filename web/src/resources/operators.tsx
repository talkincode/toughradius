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
  FilterButton,
  TopToolbar,
  CreateButton,
  ExportButton,
  required,
  minLength,
  maxLength,
  email,
  regex,
  useRecordContext,
  useGetIdentity,
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
} from '@mui/material';
import { ReactNode } from 'react';

// 验证规则
const validateUsername = [
  required('用户名不能为空'),
  minLength(3, '用户名长度至少3个字符'),
  maxLength(30, '用户名长度最多30个字符'),
  regex(/^[a-zA-Z0-9_]+$/, '用户名只能包含字母、数字和下划线'),
];

const validatePassword = [
  required('密码不能为空'),
  minLength(6, '密码长度至少6个字符'),
  maxLength(50, '密码长度最多50个字符'),
  regex(/^(?=.*[A-Za-z])(?=.*\d).+$/, '密码必须包含字母和数字'),
];

const validatePasswordOptional = [
  minLength(6, '密码长度至少6个字符'),
  maxLength(50, '密码长度最多50个字符'),
  regex(/^(?=.*[A-Za-z])(?=.*\d).+$/, '密码必须包含字母和数字'),
];

const validateEmail = [email('请输入有效的邮箱地址')];

const validateMobile = [
  regex(
    /^(0|\+?86)?(13[0-9]|14[57]|15[0-35-9]|17[0678]|18[0-9])[0-9]{8}$/,
    '请输入有效的中国手机号'
  ),
];

const validateRealname = [required('真实姓名不能为空')];

// 表单分区组件
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
    <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1 }}>
      {title}
    </Typography>
    {description && (
      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
        {description}
      </Typography>
    )}
    <Box sx={{ width: '100%' }}>
      {children}
    </Box>
  </Paper>
);

// 字段网格布局
interface FieldGridProps {
  children: ReactNode;
  columns?: number;
}

const FieldGrid = ({ children, columns = 2 }: FieldGridProps) => (
  <Box
    sx={{
      display: 'grid',
      gap: 2,
      gridTemplateColumns: {
        xs: '1fr',
        sm: `repeat(${columns}, 1fr)`,
      },
    }}
  >
    {children}
  </Box>
);

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

// 权限级别显示组件
const LevelField = () => {
  const record = useRecordContext();
  if (!record) return null;

  const levelMap: Record<string, { label: string; color: 'error' | 'warning' | 'info' }> = {
    super: { label: '超级管理员', color: 'error' },
    admin: { label: '管理员', color: 'warning' },
    operator: { label: '操作员', color: 'info' },
  };

  const levelInfo = levelMap[record.level] || { label: record.level, color: 'info' };

  return <Chip label={levelInfo.label} color={levelInfo.color} size="small" />;
};

// 操作员列表操作栏
const OperatorsListActions = () => (
  <TopToolbar>
    <FilterButton />
    <CreateButton />
    <ExportButton />
  </TopToolbar>
);

// 操作员列表
export const OperatorList = () => (
  <List actions={<OperatorsListActions />} filters={operatorFilters}>
    <Datagrid rowClick="show">
      <TextField source="id" label="ID" />
      <TextField source="username" label="用户名" />
      <TextField source="realname" label="真实姓名" />
      <EmailField source="email" label="邮箱" />
      <TextField source="mobile" label="手机号" />
      <LevelField />
      <StatusField />
      <DateField source="last_login" label="最后登录" showTime />
      <DateField source="created_at" label="创建时间" showTime />
    </Datagrid>
  </List>
);

// 筛选器
const operatorFilters = [
  <TextInput label="用户名" source="username" alwaysOn />,
  <TextInput label="真实姓名" source="realname" />,
  <SelectInput
    label="状态"
    source="status"
    choices={[
      { id: 'enabled', name: '启用' },
      { id: 'disabled', name: '禁用' },
    ]}
  />,
  <SelectInput
    label="权限级别"
    source="level"
    choices={[
      { id: 'super', name: '超级管理员' },
      { id: 'admin', name: '管理员' },
      { id: 'operator', name: '操作员' },
    ]}
  />,
];

// 操作员编辑
export const OperatorEdit = () => {
  const { identity } = useGetIdentity();
  const record = useRecordContext();
  
  // 检查是否是编辑自己的账号
  const isEditingSelf = identity && record && String(identity.id) === String(record.id);
  
  return (
    <Edit>
      <SimpleForm sx={{ maxWidth: 800 }}>
        <FormSection title="基本信息" description="操作员的基本登录信息">
          <FieldGrid>
            <TextInput source="id" label="ID" disabled fullWidth />
            <TextInput 
              source="username" 
              label="用户名" 
              validate={validateUsername}
              helperText="3-30个字符，只能包含字母、数字和下划线"
              fullWidth
            />
            <PasswordInput 
              source="password" 
              label="密码" 
              validate={validatePasswordOptional}
              helperText="留空则不修改密码，至少6个字符，必须包含字母和数字" 
              fullWidth
            />
          </FieldGrid>
        </FormSection>

        <FormSection title="个人信息" description="操作员的个人联系信息">
          <FieldGrid>
            <TextInput 
              source="realname" 
              label="真实姓名" 
              validate={validateRealname}
              fullWidth
            />
            <TextInput 
              source="email" 
              label="邮箱" 
              type="email" 
              validate={validateEmail}
              helperText="请输入有效的邮箱地址"
              fullWidth
            />
            <TextInput 
              source="mobile" 
              label="手机号" 
              validate={validateMobile}
              helperText="请输入11位中国手机号"
              fullWidth
            />
          </FieldGrid>
        </FormSection>

        <FormSection title="权限设置" description="操作员的权限级别和状态">
          <FieldGrid>
            <SelectInput
              source="level"
              label="权限级别"
              validate={required('请选择权限级别')}
              disabled={isEditingSelf}
              choices={[
                { id: 'super', name: '超级管理员' },
                { id: 'admin', name: '管理员' },
                { id: 'operator', name: '操作员' },
              ]}
              helperText={isEditingSelf ? "不能修改自己的权限级别" : "超级管理员拥有所有权限"}
              fullWidth
            />
            <SelectInput
              source="status"
              label="状态"
              validate={required('请选择状态')}
              disabled={isEditingSelf}
              choices={[
                { id: 'enabled', name: '启用' },
                { id: 'disabled', name: '禁用' },
              ]}
              helperText={isEditingSelf ? "不能修改自己的状态" : "禁用后将无法登录"}
              fullWidth
            />
          </FieldGrid>
        </FormSection>

        <FormSection title="备注信息">
          <TextInput 
            source="remark" 
            label="备注" 
            multiline 
            rows={3} 
            fullWidth
            helperText="可选，记录操作员的其他信息"
          />
        </FormSection>
      </SimpleForm>
    </Edit>
  );
};

// 操作员创建
export const OperatorCreate = () => (
  <Create>
    <SimpleForm sx={{ maxWidth: 800 }}>
      <FormSection title="基本信息" description="操作员的基本登录信息">
        <FieldGrid>
          <TextInput 
            source="username" 
            label="用户名" 
            validate={validateUsername}
            helperText="3-30个字符，只能包含字母、数字和下划线"
            fullWidth
          />
          <PasswordInput 
            source="password" 
            label="密码" 
            validate={validatePassword}
            helperText="至少6个字符，必须包含字母和数字"
            fullWidth
          />
        </FieldGrid>
      </FormSection>

      <FormSection title="个人信息" description="操作员的个人联系信息">
        <FieldGrid>
          <TextInput 
            source="realname" 
            label="真实姓名" 
            validate={validateRealname}
            fullWidth
          />
          <TextInput 
            source="email" 
            label="邮箱" 
            type="email" 
            validate={validateEmail}
            helperText="请输入有效的邮箱地址"
            fullWidth
          />
          <TextInput 
            source="mobile" 
            label="手机号" 
            validate={validateMobile}
            helperText="请输入11位中国手机号"
            fullWidth
          />
        </FieldGrid>
      </FormSection>

      <FormSection title="权限设置" description="操作员的权限级别和状态">
        <FieldGrid>
          <SelectInput
            source="level"
            label="权限级别"
            validate={required('请选择权限级别')}
            defaultValue="operator"
            choices={[
              { id: 'super', name: '超级管理员' },
              { id: 'admin', name: '管理员' },
              { id: 'operator', name: '操作员' },
            ]}
            helperText="超级管理员拥有所有权限"
            fullWidth
          />
          <SelectInput
            source="status"
            label="状态"
            validate={required('请选择状态')}
            defaultValue="enabled"
            choices={[
              { id: 'enabled', name: '启用' },
              { id: 'disabled', name: '禁用' },
            ]}
            helperText="禁用后将无法登录"
            fullWidth
          />
        </FieldGrid>
      </FormSection>

      <FormSection title="备注信息">
        <TextInput 
          source="remark" 
          label="备注" 
          multiline 
          rows={3} 
          fullWidth
          helperText="可选，记录操作员的其他信息"
        />
      </FormSection>
    </SimpleForm>
  </Create>
);

// 操作员详情
export const OperatorShow = () => (
  <Show>
    <Box sx={{ p: 2 }}>
      <Paper elevation={0} sx={{ p: 3, mb: 3, borderRadius: 2 }}>
        <Typography variant="h6" sx={{ mb: 3, fontWeight: 600 }}>
          基本信息
        </Typography>
        <TableContainer>
          <Table>
            <TableBody>
              <TableRow>
                <TableCell sx={{ fontWeight: 600, width: '30%' }}>ID</TableCell>
                <TableCell><TextField source="id" /></TableCell>
              </TableRow>
              <TableRow>
                <TableCell sx={{ fontWeight: 600 }}>用户名</TableCell>
                <TableCell><TextField source="username" /></TableCell>
              </TableRow>
              <TableRow>
                <TableCell sx={{ fontWeight: 600 }}>真实姓名</TableCell>
                <TableCell><TextField source="realname" /></TableCell>
              </TableRow>
              <TableRow>
                <TableCell sx={{ fontWeight: 600 }}>邮箱</TableCell>
                <TableCell><EmailField source="email" /></TableCell>
              </TableRow>
              <TableRow>
                <TableCell sx={{ fontWeight: 600 }}>手机号</TableCell>
                <TableCell><TextField source="mobile" /></TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </TableContainer>
      </Paper>

      <Paper elevation={0} sx={{ p: 3, mb: 3, borderRadius: 2 }}>
        <Typography variant="h6" sx={{ mb: 3, fontWeight: 600 }}>
          权限信息
        </Typography>
        <TableContainer>
          <Table>
            <TableBody>
              <TableRow>
                <TableCell sx={{ fontWeight: 600, width: '30%' }}>权限级别</TableCell>
                <TableCell><LevelField /></TableCell>
              </TableRow>
              <TableRow>
                <TableCell sx={{ fontWeight: 600 }}>状态</TableCell>
                <TableCell><StatusField /></TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </TableContainer>
      </Paper>

      <Paper elevation={0} sx={{ p: 3, mb: 3, borderRadius: 2 }}>
        <Typography variant="h6" sx={{ mb: 3, fontWeight: 600 }}>
          其他信息
        </Typography>
        <TableContainer>
          <Table>
            <TableBody>
              <TableRow>
                <TableCell sx={{ fontWeight: 600, width: '30%' }}>备注</TableCell>
                <TableCell><TextField source="remark" /></TableCell>
              </TableRow>
              <TableRow>
                <TableCell sx={{ fontWeight: 600 }}>最后登录</TableCell>
                <TableCell><DateField source="last_login" showTime /></TableCell>
              </TableRow>
              <TableRow>
                <TableCell sx={{ fontWeight: 600 }}>创建时间</TableCell>
                <TableCell><DateField source="created_at" showTime /></TableCell>
              </TableRow>
              <TableRow>
                <TableCell sx={{ fontWeight: 600 }}>更新时间</TableCell>
                <TableCell><DateField source="updated_at" showTime /></TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </TableContainer>
      </Paper>
    </Box>
  </Show>
);
