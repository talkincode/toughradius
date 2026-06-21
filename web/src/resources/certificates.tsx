import {
  List,
  Datagrid,
  TextField,
  DateField,
  BooleanField,
  Edit,
  SimpleForm,
  TextInput,
  PasswordInput,
  SelectInput,
  Create,
  Show,
  TopToolbar,
  CreateButton,
  SortButton,
  ShowButton,
  EditButton,
  DeleteButton,
  ListButton,
  Button,
  Toolbar,
  SaveButton,
  required,
  maxLength,
  useRecordContext,
  useTranslate,
  useListContext,
  useNotify,
  RaRecord,
  FunctionField,
  ToolbarProps,
} from 'react-admin';
import {
  Box,
  Card,
  CardContent,
  Stack,
  Typography,
  Chip,
  Avatar,
  IconButton,
  Tooltip,
  TextField as MuiTextField,
  MenuItem,
  alpha,
} from '@mui/material';
import {
  VerifiedUser as VerifiedUserIcon,
  Badge as BadgeIcon,
  Download as DownloadIcon,
  FilterList as FilterIcon,
  Search as SearchIcon,
  Clear as ClearIcon,
  Key as KeyIcon,
  Article as ArticleIcon,
  CalendarMonth as CalendarIcon,
} from '@mui/icons-material';
import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  ServerPagination,
  ActiveFilters,
  FormSection,
  FieldGrid,
  FieldGridItem,
  formLayoutSx,
  DetailItem,
  DetailSectionCard,
  EmptyValue,
} from '../components';
import { API_BASE } from '../utils/apiClient';

const LARGE_LIST_PER_PAGE = 50;

export const CertificateIcon = VerifiedUserIcon;

interface Certificate extends RaRecord {
  name?: string;
  cert_type?: 'server' | 'ca';
  cert?: string;
  subject?: string;
  issuer?: string;
  serial?: string;
  fingerprint?: string;
  not_before?: string;
  not_after?: string;
  has_key?: boolean;
  remark?: string;
  created_at?: string;
  updated_at?: string;
}

interface CertificateFormValues extends Partial<Certificate> {
  private_key?: string;
}

const getCertTypeColor = (certType?: string): 'primary' | 'info' | 'default' => {
  if (certType === 'server') return 'primary';
  if (certType === 'ca') return 'info';
  return 'default';
};

const gridSx = {
  display: 'grid',
  gap: 2,
  gridTemplateColumns: {
    xs: 'repeat(1, 1fr)',
    sm: 'repeat(2, 1fr)',
  },
};

const formatTimestamp = (value?: string): string => {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '-';
  return date.toLocaleString();
};

const buildCertificateTypeChoices = (translate: ReturnType<typeof useTranslate>) => [
  { id: 'server', name: translate('resources.system/certificate.cert_types.server', { _: 'Server' }) },
  { id: 'ca', name: translate('resources.system/certificate.cert_types.ca', { _: 'CA' }) },
];

const CertificateTypeChip = () => {
  const record = useRecordContext<Certificate>();
  const translate = useTranslate();
  if (!record?.cert_type) return <EmptyValue />;

  return (
    <Chip
      label={translate(`resources.system/certificate.cert_types.${record.cert_type}`, { _: record.cert_type })}
      size="small"
      color={getCertTypeColor(record.cert_type)}
      variant={record.cert_type === 'server' ? 'filled' : 'outlined'}
      sx={{ height: 22, fontWeight: 500, fontSize: '0.75rem' }}
    />
  );
};

const CertificateNameField = () => {
  const record = useRecordContext<Certificate>();
  if (!record) return null;

  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
      <Avatar sx={{ width: 32, height: 32, bgcolor: 'primary.main', fontSize: '0.85rem', fontWeight: 600 }}>
        <VerifiedUserIcon sx={{ fontSize: 18 }} />
      </Avatar>
      <Box>
        <Typography variant="body2" sx={{ fontWeight: 600, color: 'text.primary', lineHeight: 1.3 }}>
          {record.name || '-'}
        </Typography>
        <CertificateTypeChip />
      </Box>
    </Box>
  );
};

const HasKeyField = () => {
  const record = useRecordContext<Certificate>();
  const translate = useTranslate();
  if (!record) return null;

  return (
    <Chip
      label={record.has_key ? translate('ra.boolean.true', { _: 'Yes' }) : translate('ra.boolean.false', { _: 'No' })}
      size="small"
      color={record.has_key ? 'success' : 'default'}
      variant={record.has_key ? 'filled' : 'outlined'}
      sx={{ height: 22, fontWeight: 500, fontSize: '0.75rem' }}
    />
  );
};

const getFilenameFromDisposition = (disposition: string | null, fallback: string): string => {
  if (!disposition) return fallback;
  const utf8Match = disposition.match(/filename\*=UTF-8''([^;]+)/i);
  if (utf8Match?.[1]) return decodeURIComponent(utf8Match[1]);
  const filenameMatch = disposition.match(/filename="?([^";]+)"?/i);
  if (filenameMatch?.[1]) return filenameMatch[1];
  return fallback;
};

const CertificateExportButton = ({ label }: { label?: string }) => {
  const record = useRecordContext<Certificate>();
  const notify = useNotify();
  const translate = useTranslate();
  const [downloading, setDownloading] = useState(false);

  const handleExport = useCallback(async (event: React.MouseEvent) => {
    event.stopPropagation();
    if (!record?.id) return;

    setDownloading(true);
    try {
      const headers = new Headers({ Accept: 'application/octet-stream' });
      const token = localStorage.getItem('token');
      if (token) headers.set('Authorization', `Bearer ${token}`);

      const response = await fetch(`${API_BASE}/system/certificate/${encodeURIComponent(String(record.id))}/export`, { headers });
      if (!response.ok) {
        const message = await response.text().catch(() => '');
        throw new Error(message || `HTTP ${response.status}`);
      }

      const blob = await response.blob();
      const safeName = (record.name || String(record.id)).replace(/[\\/:*?"<>|]+/g, '_');
      const filename = getFilenameFromDisposition(response.headers.get('Content-Disposition'), `${safeName}.pem`);
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      anchor.href = url;
      anchor.download = filename;
      document.body.appendChild(anchor);
      anchor.click();
      anchor.remove();
      window.setTimeout(() => URL.revokeObjectURL(url), 1000);
      notify(translate('resources.system/certificate.notifications.export_success', { _: 'Certificate exported' }), { type: 'info' });
    } catch (error) {
      notify(
        `${translate('resources.system/certificate.notifications.export_error', { _: 'Export failed' })}: ${error instanceof Error ? error.message : String(error)}`,
        { type: 'error' },
      );
    } finally {
      setDownloading(false);
    }
  }, [notify, record, translate]);

  if (!record) return null;

  return (
    <Button
      label={label ?? translate('resources.system/certificate.actions.export', { _: 'Export' })}
      onClick={handleExport}
      disabled={downloading}
    >
      <DownloadIcon />
    </Button>
  );
};

const CertificateRowActions = () => (
  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
    <ShowButton />
    <EditButton />
    <CertificateExportButton label="" />
    <DeleteButton mutationMode="pessimistic" />
  </Box>
);

const CertificateFormToolbar = (props: ToolbarProps) => (
  <Toolbar {...props}>
    <SaveButton />
    <DeleteButton mutationMode="pessimistic" />
  </Toolbar>
);

const CertificateListActions = () => {
  const translate = useTranslate();
  return (
    <TopToolbar>
      <SortButton fields={['created_at', 'name', 'not_after']} label={translate('ra.action.sort', { _: '排序' })} />
      <CreateButton />
    </TopToolbar>
  );
};

const CertificateSearchHeaderCard = () => {
  const translate = useTranslate();
  const { filterValues, setFilters, displayedFilters } = useListContext();
  const [localFilters, setLocalFilters] = useState<Record<string, string>>({});
  const certTypeChoices = useMemo(() => buildCertificateTypeChoices(translate), [translate]);

  useEffect(() => {
    const newLocalFilters: Record<string, string> = {};
    if (filterValues) {
      Object.entries(filterValues).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          newLocalFilters[key] = String(value);
        }
      });
    }
    setLocalFilters(newLocalFilters);
  }, [filterValues]);

  const handleSearch = useCallback(() => {
    const newFilters: Record<string, string> = {};
    Object.entries(localFilters).forEach(([key, value]) => {
      if (value.trim()) newFilters[key] = value.trim();
    });
    setFilters(newFilters, displayedFilters);
  }, [displayedFilters, localFilters, setFilters]);

  const handleClear = useCallback(() => {
    setLocalFilters({});
    setFilters({}, displayedFilters);
  }, [displayedFilters, setFilters]);

  const handleKeyPress = useCallback((event: React.KeyboardEvent) => {
    if (event.key === 'Enter') handleSearch();
  }, [handleSearch]);

  return (
    <Card elevation={0} sx={{ mb: 2, borderRadius: 2, border: theme => `1px solid ${theme.palette.divider}`, overflow: 'hidden' }}>
      <Box
        sx={{
          px: 2.5,
          py: 1.5,
          bgcolor: theme => (theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.03)' : 'rgba(0,0,0,0.02)'),
          borderBottom: theme => `1px solid ${theme.palette.divider}`,
          display: 'flex',
          alignItems: 'center',
          gap: 1.5,
        }}
      >
        <FilterIcon sx={{ color: 'primary.main', fontSize: 20 }} />
        <Typography variant="subtitle2" sx={{ fontWeight: 600, color: 'text.primary' }}>
          {translate('resources.system/certificate.filter.title', { _: 'Filters' })}
        </Typography>
      </Box>
      <CardContent sx={{ p: 2 }}>
        <Box
          sx={{
            display: 'grid',
            gap: 1.5,
            gridTemplateColumns: { xs: 'repeat(1, 1fr)', sm: 'repeat(2, 1fr)', md: 'repeat(4, 1fr)' },
            alignItems: 'end',
          }}
        >
          <MuiTextField
            label={translate('resources.system/certificate.fields.name', { _: 'Name' })}
            value={localFilters.name || ''}
            onChange={event => setLocalFilters(prev => ({ ...prev, name: event.target.value }))}
            onKeyPress={handleKeyPress}
            size="small"
            fullWidth
          />
          <MuiTextField
            select
            label={translate('resources.system/certificate.fields.cert_type', { _: 'Type' })}
            value={localFilters.cert_type || ''}
            onChange={event => setLocalFilters(prev => ({ ...prev, cert_type: event.target.value }))}
            size="small"
            fullWidth
          >
            <MenuItem value="">{translate('ra.action.clear_input_value', { _: 'All' })}</MenuItem>
            {certTypeChoices.map(choice => (
              <MenuItem key={choice.id} value={choice.id}>{choice.name}</MenuItem>
            ))}
          </MuiTextField>
          <Box sx={{ display: 'flex', gap: 0.5, justifyContent: 'flex-end' }}>
            <Tooltip title={translate('ra.action.clear_filters', { _: 'Clear filters' })}>
              <IconButton onClick={handleClear} size="small" sx={{ bgcolor: theme => alpha(theme.palette.grey[500], 0.1) }}>
                <ClearIcon />
              </IconButton>
            </Tooltip>
            <Tooltip title={translate('ra.action.search', { _: 'Search' })}>
              <IconButton onClick={handleSearch} color="primary" sx={{ bgcolor: theme => alpha(theme.palette.primary.main, 0.1) }}>
                <SearchIcon />
              </IconButton>
            </Tooltip>
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

const CertificateListContent = () => {
  const translate = useTranslate();
  const { data, total } = useListContext<Certificate>();

  const fieldLabels = useMemo(() => ({
    name: translate('resources.system/certificate.fields.name', { _: 'Name' }),
    cert_type: translate('resources.system/certificate.fields.cert_type', { _: 'Type' }),
  }), [translate]);

  const valueLabels = useMemo(() => ({
    cert_type: {
      server: translate('resources.system/certificate.cert_types.server', { _: 'Server' }),
      ca: translate('resources.system/certificate.cert_types.ca', { _: 'CA' }),
    },
  }), [translate]);

  if (!data || data.length === 0) {
    return (
      <Box>
        <CertificateSearchHeaderCard />
        <Card elevation={0} sx={{ borderRadius: 2, border: theme => `1px solid ${theme.palette.divider}` }}>
          <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', py: 8, color: 'text.secondary' }}>
            <VerifiedUserIcon sx={{ fontSize: 64, opacity: 0.3, mb: 2 }} />
            <Typography variant="h6" sx={{ opacity: 0.6, mb: 1 }}>
              {translate('resources.system/certificate.empty.title', { _: 'No certificates' })}
            </Typography>
            <Typography variant="body2" sx={{ opacity: 0.5 }}>
              {translate('resources.system/certificate.empty.description', { _: 'Import a certificate to get started' })}
            </Typography>
          </Box>
        </Card>
      </Box>
    );
  }

  return (
    <Box>
      <CertificateSearchHeaderCard />
      <ActiveFilters fieldLabels={fieldLabels} valueLabels={valueLabels} />
      <Card elevation={0} sx={{ borderRadius: 2, border: theme => `1px solid ${theme.palette.divider}`, overflow: 'hidden' }}>
        <Box sx={{ px: 2, py: 1, bgcolor: theme => (theme.palette.mode === 'dark' ? 'rgba(255,255,255,0.02)' : 'rgba(0,0,0,0.01)'), borderBottom: theme => `1px solid ${theme.palette.divider}` }}>
          <Typography variant="body2" color="text.secondary">
            {translate('resources.system/certificate.list.total', { total: total?.toLocaleString() || 0, _: 'Total %{total} certificates' })}
          </Typography>
        </Box>
        <Box sx={{ overflowX: 'auto' }}>
          <Datagrid rowClick="show" bulkActionButtons={false}>
            <FunctionField source="name" label={translate('resources.system/certificate.fields.name', { _: 'Name' })} render={() => <CertificateNameField />} />
            <FunctionField source="cert_type" label={translate('resources.system/certificate.fields.cert_type', { _: 'Type' })} render={() => <CertificateTypeChip />} />
            <TextField source="subject" label={translate('resources.system/certificate.fields.subject', { _: 'Subject' })} />
            <DateField source="not_after" label={translate('resources.system/certificate.fields.not_after', { _: 'Not After' })} showTime />
            <FunctionField source="has_key" label={translate('resources.system/certificate.fields.has_key', { _: 'Has Key' })} render={() => <HasKeyField />} />
            <FunctionField label={translate('ra.action.actions', { _: 'Actions' })} render={() => <CertificateRowActions />} />
          </Datagrid>
        </Box>
      </Card>
    </Box>
  );
};

export const CertificateList = () => (
  <List
    actions={<CertificateListActions />}
    sort={{ field: 'created_at', order: 'DESC' }}
    perPage={LARGE_LIST_PER_PAGE}
    pagination={<ServerPagination />}
    empty={false}
  >
    <CertificateListContent />
  </List>
);

const transformCertificateCreate = (data: CertificateFormValues) => {
  const payload: CertificateFormValues = { ...data };
  if (!payload.private_key) delete payload.private_key;
  return payload;
};

const transformCertificateEdit = (data: CertificateFormValues, options?: { previousData?: CertificateFormValues }) => {
  const previousData = options?.previousData ?? {};
  const payload: CertificateFormValues = {};

  if (data.name !== previousData.name) payload.name = data.name;
  if ((data.remark ?? '') !== (previousData.remark ?? '')) payload.remark = data.remark ?? '';
  if (data.cert && data.cert !== previousData.cert) payload.cert = data.cert;
  if (data.private_key) payload.private_key = data.private_key;

  return payload;
};

export const CertificateCreate = () => {
  const translate = useTranslate();
  const certTypeChoices = useMemo(() => buildCertificateTypeChoices(translate), [translate]);

  return (
    <Create transform={transformCertificateCreate}>
      <SimpleForm sx={formLayoutSx}>
        <FormSection
          title={translate('resources.system/certificate.sections.import.title', { _: 'Import Certificate' })}
          description={translate('resources.system/certificate.sections.import.description', { _: 'Paste PEM certificate material securely' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput
                source="name"
                label={translate('resources.system/certificate.fields.name', { _: 'Name' })}
                validate={[required(), maxLength(100)]}
                helperText={translate('resources.system/certificate.helpers.name', { _: 'Unique local certificate name' })}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem>
              <SelectInput
                source="cert_type"
                label={translate('resources.system/certificate.fields.cert_type', { _: 'Type' })}
                validate={[required()]}
                choices={certTypeChoices}
                defaultValue="server"
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="cert"
                label={translate('resources.system/certificate.fields.cert', { _: 'Certificate (PEM)' })}
                validate={[required()]}
                multiline
                minRows={8}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <PasswordInput
                source="private_key"
                label={translate('resources.system/certificate.fields.private_key', { _: 'Private Key (PEM)' })}
                multiline
                minRows={8}
                fullWidth
                size="small"
                helperText={translate('resources.system/certificate.helpers.private_key', { _: 'Required for server certificates; optional for CA certificates' })}
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="remark"
                label={translate('resources.system/certificate.fields.remark', { _: 'Remark' })}
                validate={[maxLength(500)]}
                multiline
                minRows={3}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Create>
  );
};

export const CertificateEdit = () => {
  const translate = useTranslate();

  return (
    <Edit transform={transformCertificateEdit}>
      <SimpleForm toolbar={<CertificateFormToolbar />} sx={formLayoutSx}>
        <FormSection
          title={translate('resources.system/certificate.sections.basic.title', { _: 'Basic Information' })}
          description={translate('resources.system/certificate.sections.basic.description', { _: 'Editable local certificate information' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            <FieldGridItem>
              <TextInput source="id" disabled label={translate('resources.system/certificate.fields.id', { _: 'ID' })} fullWidth size="small" />
            </FieldGridItem>
            <FieldGridItem>
              <TextInput
                source="name"
                label={translate('resources.system/certificate.fields.name', { _: 'Name' })}
                validate={[required(), maxLength(100)]}
                fullWidth
                size="small"
              />
            </FieldGridItem>
            <FieldGridItem span={{ xs: 1, sm: 2 }}>
              <TextInput
                source="remark"
                label={translate('resources.system/certificate.fields.remark', { _: 'Remark' })}
                validate={[maxLength(500)]}
                multiline
                minRows={3}
                fullWidth
                size="small"
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.system/certificate.sections.metadata.title', { _: 'Certificate Metadata' })}
          description={translate('resources.system/certificate.sections.metadata.description', { _: 'Parsed certificate details' })}
        >
          <FieldGrid columns={{ xs: 1, sm: 2 }}>
            {['cert_type', 'subject', 'issuer', 'serial', 'fingerprint', 'not_before', 'not_after', 'has_key'].map(field => (
              <FieldGridItem key={field} span={field === 'fingerprint' || field === 'subject' || field === 'issuer' ? { xs: 1, sm: 2 } : undefined}>
                {field === 'has_key' ? (
                  <BooleanField source="has_key" label={translate('resources.system/certificate.fields.has_key', { _: 'Has Key' })} />
                ) : (
                  <TextInput
                    source={field}
                    label={translate(`resources.system/certificate.fields.${field}`, { _: field })}
                    readOnly
                    fullWidth
                    size="small"
                  />
                )}
              </FieldGridItem>
            ))}
          </FieldGrid>
        </FormSection>

        <FormSection
          title={translate('resources.system/certificate.sections.replace.title', { _: 'Replace Material' })}
          description={translate('resources.system/certificate.sections.replace.description', { _: 'Leave blank to keep the current certificate/key' })}
        >
          <FieldGrid columns={{ xs: 1 }}>
            <FieldGridItem>
              <TextInput
                source="cert"
                label={translate('resources.system/certificate.fields.cert', { _: 'Certificate (PEM)' })}
                multiline
                minRows={8}
                fullWidth
                size="small"
                helperText={translate('resources.system/certificate.helpers.replace_material', { _: 'Leave blank to keep the current certificate/key' })}
                parse={value => value || undefined}
                format={() => ''}
              />
            </FieldGridItem>
            <FieldGridItem>
              <PasswordInput
                source="private_key"
                label={translate('resources.system/certificate.fields.private_key', { _: 'Private Key (PEM)' })}
                multiline
                minRows={8}
                fullWidth
                size="small"
                helperText={translate('resources.system/certificate.helpers.replace_material', { _: 'Leave blank to keep the current certificate/key' })}
                parse={value => value || undefined}
                format={() => ''}
              />
            </FieldGridItem>
          </FieldGrid>
        </FormSection>
      </SimpleForm>
    </Edit>
  );
};

const CertificateShowActions = () => (
  <TopToolbar>
    <ListButton />
    <EditButton />
    <CertificateExportButton />
  </TopToolbar>
);

const CertificateDetails = () => {
  const record = useRecordContext<Certificate>();
  const translate = useTranslate();

  if (!record) return null;

  return (
    <Box sx={{ width: '100%', p: { xs: 2, sm: 3, md: 4 } }}>
      <Stack spacing={3}>
        <Card elevation={0} sx={{ borderRadius: 4, border: theme => `1px solid ${alpha(theme.palette.primary.main, 0.2)}`, overflow: 'hidden' }}>
          <CardContent sx={{ p: 3 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              <Avatar sx={{ width: 64, height: 64, bgcolor: 'primary.main' }}>
                <VerifiedUserIcon sx={{ fontSize: 34 }} />
              </Avatar>
              <Box>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
                  <Typography variant="h5" sx={{ fontWeight: 700 }}>{record.name || <EmptyValue />}</Typography>
                  <CertificateTypeChip />
                </Box>
                <Typography variant="body2" color="text.secondary" sx={{ fontFamily: 'monospace', wordBreak: 'break-all' }}>
                  {record.fingerprint || '-'}
                </Typography>
              </Box>
            </Box>
          </CardContent>
        </Card>

        <DetailSectionCard
          title={translate('resources.system/certificate.sections.metadata.title', { _: 'Certificate Metadata' })}
          description={translate('resources.system/certificate.sections.metadata.description', { _: 'Parsed certificate details' })}
          icon={<BadgeIcon />}
          color="primary"
        >
          <Box sx={gridSx}>
            <DetailItem label={translate('resources.system/certificate.fields.subject', { _: 'Subject' })} value={record.subject || <EmptyValue />} highlight />
            <DetailItem label={translate('resources.system/certificate.fields.issuer', { _: 'Issuer' })} value={record.issuer || <EmptyValue />} />
            <DetailItem label={translate('resources.system/certificate.fields.serial', { _: 'Serial' })} value={record.serial || <EmptyValue />} />
            <DetailItem label={translate('resources.system/certificate.fields.fingerprint', { _: 'Fingerprint' })} value={record.fingerprint || <EmptyValue />} />
            <DetailItem label={translate('resources.system/certificate.fields.cert_type', { _: 'Type' })} value={<CertificateTypeChip />} />
            <DetailItem label={translate('resources.system/certificate.fields.has_key', { _: 'Has Key' })} value={<HasKeyField />} />
          </Box>
        </DetailSectionCard>

        <DetailSectionCard
          title={translate('resources.system/certificate.sections.validity.title', { _: 'Validity' })}
          description={translate('resources.system/certificate.sections.validity.description', { _: 'Certificate validity window' })}
          icon={<CalendarIcon />}
          color="success"
        >
          <Box sx={gridSx}>
            <DetailItem label={translate('resources.system/certificate.fields.not_before', { _: 'Not Before' })} value={formatTimestamp(record.not_before)} />
            <DetailItem label={translate('resources.system/certificate.fields.not_after', { _: 'Not After' })} value={formatTimestamp(record.not_after)} highlight />
            <DetailItem label={translate('resources.system/certificate.fields.created_at', { _: 'Created At' })} value={formatTimestamp(record.created_at)} />
            <DetailItem label={translate('resources.system/certificate.fields.updated_at', { _: 'Updated At' })} value={formatTimestamp(record.updated_at)} />
          </Box>
        </DetailSectionCard>

        <DetailSectionCard
          title={translate('resources.system/certificate.sections.remark.title', { _: 'Remark' })}
          icon={<ArticleIcon />}
          color="info"
        >
          <DetailItem label={translate('resources.system/certificate.fields.remark', { _: 'Remark' })} value={record.remark || <EmptyValue />} />
        </DetailSectionCard>

        <DetailSectionCard
          title={translate('resources.system/certificate.sections.pem.title', { _: 'Certificate PEM' })}
          description={translate('resources.system/certificate.sections.pem.description', { _: 'Public certificate material only' })}
          icon={<KeyIcon />}
          color="warning"
        >
          <Box
            component="pre"
            sx={{
              m: 0,
              p: 2,
              borderRadius: 2,
              bgcolor: theme => (theme.palette.mode === 'dark' ? 'rgba(0,0,0,0.35)' : 'rgba(0,0,0,0.04)'),
              border: theme => `1px solid ${theme.palette.divider}`,
              overflowX: 'auto',
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-word',
              fontFamily: 'monospace',
              fontSize: '0.85rem',
            }}
          >
            {record.cert || ''}
          </Box>
        </DetailSectionCard>
      </Stack>
    </Box>
  );
};

export const CertificateShow = () => (
  <Show actions={<CertificateShowActions />}>
    <CertificateDetails />
  </Show>
);
