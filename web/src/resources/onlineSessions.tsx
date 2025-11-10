import {
  List,
  Datagrid,
  TextField,
  DateField,
  FunctionField,
  Show,
  SimpleShowLayout,
  Filter,
  TextInput,
  DateInput,
  useRecordContext,
  useTranslate,
  FilterProps,
} from 'react-admin';
import { Box, Paper, Typography, Chip } from '@mui/material';
import { ReactNode } from 'react';

interface OnlineSession {
  acct_session_id?: string;
  username?: string;
  nas_addr?: string;
  framed_ipaddr?: string;
  mac_addr?: string;
  nas_port?: string | number;
  service_type?: string;
  framed_netmask?: string;
  session_timeout?: number;
  acct_start_time?: string | number;
  acct_session_time?: number;
  acct_input_octets?: number;
  acct_output_octets?: number;
  acct_input_packets?: number;
  acct_output_packets?: number;
}

const formatDuration = (seconds?: number): string => {
  if (seconds === undefined || seconds === null) {
    return '-';
  }
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;
  const parts = [];
  if (hours) {
    parts.push(`${hours}h`);
  }
  if (minutes) {
    parts.push(`${minutes}m`);
  }
  parts.push(`${secs}s`);
  return parts.join(' ');
};

const formatBytes = (bytes?: number): string => {
  if (bytes === undefined || bytes === null) {
    return '-';
  }
  if (bytes === 0) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = bytes;
  let index = 0;
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024;
    index += 1;
  }
  return `${parseFloat(value.toFixed(2))} ${units[index]}`;
};

const formatTimestamp = (value?: string | number): string => {
  if (!value) {
    return '-';
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return '-';
  }
  return date.toLocaleString();
};

const OnlineSessionFilter = (props: FilterProps) => {
  const translate = useTranslate();
  return (
    <Filter {...props}>
      <TextInput
        source="username"
        label={translate('resources.radius/online.fields.username')}
        alwaysOn
      />
      <TextInput
        source="framed_ipaddr"
        label={translate('resources.radius/online.fields.framed_ipaddr')}
      />
      <TextInput
        source="nas_addr"
        label={translate('resources.radius/online.fields.nas_addr')}
      />
      <DateInput
        source="start_time_gte"
        label={translate('resources.radius/online.fields.start_time_gte')}
      />
      <DateInput
        source="start_time_lte"
        label={translate('resources.radius/online.fields.start_time_lte')}
      />
    </Filter>
  );
};

const NasPortField = () => {
  const translate = useTranslate();
  return (
    <FunctionField
      source="nas_port"
      label={translate('resources.radius/online.fields.nas_port')}
      render={(record: OnlineSession) => {
        if (!record?.nas_port) {
          return '-';
        }
        return <Chip label={String(record.nas_port)} size="small" variant="outlined" />;
      }}
    />
  );
};

const TimeoutField = () => {
  const translate = useTranslate();
  return (
    <FunctionField
      source="session_timeout"
      label={translate('resources.radius/online.fields.session_timeout')}
      render={(record: OnlineSession) => {
        if (record?.session_timeout === undefined || record?.session_timeout === null) {
          return '-';
        }
        return `${record.session_timeout}s`;
      }}
    />
  );
};

const SessionDurationField = () => {
  const translate = useTranslate();
  return (
    <FunctionField
      source="acct_session_time"
      label={translate('resources.radius/online.fields.session_time')}
      render={(record: OnlineSession) => formatDuration(record.acct_session_time)}
    />
  );
};

const UploadField = () => {
  const translate = useTranslate();
  return (
    <FunctionField
      source="acct_input_octets"
      label={translate('resources.radius/online.fields.acct_input_octets')}
      render={(record: OnlineSession) => formatBytes(record.acct_input_octets)}
    />
  );
};

const DownloadField = () => {
  const translate = useTranslate();
  return (
    <FunctionField
      source="acct_output_octets"
      label={translate('resources.radius/online.fields.acct_output_octets')}
      render={(record: OnlineSession) => formatBytes(record.acct_output_octets)}
    />
  );
};

interface DetailRow {
  label: string;
  value?: ReactNode;
}

interface DetailSectionProps {
  title: string;
  description?: string;
  rows: DetailRow[];
}

const DetailSection = ({ title, description, rows }: DetailSectionProps) => (
  <Paper
    elevation={0}
    sx={{
      borderRadius: 2,
      border: theme => `1px solid ${theme.palette.divider}`,
      backgroundColor: theme.palette.background.paper,
      p: { xs: 2, md: 3 },
      width: '100%',
    }}
  >
    <Box>
      <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
        {title}
      </Typography>
      {description && (
        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
          {description}
        </Typography>
      )}
    </Box>
    <Box
      sx={{
        mt: 2,
        display: 'grid',
        gap: 2,
        gridTemplateColumns: {
          xs: 'repeat(1, minmax(0, 1fr))',
          sm: 'repeat(2, minmax(0, 1fr))',
          md: 'repeat(3, minmax(0, 1fr))',
        },
      }}
    >
      {rows.map(row => (
        <Box key={row.label}>
          <Typography variant="caption" color="text.secondary">
            {row.label}
          </Typography>
          <Typography variant="body1" sx={{ mt: 0.5, fontWeight: 500, wordBreak: 'break-word' }}>
            {row.value ?? '-'}
          </Typography>
        </Box>
      ))}
    </Box>
  </Paper>
);

const OnlineSessionDetails = () => {
  const record = useRecordContext<OnlineSession>();
  const translate = useTranslate();
  if (!record) {
    return null;
  }

  const totalTraffic =
    (record.acct_input_octets ?? 0) + (record.acct_output_octets ?? 0);

  const sections = [
    {
      title: translate('resources.radius/online.sections.overview'),
      description: translate('resources.radius/online.sections.overview_desc'),
      rows: [
        {
          label: translate('resources.radius/online.fields.acct_session_id'),
          value: record.acct_session_id,
        },
        { label: translate('resources.radius/online.fields.username'), value: record.username },
        { label: translate('resources.radius/online.fields.nas_addr'), value: record.nas_addr },
        { label: translate('resources.radius/online.fields.nas_port'), value: record.nas_port },
        { label: translate('resources.radius/online.fields.service_type'), value: record.service_type },
      ],
    },
    {
      title: translate('resources.radius/online.sections.device'),
      description: translate('resources.radius/online.sections.device_desc'),
      rows: [
        { label: translate('resources.radius/online.fields.framed_ipaddr'), value: record.framed_ipaddr },
        { label: translate('resources.radius/online.fields.framed_netmask'), value: record.framed_netmask },
        { label: translate('resources.radius/online.fields.mac_addr'), value: record.mac_addr },
      ],
    },
    {
      title: translate('resources.radius/online.sections.timing'),
      description: translate('resources.radius/online.sections.timing_desc'),
      rows: [
        {
          label: translate('resources.radius/online.fields.session_timeout'),
          value: record.session_timeout !== undefined && record.session_timeout !== null
            ? `${record.session_timeout}s`
            : undefined,
        },
        {
          label: translate('resources.radius/online.fields.acct_start_time'),
          value: formatTimestamp(record.acct_start_time),
        },
        {
          label: translate('resources.radius/online.fields.session_time'),
          value: formatDuration(record.acct_session_time),
        },
      ],
    },
    {
      title: translate('resources.radius/online.sections.traffic'),
      description: translate('resources.radius/online.sections.traffic_desc'),
      rows: [
        {
          label: translate('resources.radius/online.fields.acct_input_octets'),
          value: formatBytes(record.acct_input_octets),
        },
        {
          label: translate('resources.radius/online.fields.acct_output_octets'),
          value: formatBytes(record.acct_output_octets),
        },
        {
          label: translate('resources.radius/online.fields.total_traffic'),
          value: formatBytes(totalTraffic),
        },
        {
          label: translate('resources.radius/online.fields.acct_input_packets'),
          value: record.acct_input_packets,
        },
        {
          label: translate('resources.radius/online.fields.acct_output_packets'),
          value: record.acct_output_packets,
        },
      ],
    },
  ];

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, width: '100%', mt: 1 }}>
      {sections.map(section => (
        <DetailSection
          key={section.title}
          title={section.title}
          description={section.description}
          rows={section.rows}
        />
      ))}
    </Box>
  );
};

export const OnlineSessionList = () => {
  const translate = useTranslate();
  return (
    <List filters={<OnlineSessionFilter />} sort={{ field: 'acct_start_time', order: 'DESC' }}>
      <Datagrid
        rowClick="show"
        bulkActionButtons={false}
        sx={{
          '& .RaDatagrid-headerCell': {
            fontWeight: 600,
          },
          '& .RaDatagrid-row': {
            borderRadius: 2,
            transition: 'box-shadow 0.2s ease',
          },
          '& .RaDatagrid-row:hover': {
            boxShadow: theme => `0 4px 12px ${theme.palette.divider}`,
          },
        }}
      >
        <TextField
          source="acct_session_id"
          label={translate('resources.radius/online.fields.acct_session_id')}
        />
        <TextField source="username" label={translate('resources.radius/online.fields.username')} />
        <TextField
          source="framed_ipaddr"
          label={translate('resources.radius/online.fields.framed_ipaddr')}
        />
        <TextField source="nas_addr" label={translate('resources.radius/online.fields.nas_addr')} />
        <NasPortField />
        <TimeoutField />
        <DateField
          source="acct_start_time"
          label={translate('resources.radius/online.fields.acct_start_time')}
          showTime
        />
        <SessionDurationField />
        <UploadField />
        <DownloadField />
        <TextField source="mac_addr" label={translate('resources.radius/online.fields.mac_addr')} />
      </Datagrid>
    </List>
  );
};

export const OnlineSessionShow = () => (
  <Show>
    <SimpleShowLayout>
      <OnlineSessionDetails />
    </SimpleShowLayout>
  </Show>
);
