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
  Create,
  Show,
  SimpleShowLayout,
  BooleanField,
  BooleanInput,
} from 'react-admin';

// RADIUS 用户列表
export const RadiusUserList = () => (
  <List>
    <Datagrid rowClick="edit">
      <TextField source="id" label="ID" />
      <TextField source="username" label="用户名" />
      <TextField source="realname" label="真实姓名" />
      <EmailField source="email" label="邮箱" />
      <TextField source="mobile" label="手机号" />
      <TextField source="address" label="地址" />
      <BooleanField source="status" label="状态" />
      <TextField source="profile" label="配置文件" />
      <DateField source="created_at" label="创建时间" showTime />
      <DateField source="expire_time" label="过期时间" showTime />
    </Datagrid>
  </List>
);

// RADIUS 用户编辑
export const RadiusUserEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="id" disabled />
      <TextInput source="username" label="用户名" required />
      <TextInput source="password" label="密码" type="password" />
      <TextInput source="realname" label="真实姓名" />
      <TextInput source="email" label="邮箱" type="email" />
      <TextInput source="mobile" label="手机号" />
      <TextInput source="address" label="地址" multiline />
      <BooleanInput source="status" label="启用状态" />
      <SelectInput
        source="profile"
        label="配置文件"
        choices={[
          { id: 'default', name: '默认配置' },
          { id: 'premium', name: '高级配置' },
          { id: 'business', name: '企业配置' },
        ]}
      />
      <TextInput source="expire_time" label="过期时间" type="datetime-local" />
      <TextInput source="remark" label="备注" multiline />
    </SimpleForm>
  </Edit>
);

// RADIUS 用户创建
export const RadiusUserCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="username" label="用户名" required />
      <TextInput source="password" label="密码" type="password" required />
      <TextInput source="realname" label="真实姓名" />
      <TextInput source="email" label="邮箱" type="email" />
      <TextInput source="mobile" label="手机号" />
      <TextInput source="address" label="地址" multiline />
      <BooleanInput source="status" label="启用状态" defaultValue={true} />
      <SelectInput
        source="profile"
        label="配置文件"
        defaultValue="default"
        choices={[
          { id: 'default', name: '默认配置' },
          { id: 'premium', name: '高级配置' },
          { id: 'business', name: '企业配置' },
        ]}
      />
      <TextInput source="expire_time" label="过期时间" type="datetime-local" />
      <TextInput source="remark" label="备注" multiline />
    </SimpleForm>
  </Create>
);

// RADIUS 用户详情
export const RadiusUserShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="id" label="ID" />
      <TextField source="username" label="用户名" />
      <TextField source="realname" label="真实姓名" />
      <EmailField source="email" label="邮箱" />
      <TextField source="mobile" label="手机号" />
      <TextField source="address" label="地址" />
      <BooleanField source="status" label="状态" />
      <TextField source="profile" label="配置文件" />
      <DateField source="created_at" label="创建时间" showTime />
      <DateField source="expire_time" label="过期时间" showTime />
      <TextField source="remark" label="备注" />
    </SimpleShowLayout>
  </Show>
);
