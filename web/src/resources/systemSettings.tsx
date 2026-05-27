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
  Create,
  Show,
  SimpleShowLayout,
  FilterButton,
  TopToolbar,
  CreateButton,
  ExportButton,
} from 'react-admin';

// 系统设置列表操作栏
const SettingsListActions = () => (
  <TopToolbar>
    <FilterButton />
    <CreateButton />
    <ExportButton />
  </TopToolbar>
);

// 系统设置列表
export const SystemSettingsList = () => (
  <List actions={<SettingsListActions />} filters={settingsFilters}>
    <Datagrid rowClick="edit">
      <TextField source="id" label="ID" />
      <TextField source="type" label="类型" />
      <TextField source="name" label="配置名称" />
      <TextField source="value" label="配置值" />
      <TextField source="sort" label="排序" />
      <TextField source="remark" label="备注" />
      <DateField source="created_at" label="创建时间" showTime />
      <DateField source="updated_at" label="更新时间" showTime />
    </Datagrid>
  </List>
);

// 筛选器
const settingsFilters = [
  <TextInput label="类型" source="type" alwaysOn />,
  <TextInput label="名称" source="name" />,
];

// 系统设置编辑
export const SystemSettingsEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="id" disabled />
      <SelectInput
        source="type"
        label="类型"
        required
        choices={[
          { id: 'system', name: '系统配置' },
          { id: 'radius', name: 'RADIUS配置' },
          { id: 'security', name: '安全配置' },
          { id: 'network', name: '网络配置' },
          { id: 'email', name: '邮件配置' },
          { id: 'other', name: '其他配置' },
        ]}
      />
      <TextInput source="name" label="配置名称" required />
      <TextInput source="value" label="配置值" required multiline rows={3} />
      <NumberInput source="sort" label="排序" defaultValue={0} />
      <TextInput source="remark" label="备注" multiline rows={2} />
    </SimpleForm>
  </Edit>
);

// 系统设置创建
export const SystemSettingsCreate = () => (
  <Create>
    <SimpleForm>
      <SelectInput
        source="type"
        label="类型"
        required
        defaultValue="system"
        choices={[
          { id: 'system', name: '系统配置' },
          { id: 'radius', name: 'RADIUS配置' },
          { id: 'security', name: '安全配置' },
          { id: 'network', name: '网络配置' },
          { id: 'email', name: '邮件配置' },
          { id: 'other', name: '其他配置' },
        ]}
      />
      <TextInput source="name" label="配置名称" required />
      <TextInput source="value" label="配置值" required multiline rows={3} />
      <NumberInput source="sort" label="排序" defaultValue={0} />
      <TextInput source="remark" label="备注" multiline rows={2} />
    </SimpleForm>
  </Create>
);

// 系统设置详情
export const SystemSettingsShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="id" label="ID" />
      <TextField source="type" label="类型" />
      <TextField source="name" label="配置名称" />
      <TextField source="value" label="配置值" />
      <TextField source="sort" label="排序" />
      <TextField source="remark" label="备注" />
      <DateField source="created_at" label="创建时间" showTime />
      <DateField source="updated_at" label="更新时间" showTime />
    </SimpleShowLayout>
  </Show>
);
