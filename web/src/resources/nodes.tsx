import {
  List,
  Datagrid,
  TextField,
  DateField,
  Edit,
  SimpleForm,
  TextInput,
  Create,
  Show,
  SimpleShowLayout,
  FilterButton,
  TopToolbar,
  CreateButton,
  ExportButton,
} from 'react-admin';

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
    <Datagrid rowClick="edit">
      <TextField source="id" label="ID" />
      <TextField source="name" label="节点名称" />
      <TextField source="tags" label="标签" />
      <TextField source="remark" label="备注" />
      <DateField source="created_at" label="创建时间" showTime />
      <DateField source="updated_at" label="更新时间" showTime />
    </Datagrid>
  </List>
);

// 筛选器
const nodeFilters = [
  <TextInput label="节点名称" source="name" alwaysOn />,
];

// 网络节点编辑
export const NodeEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="id" disabled />
      <TextInput source="name" label="节点名称" required />
      <TextInput source="tags" label="标签" helperText="多个标签用逗号分隔" />
      <TextInput source="remark" label="备注" multiline rows={3} />
    </SimpleForm>
  </Edit>
);

// 网络节点创建
export const NodeCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="name" label="节点名称" required />
      <TextInput source="tags" label="标签" helperText="多个标签用逗号分隔" />
      <TextInput source="remark" label="备注" multiline rows={3} />
    </SimpleForm>
  </Create>
);

// 网络节点详情
export const NodeShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="id" label="ID" />
      <TextField source="name" label="节点名称" />
      <TextField source="tags" label="标签" />
      <TextField source="remark" label="备注" />
      <DateField source="created_at" label="创建时间" showTime />
      <DateField source="updated_at" label="更新时间" showTime />
    </SimpleShowLayout>
  </Show>
);
