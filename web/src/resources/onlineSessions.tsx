import {
  List,
  Datagrid,
  TextField,
  DateField,
  NumberField,
  FunctionField,
  Show,
  SimpleShowLayout,
  Filter,
  TextInput,
  DateInput,
} from 'react-admin';

// 在线会话过滤器
const SessionFilter = (props: any) => (
  <Filter {...props}>
    <TextInput label="用户名" source="username" alwaysOn />
    <TextInput label="IP地址" source="framed_ipaddr" />
    <TextInput label="NAS IP" source="nas_addr" />
    <DateInput label="开始时间" source="start_time_gte" />
  </Filter>
);

// 在线会话列表
export const OnlineSessionList = () => (
  <List filters={<SessionFilter />}>
    <Datagrid rowClick="show">
      <TextField source="acct_session_id" label="会话ID" />
      <TextField source="username" label="用户名" />
      <TextField source="nas_addr" label="NAS地址" />
      <TextField source="framed_ipaddr" label="用户IP" />
      <TextField source="mac_addr" label="MAC地址" />
      <NumberField source="session_timeout" label="超时时间(秒)" />
      <DateField source="acct_start_time" label="开始时间" showTime />
      <FunctionField
        label="在线时长"
        render={(record: any) => {
          if (record.acct_start_time) {
            const duration = Math.floor(
              (Date.now() - new Date(record.acct_start_time).getTime()) / 1000
            );
            const hours = Math.floor(duration / 3600);
            const minutes = Math.floor((duration % 3600) / 60);
            return `${hours}小时${minutes}分钟`;
          }
          return '-';
        }}
      />
      <NumberField source="acct_input_octets" label="上传流量(B)" />
      <NumberField source="acct_output_octets" label="下载流量(B)" />
    </Datagrid>
  </List>
);

// 在线会话详情
export const OnlineSessionShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="acct_session_id" label="会话ID" />
      <TextField source="username" label="用户名" />
      <TextField source="nas_addr" label="NAS地址" />
      <TextField source="nas_port" label="NAS端口" />
      <TextField source="service_type" label="服务类型" />
      <TextField source="framed_ipaddr" label="用户IP" />
      <TextField source="framed_netmask" label="子网掩码" />
      <TextField source="mac_addr" label="MAC地址" />
      <NumberField source="session_timeout" label="超时时间(秒)" />
      <DateField source="acct_start_time" label="开始时间" showTime />
      <NumberField source="acct_session_time" label="会话时长(秒)" />
      <NumberField source="acct_input_octets" label="上传流量(B)" />
      <NumberField source="acct_output_octets" label="下载流量(B)" />
      <NumberField source="acct_input_packets" label="上传包数" />
      <NumberField source="acct_output_packets" label="下载包数" />
    </SimpleShowLayout>
  </Show>
);
