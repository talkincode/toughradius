import {
  List,
  Datagrid,
  TextField,
  DateField,
  NumberField,
  FunctionField,
  Show,
  SimpleShowLayout,
  TextInput,
  DateInput,
} from 'react-admin';

interface OnlineSession {
  acct_session_time?: number;
}

// 在线会话过滤器
const onlineFilters = [
  <TextInput key="username" source="username" alwaysOn />,
  <TextInput key="framed_ipaddr" source="framed_ipaddr" />,
  <TextInput key="nas_addr" source="nas_addr" />,
  <DateInput key="start_time_gte" source="start_time_gte" />,
];

// 在线会话列表
export const OnlineSessionList = () => (
  <List filters={onlineFilters} sort={{ field: 'acct_start_time', order: 'DESC' }}>
    <Datagrid rowClick="show" bulkActionButtons={false}>
      <TextField source="acct_session_id" />
      <TextField source="username" />
      <TextField source="nas_addr" />
      <TextField source="framed_ipaddr" />
      <TextField source="mac_addr" />
      <NumberField source="session_timeout" />
      <DateField source="acct_start_time" showTime />
      <FunctionField
        source="acct_session_time"
        render={(record: OnlineSession) => {
          if (!record || !record.acct_session_time) return '-';
          const hours = Math.floor(record.acct_session_time / 3600);
          const minutes = Math.floor((record.acct_session_time % 3600) / 60);
          const seconds = record.acct_session_time % 60;
          return `${hours}h ${minutes}m ${seconds}s`;
        }}
      />
      <NumberField source="acct_input_octets" />
      <NumberField source="acct_output_octets" />
    </Datagrid>
  </List>
);

// 在线会话详情
export const OnlineSessionShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="acct_session_id" />
      <TextField source="username" />
      <TextField source="nas_addr" />
      <TextField source="nas_port" />
      <TextField source="service_type" />
      <TextField source="framed_ipaddr" />
      <TextField source="framed_netmask" />
      <TextField source="mac_addr" />
      <NumberField source="session_timeout" />
      <DateField source="acct_start_time" showTime />
      <NumberField source="acct_session_time" />
      <NumberField source="acct_input_octets" />
      <NumberField source="acct_output_octets" />
      <NumberField source="acct_input_packets" />
      <NumberField source="acct_output_packets" />
    </SimpleShowLayout>
  </Show>
);
