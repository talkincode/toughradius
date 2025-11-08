import {
  List,
  Datagrid,
  TextField,
  BooleanField,
  DateField,
  Edit,
  SimpleForm,
  TextInput,
  BooleanInput,
  Create,
  Show,
  SimpleShowLayout,
  SelectInput,
  NumberInput,
} from 'react-admin';

// RADIUS 配置列表
export const RadiusProfileList = () => (
  <List>
    <Datagrid rowClick="edit">
      <TextField source="id" label="ID" />
      <TextField source="profile_name" label="配置名称" />
      <TextField source="product_policy" label="产品策略" />
      <TextField source="concur_number" label="并发数" />
      <TextField source="bandwidth_up" label="上传带宽(Kbps)" />
      <TextField source="bandwidth_down" label="下载带宽(Kbps)" />
      <TextField source="ip_pool" label="IP池" />
      <BooleanField source="bind_mac" label="绑定MAC" />
      <BooleanField source="bind_vlan" label="绑定VLAN" />
      <BooleanField source="status" label="状态" />
      <DateField source="created_at" label="创建时间" showTime />
    </Datagrid>
  </List>
);

// RADIUS 配置编辑
export const RadiusProfileEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="id" disabled />
      <TextInput source="profile_name" label="配置名称" required />
      <SelectInput
        source="product_policy"
        label="产品策略"
        choices={[
          { id: 'PrePaid', name: '预付费' },
          { id: 'PostPaid', name: '后付费' },
          { id: 'Free', name: '免费' },
        ]}
      />
      <NumberInput source="concur_number" label="并发数" min={1} defaultValue={1} />
      <NumberInput source="bandwidth_up" label="上传带宽(Kbps)" min={0} />
      <NumberInput source="bandwidth_down" label="下载带宽(Kbps)" min={0} />
      <TextInput source="ip_pool" label="IP池" />
      <BooleanInput source="bind_mac" label="绑定MAC" />
      <BooleanInput source="bind_vlan" label="绑定VLAN" />
      <BooleanInput source="status" label="启用状态" />
      <TextInput source="input_rate_limit" label="上传速率限制" />
      <TextInput source="output_rate_limit" label="下载速率限制" />
      <TextInput source="domain_name" label="域名" />
      <TextInput source="remark" label="备注" multiline />
    </SimpleForm>
  </Edit>
);

// RADIUS 配置创建
export const RadiusProfileCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="profile_name" label="配置名称" required />
      <SelectInput
        source="product_policy"
        label="产品策略"
        defaultValue="PrePaid"
        choices={[
          { id: 'PrePaid', name: '预付费' },
          { id: 'PostPaid', name: '后付费' },
          { id: 'Free', name: '免费' },
        ]}
      />
      <NumberInput source="concur_number" label="并发数" min={1} defaultValue={1} />
      <NumberInput source="bandwidth_up" label="上传带宽(Kbps)" min={0} defaultValue={1024} />
      <NumberInput source="bandwidth_down" label="下载带宽(Kbps)" min={0} defaultValue={1024} />
      <TextInput source="ip_pool" label="IP池" />
      <BooleanInput source="bind_mac" label="绑定MAC" defaultValue={false} />
      <BooleanInput source="bind_vlan" label="绑定VLAN" defaultValue={false} />
      <BooleanInput source="status" label="启用状态" defaultValue={true} />
      <TextInput source="input_rate_limit" label="上传速率限制" />
      <TextInput source="output_rate_limit" label="下载速率限制" />
      <TextInput source="domain_name" label="域名" />
      <TextInput source="remark" label="备注" multiline />
    </SimpleForm>
  </Create>
);

// RADIUS 配置详情
export const RadiusProfileShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="id" label="ID" />
      <TextField source="profile_name" label="配置名称" />
      <TextField source="product_policy" label="产品策略" />
      <TextField source="concur_number" label="并发数" />
      <TextField source="bandwidth_up" label="上传带宽(Kbps)" />
      <TextField source="bandwidth_down" label="下载带宽(Kbps)" />
      <TextField source="ip_pool" label="IP池" />
      <BooleanField source="bind_mac" label="绑定MAC" />
      <BooleanField source="bind_vlan" label="绑定VLAN" />
      <BooleanField source="status" label="状态" />
      <TextField source="input_rate_limit" label="上传速率限制" />
      <TextField source="output_rate_limit" label="下载速率限制" />
      <TextField source="domain_name" label="域名" />
      <TextField source="remark" label="备注" />
      <DateField source="created_at" label="创建时间" showTime />
      <DateField source="updated_at" label="更新时间" showTime />
    </SimpleShowLayout>
  </Show>
);
