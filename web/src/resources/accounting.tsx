import {
  List,
  Datagrid,
  TextField,
  DateField,
  FunctionField,
  Show,
  Filter,
  TextInput,
  DateInput,
  useRecordContext,
  useTranslate,
  FilterProps,
} from 'react-admin';
import { Box, Paper, Typography, Chip } from '@mui/material';
import { ReactNode } from 'react';

interface AccountingRecord {
  id?: string;
  acct_session_id?: string;
  username?: string;
  nas_id?: string;
  nas_addr?: string;
  nas_paddr?: string;
  nas_port?: string | number;
  nas_class?: string;
  nas_port_id?: string;
  nas_port_type?: number;
  service_type?: number;
  session_timeout?: number;
  framed_ipaddr?: string;
  framed_netmask?: string;
  framed_ipv6_prefix?: string;
  framed_ipv6_address?: string;
  delegated_ipv6_prefix?: string;
  mac_addr?: string;
  acct_session_time?: number;
  acct_input_total?: number;
  acct_output_total?: number;
  acct_input_packets?: number;
  acct_output_packets?: number;
  acct_start_time?: string | number;
  acct_stop_time?: string | number;
  last_update?: string | number;
  acct_terminate_cause?: string;
  acct_unique_id?: string;
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

const AccountingFilter = (props: Omit<FilterProps, 'children'>) => {
  const translate = useTranslate();
  return (
    <Filter {...props}>
      <TextInput
        source="username"
        label={translate('resources.radius/accounting.fields.username')}
        alwaysOn
      />
      <TextInput
        source="acct_session_id"
        label={translate('resources.radius/accounting.fields.acct_session_id')}
      />
      <TextInput
        source="framed_ipaddr"
        label={translate('resources.radius/accounting.fields.framed_ipaddr')}
      />
      <TextInput
        source="nas_addr"
        label={translate('resources.radius/accounting.fields.nas_addr')}
      />
      <DateInput
        source="acct_start_time_gte"
        label={translate('resources.radius/accounting.fields.acct_start_time_gte')}
      />
      <DateInput
        source="acct_start_time_lte"
        label={translate('resources.radius/accounting.fields.acct_start_time_lte')}
      />
    </Filter>
  );
};

const NasPortField = () => {
  const translate = useTranslate();
  return (
    <FunctionField
      source="nas_port"
      label={translate('resources.radius/accounting.fields.nas_port')}
      render={(record: AccountingRecord) => {
        if (!record?.nas_port) {
          return '-';
        }
        return <Chip label={String(record.nas_port)} size="small" variant="outlined" />;
      }}
    />
  );
};

const SessionDurationField = () => {
  const translate = useTranslate();
  return (
    <FunctionField
      source="acct_session_time"
      label={translate('resources.radius/accounting.fields.session_time')}
      render={(record: AccountingRecord) => formatDuration(record.acct_session_time)}
    />
  );
};

const UploadField = () => {
  const translate = useTranslate();
  return (
    <FunctionField
      source="acct_input_total"
      label={translate('resources.radius/accounting.fields.acct_input_total')}
      render={(record: AccountingRecord) => formatBytes(record.acct_input_total)}
    />
  );
};

const DownloadField = () => {
  const translate = useTranslate();
  return (
    <FunctionField
      source="acct_output_total"
      label={translate('resources.radius/accounting.fields.acct_output_total')}
      render={(record: AccountingRecord) => formatBytes(record.acct_output_total)}
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
      borderRadius: 1.5,
      border: theme => `1px solid ${theme.palette.divider}`,
      backgroundColor: theme => theme.palette.background.paper,
      p: { xs: 1.5, md: 2 },
      width: '100%',
    }}
  >
    <Box>
      <Typography variant="subtitle2" sx={{ fontWeight: 600, fontSize: '0.9rem' }}>
        {title}
      </Typography>
      {description && (
        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.25, fontSize: '0.8rem' }}>
          {description}
        </Typography>
      )}
    </Box>
    <Box
      sx={{
        mt: 1.5,
        display: 'grid',
        gap: 1.5,
        gridTemplateColumns: {
          xs: 'repeat(1, minmax(0, 1fr))',
          sm: 'repeat(2, minmax(0, 1fr))',
          md: 'repeat(3, minmax(0, 1fr))',
          lg: 'repeat(4, minmax(0, 1fr))',
        },
      }}
    >
      {rows.map(row => (
        <Box key={row.label}>
          <Typography variant="caption" color="text.secondary" sx={{ fontSize: '0.75rem' }}>
            {row.label}
          </Typography>
          <Typography variant="body2" sx={{ mt: 0.25, fontWeight: 500, wordBreak: 'break-word', fontSize: '0.85rem' }}>
            {row.value ?? '-'}
          </Typography>
        </Box>
      ))}
    </Box>
  </Paper>
);

const AccountingDetails = () => {
  const record = useRecordContext<AccountingRecord>();
  const translate = useTranslate();
  if (!record) {
    return null;
  }

  const totalTraffic =
    (record.acct_input_total ?? 0) + (record.acct_output_total ?? 0);

  const sections = [
    {
      title: translate('resources.radius/accounting.sections.overview'),
      description: translate('resources.radius/accounting.sections.overview_desc'),
      rows: [
        {
          label: translate('resources.radius/accounting.fields.acct_session_id'),
          value: record.acct_session_id,
        },
        {
          label: translate('resources.radius/accounting.fields.username'),
          value: record.username,
        },
        {
          label: translate('resources.radius/accounting.fields.nas_addr'),
          value: record.nas_addr,
        },
        {
          label: translate('resources.radius/accounting.fields.nas_id'),
          value: record.nas_id,
        },
        {
          label: translate('resources.radius/accounting.fields.nas_port'),
          value: record.nas_port,
        },
        {
          label: translate('resources.radius/accounting.fields.service_type'),
          value: record.service_type,
        },
      ],
    },
    {
      title: translate('resources.radius/accounting.sections.device'),
      description: translate('resources.radius/accounting.sections.device_desc'),
      rows: [
        {
          label: translate('resources.radius/accounting.fields.framed_ipaddr'),
          value: record.framed_ipaddr,
        },
        {
          label: translate('resources.radius/accounting.fields.framed_netmask'),
          value: record.framed_netmask,
        },
        {
          label: translate('resources.radius/accounting.fields.mac_addr'),
          value: record.mac_addr,
        },
        {
          label: translate('resources.radius/accounting.fields.framed_ipv6_address'),
          value: record.framed_ipv6_address,
        },
        {
          label: translate('resources.radius/accounting.fields.framed_ipv6_prefix'),
          value: record.framed_ipv6_prefix,
        },
      ],
    },
    {
      title: translate('resources.radius/accounting.sections.timing'),
      description: translate('resources.radius/accounting.sections.timing_desc'),
      rows: [
        {
          label: translate('resources.radius/accounting.fields.acct_start_time'),
          value: formatTimestamp(record.acct_start_time),
        },
        {
          label: translate('resources.radius/accounting.fields.acct_stop_time'),
          value: formatTimestamp(record.acct_stop_time),
        },
        {
          label: translate('resources.radius/accounting.fields.session_time'),
          value: formatDuration(record.acct_session_time),
        },
        {
          label: translate('resources.radius/accounting.fields.session_timeout'),
          value: record.session_timeout !== undefined && record.session_timeout !== null
            ? `${record.session_timeout}s`
            : undefined,
        },
        {
          label: translate('resources.radius/accounting.fields.last_update'),
          value: formatTimestamp(record.last_update),
        },
      ],
    },
    {
      title: translate('resources.radius/accounting.sections.traffic'),
      description: translate('resources.radius/accounting.sections.traffic_desc'),
      rows: [
        {
          label: translate('resources.radius/accounting.fields.acct_input_total'),
          value: formatBytes(record.acct_input_total),
        },
        {
          label: translate('resources.radius/accounting.fields.acct_output_total'),
          value: formatBytes(record.acct_output_total),
        },
        {
          label: translate('resources.radius/accounting.fields.total_traffic'),
          value: formatBytes(totalTraffic),
        },
        {
          label: translate('resources.radius/accounting.fields.acct_input_packets'),
          value: record.acct_input_packets,
        },
        {
          label: translate('resources.radius/accounting.fields.acct_output_packets'),
          value: record.acct_output_packets,
        },
      ],
    },
    {
      title: translate('resources.radius/accounting.sections.session_details'),
      description: translate('resources.radius/accounting.sections.session_details_desc'),
      rows: [
        {
          label: translate('resources.radius/accounting.fields.nas_class'),
          value: record.nas_class,
        },
        {
          label: translate('resources.radius/accounting.fields.nas_port_id'),
          value: record.nas_port_id,
        },
        {
          label: translate('resources.radius/accounting.fields.nas_port_type'),
          value: record.nas_port_type,
        },
        {
          label: translate('resources.radius/accounting.fields.acct_terminate_cause'),
          value: record.acct_terminate_cause,
        },
      ],
    },
  ];

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5, width: '100%', mt: 0.5 }}>
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

export const AccountingList = () => {
  const translate = useTranslate();
  return (
    <List filters={<AccountingFilter />} sort={{ field: 'acct_start_time', order: 'DESC' }}>
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
          label={translate('resources.radius/accounting.fields.acct_session_id')}
        />
        <TextField
          source="username"
          label={translate('resources.radius/accounting.fields.username')}
        />
        <TextField
          source="framed_ipaddr"
          label={translate('resources.radius/accounting.fields.framed_ipaddr')}
        />
        <TextField
          source="nas_addr"
          label={translate('resources.radius/accounting.fields.nas_addr')}
        />
        <NasPortField />
        <DateField
          source="acct_start_time"
          label={translate('resources.radius/accounting.fields.acct_start_time')}
          showTime
        />
        <DateField
          source="acct_stop_time"
          label={translate('resources.radius/accounting.fields.acct_stop_time')}
          showTime
        />
        <SessionDurationField />
        <UploadField />
        <DownloadField />
        <TextField
          source="mac_addr"
          label={translate('resources.radius/accounting.fields.mac_addr')}
        />
      </Datagrid>
    </List>
  );
};

export const AccountingShow = () => (
  <Show>
    <AccountingDetails />
  </Show>
);
