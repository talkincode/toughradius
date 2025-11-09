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
  SimpleShowLayout,
  FilterButton,
  TopToolbar,
  CreateButton,
  ExportButton,
} from 'react-admin';

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
    <Datagrid rowClick="edit">
      <TextField source="id" label="ID" />
      <TextField source="username" label="用户名" />
      <TextField source="realname" label="真实姓名" />
      <EmailField source="email" label="邮箱" />
      <TextField source="mobile" label="手机号" />
      <TextField source="level" label="权限级别" />
      <TextField source="status" label="状态" />
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
export const OperatorEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="id" disabled />
      <TextInput source="username" label="用户名" required />
      <PasswordInput 
        source="password" 
        label="密码" 
        helperText="留空则不修改密码" 
      />
      <TextInput source="realname" label="真实姓名" required />
      <TextInput source="email" label="邮箱" type="email" />
      <TextInput source="mobile" label="手机号" />
      <SelectInput
        source="level"
        label="权限级别"
        required
        choices={[
          { id: 'super', name: '超级管理员' },
          { id: 'admin', name: '管理员' },
          { id: 'operator', name: '操作员' },
        ]}
      />
      <SelectInput
        source="status"
        label="状态"
        required
        choices={[
          { id: 'enabled', name: '启用' },
          { id: 'disabled', name: '禁用' },
        ]}
      />
      <TextInput source="remark" label="备注" multiline rows={3} />
    </SimpleForm>
  </Edit>
);

// 操作员创建
export const OperatorCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="username" label="用户名" required />
      <PasswordInput source="password" label="密码" required />
      <TextInput source="realname" label="真实姓名" required />
      <TextInput source="email" label="邮箱" type="email" />
      <TextInput source="mobile" label="手机号" />
      <SelectInput
        source="level"
        label="权限级别"
        required
        defaultValue="operator"
        choices={[
          { id: 'super', name: '超级管理员' },
          { id: 'admin', name: '管理员' },
          { id: 'operator', name: '操作员' },
        ]}
      />
      <SelectInput
        source="status"
        label="状态"
        required
        defaultValue="enabled"
        choices={[
          { id: 'enabled', name: '启用' },
          { id: 'disabled', name: '禁用' },
        ]}
      />
      <TextInput source="remark" label="备注" multiline rows={3} />
    </SimpleForm>
  </Create>
);

// 操作员详情
export const OperatorShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="id" label="ID" />
      <TextField source="username" label="用户名" />
      <TextField source="realname" label="真实姓名" />
      <EmailField source="email" label="邮箱" />
      <TextField source="mobile" label="手机号" />
      <TextField source="level" label="权限级别" />
      <TextField source="status" label="状态" />
      <TextField source="remark" label="备注" />
      <DateField source="last_login" label="最后登录" showTime />
      <DateField source="created_at" label="创建时间" showTime />
      <DateField source="updated_at" label="更新时间" showTime />
    </SimpleShowLayout>
  </Show>
);
