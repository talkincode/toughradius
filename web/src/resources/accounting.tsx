import {
  List,
  Datagrid,
  TextField,
  DateField,
  NumberField,
  Show,
  SimpleShowLayout,
  Filter,
  TextInput,
  DateInput,
  SelectInput,
  FunctionField,
} from 'react-admin';

// 计费记录过滤器
const AccountingFilter = (props: any) => (
  <Filter {...props}>
    <TextInput label="用户名" source="username" alwaysOn />
    <TextInput label="会话ID" source="acct_session_id" />
    <SelectInput
      label="会话状态"
      source="acct_status_type"
      choices={[
        { id: 'Start', name: '开始' },
        { id: 'Stop', name: '结束' },
        { id: 'Interim-Update', name: '中间更新' },
      ]}
    />
    <DateInput label="开始日期" source="acct_start_time_gte" />
    <DateInput label="结束日期" source="acct_start_time_lte" />
  </Filter>
);

// 格式化字节数
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
};

// 计费记录列表
export const AccountingList = () => (
  <List filters={<AccountingFilter />} sort={{ field: 'acct_start_time', order: 'DESC' }}>
    <Datagrid rowClick="show">
      <TextField source="id" label="ID" />
      <TextField source="username" label="用户名" />
      <TextField source="acct_session_id" label="会话ID" />
      <TextField source="acct_status_type" label="状态类型" />
      <TextField source="nas_addr" label="NAS地址" />
      <TextField source="framed_ipaddr" label="用户IP" />
      <DateField source="acct_start_time" label="开始时间" showTime />
      <DateField source="acct_stop_time" label="结束时间" showTime />
      <FunctionField
        label="会话时长"
        render={(record: any) => {
          if (record.acct_session_time) {
            const hours = Math.floor(record.acct_session_time / 3600);
            const minutes = Math.floor((record.acct_session_time % 3600) / 60);
            return `${hours}小时${minutes}分钟`;
          }
          return '-';
        }}
      />
      <FunctionField
        label="总流量"
        render={(record: any) => {
          const total = (record.acct_input_octets || 0) + (record.acct_output_octets || 0);
          return formatBytes(total);
        }}
      />
    </Datagrid>
  </List>
);

// 计费记录详情
export const AccountingShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="id" label="ID" />
      <TextField source="username" label="用户名" />
      <TextField source="acct_session_id" label="会话ID" />
      <TextField source="acct_status_type" label="状态类型" />
      <TextField source="nas_addr" label="NAS地址" />
      <TextField source="nas_port_id" label="NAS端口ID" />
      <TextField source="nas_port_type" label="NAS端口类型" />
      <TextField source="service_type" label="服务类型" />
      <TextField source="framed_protocol" label="帧协议" />
      <TextField source="framed_ipaddr" label="用户IP" />
      <TextField source="framed_netmask" label="子网掩码" />
      <TextField source="mac_addr" label="MAC地址" />
      <DateField source="acct_start_time" label="开始时间" showTime />
      <DateField source="acct_stop_time" label="结束时间" showTime />
      <NumberField source="acct_session_time" label="会话时长(秒)" />
      <FunctionField
        label="上传流量"
        render={(record: any) => formatBytes(record.acct_input_octets || 0)}
      />
      <FunctionField
        label="下载流量"
        render={(record: any) => formatBytes(record.acct_output_octets || 0)}
      />
      <NumberField source="acct_input_packets" label="上传包数" />
      <NumberField source="acct_output_packets" label="下载包数" />
      <TextField source="acct_terminate_cause" label="终止原因" />
      <TextField source="acct_unique_id" label="唯一ID" />
    </SimpleShowLayout>
  </Show>
);
