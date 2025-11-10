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
  useTranslate,
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

// 验证规则工厂函数
const useValidationRules = () => {
  const translate = useTranslate();

  return {
    validateUsername: [
      required(translate('resources.system/operators.validation.username_required')),
      minLength(3, translate('resources.system/operators.validation.username_min')),
      maxLength(30, translate('resources.system/operators.validation.username_max')),
      regex(/^[a-zA-Z0-9_]+$/, translate('resources.system/operators.validation.username_format')),
    ],
    validatePassword: [
      required(translate('resources.system/operators.validation.password_required')),
      minLength(6, translate('resources.system/operators.validation.password_min')),
      maxLength(50, translate('resources.system/operators.validation.password_max')),
      regex(/^(?=.*[A-Za-z])(?=.*\d).+$/, translate('resources.system/operators.validation.password_format')),
    ],
    validatePasswordOptional: [
      minLength(6, translate('resources.system/operators.validation.password_min')),
      maxLength(50, translate('resources.system/operators.validation.password_max')),
      regex(/^(?=.*[A-Za-z])(?=.*\d).+$/, translate('resources.system/operators.validation.password_format')),
    ],
    validateEmail: [email(translate('resources.system/operators.validation.email_invalid'))],
    validateMobile: [
      regex(
        /^(0|\+?86)?(13[0-9]|14[57]|15[0-35-9]|17[0678]|18[0-9])[0-9]{8}$/,
        translate('resources.system/operators.validation.mobile_invalid')
      ),
    ],
    validateRealname: [required(translate('resources.system/operators.validation.realname_required'))],
    validateLevel: [required(translate('resources.system/operators.validation.level_required'))],
    validateStatus: [required(translate('resources.system/operators.validation.status_required'))],
  };
};

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
  const translate = useTranslate();
  if (!record) return null;

  return (
    <Chip
      label={translate(`resources.system/operators.status_${record.status}`)}
      color={record.status === 'enabled' ? 'success' : 'default'}
      size="small"
    />
  );
};

// 权限级别显示组件
const LevelField = () => {
  const record = useRecordContext();
  const translate = useTranslate();
  if (!record) return null;

  const levelMap: Record<string, { color: 'error' | 'warning' | 'info' }> = {
    super: { color: 'error' },
    admin: { color: 'warning' },
    operator: { color: 'info' },
  };

  const levelInfo = levelMap[record.level] || { color: 'info' };
  const label = translate(`resources.system/operators.levels.${record.level}`, { _: record.level });

  return <Chip label={label} color={levelInfo.color} size="small" />;
};

// 操作员列表操作栏
const OperatorsListActions = () => (
  <TopToolbar>
    <FilterButton />
    <CreateButton />
    <ExportButton />
  </TopToolbar>
);

// 筛选器
const useOperatorFilters = () => {
  const translate = useTranslate();
  
  return [
    <TextInput 
      key="username" 
      source="username" 
      label={translate('resources.system/operators.fields.username')} 
      alwaysOn 
    />,
    <TextInput 
      key="realname" 
      source="realname" 
      label={translate('resources.system/operators.fields.realname')} 
    />,
    <SelectInput
      key="status"
      source="status"
      label={translate('resources.system/operators.fields.status')}
      choices={[
        { id: 'enabled', name: translate('resources.system/operators.status.enabled') },
        { id: 'disabled', name: translate('resources.system/operators.status.disabled') },
      ]}
    />,
    <SelectInput
      key="level"
      source="level"
      label={translate('resources.system/operators.fields.level')}
      choices={[
        { id: 'super', name: translate('resources.system/operators.levels.super') },
        { id: 'admin', name: translate('resources.system/operators.levels.admin') },
        { id: 'operator', name: translate('resources.system/operators.levels.operator') },
      ]}
    />,
  ];
};

// 操作员列表
export const OperatorList = () => {
  const translate = useTranslate();
  
  return (
    <List actions={<OperatorsListActions />} filters={useOperatorFilters()}>
      <Datagrid rowClick="show">
        <TextField source="id" label={translate('resources.system/operators.fields.id')} />
        <TextField source="username" label={translate('resources.system/operators.fields.username')} />
        <TextField source="realname" label={translate('resources.system/operators.fields.realname')} />
        <EmailField source="email" label={translate('resources.system/operators.fields.email')} />
        <TextField source="mobile" label={translate('resources.system/operators.fields.mobile')} />
        <LevelField />
        <StatusField />
        <DateField source="last_login" label={translate('resources.system/operators.fields.last_login')} showTime />
        <DateField source="created_at" label={translate('resources.system/operators.fields.created_at')} showTime />
      </Datagrid>
    </List>
  );
};

// 操作员编辑
export const OperatorEdit = () => {
  const { identity } = useGetIdentity();
  const record = useRecordContext();
  const translate = useTranslate();
  const validation = useValidationRules();
  
  // 检查是否是编辑自己的账号
  const isEditingSelf = identity && record && String(identity.id) === String(record.id);
  
  // 检查用户权限 - 只有超级管理员和管理员可以看到权限设置
  const canManagePermissions = identity?.level === 'super' || identity?.level === 'admin';
  
  return (
    <Edit>
      <SimpleForm sx={{ maxWidth: 800 }}>
        <FormSection 
          title={translate('resources.system/operators.sections.basic.title')} 
          description={translate('resources.system/operators.sections.basic.description')}
        >
          <FieldGrid>
            <TextInput source="id" label={translate('resources.system/operators.fields.id')} disabled fullWidth />
            <TextInput 
              source="username" 
              label={translate('resources.system/operators.fields.username')} 
              validate={validation.validateUsername}
              helperText={translate('resources.system/operators.helpers.username')}
              fullWidth
            />
            <PasswordInput 
              source="password" 
              label={translate('resources.system/operators.fields.password')} 
              validate={validation.validatePasswordOptional}
              helperText={translate('resources.system/operators.helpers.password_optional')} 
              fullWidth
            />
          </FieldGrid>
        </FormSection>

        <FormSection 
          title={translate('resources.system/operators.sections.personal.title')} 
          description={translate('resources.system/operators.sections.personal.description')}
        >
          <FieldGrid>
            <TextInput 
              source="realname" 
              label={translate('resources.system/operators.fields.realname')} 
              validate={validation.validateRealname}
              fullWidth
            />
            <TextInput 
              source="email" 
              label={translate('resources.system/operators.fields.email')} 
              type="email" 
              validate={validation.validateEmail}
              helperText={translate('resources.system/operators.helpers.email')}
              fullWidth
            />
            <TextInput 
              source="mobile" 
              label={translate('resources.system/operators.fields.mobile')} 
              validate={validation.validateMobile}
              helperText={translate('resources.system/operators.helpers.mobile')}
              fullWidth
            />
          </FieldGrid>
        </FormSection>

        {/* 只有超级管理员和管理员可以看到权限设置 */}
        {canManagePermissions && (
          <FormSection 
            title={translate('resources.system/operators.sections.permissions.title')} 
            description={translate('resources.system/operators.sections.permissions.description')}
          >
            <FieldGrid>
              <SelectInput
                source="level"
                label={translate('resources.system/operators.fields.level')}
                validate={validation.validateLevel}
                disabled={isEditingSelf}
                choices={[
                  { id: 'super', name: translate('resources.system/operators.levels.super') },
                  { id: 'admin', name: translate('resources.system/operators.levels.admin') },
                  { id: 'operator', name: translate('resources.system/operators.levels.operator') },
                ]}
                helperText={isEditingSelf ? translate('resources.system/operators.helpers.cannot_change_own_level') : translate('resources.system/operators.helpers.level')}
                fullWidth
              />
              <SelectInput
                source="status"
                label={translate('resources.system/operators.fields.status')}
                validate={validation.validateStatus}
                disabled={isEditingSelf}
                choices={[
                  { id: 'enabled', name: translate('resources.system/operators.status.enabled') },
                  { id: 'disabled', name: translate('resources.system/operators.status.disabled') },
                ]}
                helperText={isEditingSelf ? translate('resources.system/operators.helpers.cannot_change_own_status') : translate('resources.system/operators.helpers.status')}
                fullWidth
              />
            </FieldGrid>
          </FormSection>
        )}

        <FormSection title={translate('resources.system/operators.sections.remark.title')}>
          <TextInput 
            source="remark" 
            label={translate('resources.system/operators.fields.remark')} 
            multiline 
            rows={3} 
            fullWidth
            helperText={translate('resources.system/operators.helpers.remark')}
          />
        </FormSection>
      </SimpleForm>
    </Edit>
  );
};

// 操作员创建
export const OperatorCreate = () => {
  const translate = useTranslate();
  const validation = useValidationRules();
  
  return (
    <Create>
      <SimpleForm sx={{ maxWidth: 800 }}>
        <FormSection 
          title={translate('resources.system/operators.sections.basic.title')} 
          description={translate('resources.system/operators.sections.basic.description')}
        >
          <FieldGrid>
            <TextInput 
              source="username" 
              label={translate('resources.system/operators.fields.username')} 
              validate={validation.validateUsername}
              helperText={translate('resources.system/operators.helpers.username')}
              fullWidth
            />
            <PasswordInput 
              source="password" 
              label={translate('resources.system/operators.fields.password')} 
              validate={validation.validatePassword}
              helperText={translate('resources.system/operators.helpers.password')}
              fullWidth
            />
          </FieldGrid>
        </FormSection>

        <FormSection 
          title={translate('resources.system/operators.sections.personal.title')} 
          description={translate('resources.system/operators.sections.personal.description')}
        >
          <FieldGrid>
            <TextInput 
              source="realname" 
              label={translate('resources.system/operators.fields.realname')} 
              validate={validation.validateRealname}
              fullWidth
            />
            <TextInput 
              source="email" 
              label={translate('resources.system/operators.fields.email')} 
              type="email" 
              validate={validation.validateEmail}
              helperText={translate('resources.system/operators.helpers.email')}
              fullWidth
            />
            <TextInput 
              source="mobile" 
              label={translate('resources.system/operators.fields.mobile')} 
              validate={validation.validateMobile}
              helperText={translate('resources.system/operators.helpers.mobile')}
              fullWidth
            />
          </FieldGrid>
        </FormSection>

        <FormSection 
          title={translate('resources.system/operators.sections.permissions.title')} 
          description={translate('resources.system/operators.sections.permissions.description')}
        >
          <FieldGrid>
            <SelectInput
              source="level"
              label={translate('resources.system/operators.fields.level')}
              validate={validation.validateLevel}
              defaultValue="operator"
              choices={[
                { id: 'super', name: translate('resources.system/operators.levels.super') },
                { id: 'admin', name: translate('resources.system/operators.levels.admin') },
                { id: 'operator', name: translate('resources.system/operators.levels.operator') },
              ]}
              helperText={translate('resources.system/operators.helpers.level')}
              fullWidth
            />
            <SelectInput
              source="status"
              label={translate('resources.system/operators.fields.status')}
              validate={validation.validateStatus}
              defaultValue="enabled"
              choices={[
                { id: 'enabled', name: translate('resources.system/operators.status.enabled') },
                { id: 'disabled', name: translate('resources.system/operators.status.disabled') },
              ]}
              helperText={translate('resources.system/operators.helpers.status')}
              fullWidth
            />
          </FieldGrid>
        </FormSection>

        <FormSection title={translate('resources.system/operators.sections.remark.title')}>
          <TextInput 
            source="remark" 
            label={translate('resources.system/operators.fields.remark')} 
            multiline 
            rows={3} 
            fullWidth
            helperText={translate('resources.system/operators.helpers.remark')}
          />
        </FormSection>
      </SimpleForm>
    </Create>
  );
};

// 操作员详情
export const OperatorShow = () => {
  const translate = useTranslate();
  
  return (
    <Show>
      <Box sx={{ p: 2 }}>
        <Paper elevation={0} sx={{ p: 3, mb: 3, borderRadius: 2 }}>
          <Typography variant="h6" sx={{ mb: 3, fontWeight: 600 }}>
            {translate('resources.system/operators.sections.basic.title')}
          </Typography>
          <TableContainer>
            <Table>
              <TableBody>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600, width: '30%' }}>{translate('resources.system/operators.fields.id')}</TableCell>
                  <TableCell><TextField source="id" /></TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600 }}>{translate('resources.system/operators.fields.username')}</TableCell>
                  <TableCell><TextField source="username" /></TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600 }}>{translate('resources.system/operators.fields.realname')}</TableCell>
                  <TableCell><TextField source="realname" /></TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600 }}>{translate('resources.system/operators.fields.email')}</TableCell>
                  <TableCell><EmailField source="email" /></TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600 }}>{translate('resources.system/operators.fields.mobile')}</TableCell>
                  <TableCell><TextField source="mobile" /></TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </TableContainer>
        </Paper>

        <Paper elevation={0} sx={{ p: 3, mb: 3, borderRadius: 2 }}>
          <Typography variant="h6" sx={{ mb: 3, fontWeight: 600 }}>
            {translate('resources.system/operators.sections.permissions.title')}
          </Typography>
          <TableContainer>
            <Table>
              <TableBody>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600, width: '30%' }}>{translate('resources.system/operators.fields.level')}</TableCell>
                  <TableCell><LevelField /></TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600 }}>{translate('resources.system/operators.fields.status')}</TableCell>
                  <TableCell><StatusField /></TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </TableContainer>
        </Paper>

        <Paper elevation={0} sx={{ p: 3, mb: 3, borderRadius: 2 }}>
          <Typography variant="h6" sx={{ mb: 3, fontWeight: 600 }}>
            {translate('resources.system/operators.sections.other.title')}
          </Typography>
          <TableContainer>
            <Table>
              <TableBody>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600, width: '30%' }}>{translate('resources.system/operators.fields.remark')}</TableCell>
                  <TableCell><TextField source="remark" /></TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600 }}>{translate('resources.system/operators.fields.last_login')}</TableCell>
                  <TableCell><DateField source="last_login" showTime /></TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600 }}>{translate('resources.system/operators.fields.created_at')}</TableCell>
                  <TableCell><DateField source="created_at" showTime /></TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600 }}>{translate('resources.system/operators.fields.updated_at')}</TableCell>
                  <TableCell><DateField source="updated_at" showTime /></TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </TableContainer>
        </Paper>
      </Box>
    </Show>
  );
};
