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
  SimpleShowLayout,
  FilterButton,
  TopToolbar,
  CreateButton,
  ExportButton,
  ReferenceField,
  PasswordInput,
} from 'react-admin';

// NAS 列表操作栏
const NASListActions = () => (
  <TopToolbar>
    <FilterButton />
    <CreateButton />
    <ExportButton />
  </TopToolbar>
);

// NAS 设备列表
export const NASList = () => (
  <List actions={<NASListActions />} filters={nasFilters}>
    <Datagrid rowClick="edit">
      <TextField source="id" label="ID" />
      <TextField source="name" label="设备名称" />
      <TextField source="identifier" label="设备标识" />
      <TextField source="ipaddr" label="设备IP" />
      <TextField source="vendor_code" label="厂商代码" />
      <TextField source="model" label="设备型号" />
      <TextField source="status" label="状态" />
      <ReferenceField source="node_id" reference="network/nodes" label="所属节点">
        <TextField source="name" />
      </ReferenceField>
      <DateField source="created_at" label="创建时间" showTime />
    </Datagrid>
  </List>
);

// 筛选器
const nasFilters = [
  <TextInput label="设备名称" source="name" alwaysOn />,
  <TextInput label="设备IP" source="ipaddr" />,
  <SelectInput
    label="状态"
    source="status"
    choices={[
      { id: 'enabled', name: '启用' },
      { id: 'disabled', name: '禁用' },
    ]}
  />,
];

// NAS 设备编辑
export const NASEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="id" disabled />
      <TextInput source="name" label="设备名称" required />
      <TextInput source="identifier" label="设备标识-RADIUS" required />
      <TextInput source="hostname" label="设备主机地址" />
      <TextInput source="ipaddr" label="设备IP" required />
      <PasswordInput source="secret" label="RADIUS秘钥" required />
      <NumberInput source="coa_port" label="COA端口" defaultValue={3799} />
      <TextInput source="model" label="设备型号" />
      <SelectInput
        source="vendor_code"
        label="厂商代码"
        required
        choices={[
          { id: '9', name: 'Cisco' },
          { id: '2011', name: 'Huawei' },
          { id: '14988', name: 'Mikrotik' },
          { id: '25506', name: 'H3C' },
          { id: '3902', name: 'ZTE' },
          { id: '10055', name: 'Ikuai' },
          { id: '0', name: 'Standard' },
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
      <ReferenceInput source="node_id" reference="network/nodes" label="所属节点">
        <SelectInput optionText="name" />
      </ReferenceInput>
      <TextInput source="tags" label="标签" helperText="多个标签用逗号分隔" />
      <TextInput source="remark" label="备注" multiline rows={3} />
    </SimpleForm>
  </Edit>
);

// NAS 设备创建
export const NASCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="name" label="设备名称" required />
      <TextInput source="identifier" label="设备标识-RADIUS" required />
      <TextInput source="hostname" label="设备主机地址" />
      <TextInput source="ipaddr" label="设备IP" required />
      <PasswordInput source="secret" label="RADIUS秘钥" required />
      <NumberInput source="coa_port" label="COA端口" defaultValue={3799} />
      <TextInput source="model" label="设备型号" />
      <SelectInput
        source="vendor_code"
        label="厂商代码"
        required
        defaultValue="0"
        choices={[
          { id: '9', name: 'Cisco' },
          { id: '2011', name: 'Huawei' },
          { id: '14988', name: 'Mikrotik' },
          { id: '25506', name: 'H3C' },
          { id: '3902', name: 'ZTE' },
          { id: '10055', name: 'Ikuai' },
          { id: '0', name: 'Standard' },
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
      <ReferenceInput source="node_id" reference="network/nodes" label="所属节点">
        <SelectInput optionText="name" />
      </ReferenceInput>
      <TextInput source="tags" label="标签" helperText="多个标签用逗号分隔" />
      <TextInput source="remark" label="备注" multiline rows={3} />
    </SimpleForm>
  </Create>
);

// NAS 设备详情
export const NASShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="id" label="ID" />
      <TextField source="name" label="设备名称" />
      <TextField source="identifier" label="设备标识" />
      <TextField source="hostname" label="设备主机地址" />
      <TextField source="ipaddr" label="设备IP" />
      <TextField source="coa_port" label="COA端口" />
      <TextField source="model" label="设备型号" />
      <TextField source="vendor_code" label="厂商代码" />
      <TextField source="status" label="状态" />
      <ReferenceField source="node_id" reference="network/nodes" label="所属节点">
        <TextField source="name" />
      </ReferenceField>
      <TextField source="tags" label="标签" />
      <TextField source="remark" label="备注" />
      <DateField source="created_at" label="创建时间" showTime />
      <DateField source="updated_at" label="更新时间" showTime />
    </SimpleShowLayout>
  </Show>
);
