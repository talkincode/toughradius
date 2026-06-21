import React, { useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Box,
  Typography,
  TextField,
  Switch,
  FormControl,
  FormControlLabel,
  FormHelperText,
  InputLabel,
  Select,
  MenuItem,
  Button,
  Alert,
  Chip,
  Tooltip,
  IconButton,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
} from '@mui/material';
import {
  Save as SaveIcon,
  Refresh as RefreshIcon,
  ExpandMore as ExpandMoreIcon,
  Info as InfoIcon,
  Settings as SettingsIcon,
  Security as SecurityIcon,
  Router as RouterIcon,
  Backup as BackupIcon,
  RestorePage as RestoreIcon,
  AccountTree as AccountTreeIcon,
} from '@mui/icons-material';
import { useDataProvider, useNotify, useTranslate, useGetList } from 'react-admin';
import { useApiQuery } from '../hooks/useApiQuery';
import { API_BASE } from '../utils/apiClient';

// 配置项类型定义
interface ConfigSchema {
  key: string;
  type: 'string' | 'int' | 'bool' | 'duration' | 'json';
  default: string;
  enum?: string[];
  min?: number;
  max?: number;
  description: string;
  title?: string;
  title_i18n?: string;
  description_i18n?: string;
  group?: string;
}

interface ConfigValue {
  id?: string;
  type: string;
  name: string;
  value: string;
  sort?: number;
  remark?: string;
  updated_at?: string;
}

const SCHEMA_QUERY_KEY = ['system', 'config', 'schemas'] as const;
const SETTINGS_QUERY_KEY = ['system', 'settings'] as const;

// Deterministic display order for config groups. Groups not listed here render
// afterwards in their first-appearance order. Keeps RADIUS general config first
// and the EAP block (plus its collapsed advanced paths) grouped right after it.
const GROUP_ORDER = ['radius', 'eap', 'eap_advanced', 'ldap', 'security', 'system'];

export const SystemConfigPage: React.FC = () => {
  const [configs, setConfigs] = useState<Record<string, ConfigValue>>({});
  const [expandedGroups, setExpandedGroups] = useState<string[]>(['radius', 'eap']);
  const [resetDialogOpen, setResetDialogOpen] = useState(false);
  const [backupLoading, setBackupLoading] = useState(false);
  const [restoreLoading, setRestoreLoading] = useState(false);
  const [restoreDialogOpen, setRestoreDialogOpen] = useState(false);
  const [pendingRestoreFile, setPendingRestoreFile] = useState<File | null>(null);
  const restoreInputRef = React.useRef<HTMLInputElement>(null);

  const dataProvider = useDataProvider();
  const notify = useNotify();
  const translate = useTranslate();
  const queryClient = useQueryClient();

  // Load managed certificates so the EAP block can offer them as a dropdown
  // selection (radius.EapTlsServerCert / radius.EapTlsClientCa) instead of
  // free-text file paths.
  const { data: certificateList } = useGetList('system/certificate', {
    pagination: { page: 1, perPage: 1000 },
    sort: { field: 'name', order: 'ASC' },
  });
  const serverCertOptions = useMemo(
    () => (certificateList ?? []).filter((c) => c.cert_type === 'server').map((c) => String(c.name)),
    [certificateList],
  );
  const caCertOptions = useMemo(
    () => (certificateList ?? []).filter((c) => c.cert_type === 'ca').map((c) => String(c.name)),
    [certificateList],
  );

  const schemaQuery = useApiQuery<ConfigSchema[]>({
    path: '/system/config/schemas',
    queryKey: SCHEMA_QUERY_KEY,
    staleTime: 5 * 60 * 1000,
    retry: 1,
  });

  const settingsQuery = useQuery<ConfigValue[]>({
    queryKey: SETTINGS_QUERY_KEY,
    queryFn: async () => {
      const response = await dataProvider.getList('system/settings', {
        pagination: { page: 1, perPage: 1000 },
        sort: { field: 'type', order: 'ASC' },
        filter: {},
      });
      return response.data as ConfigValue[];
    },
    staleTime: 5 * 60 * 1000,
  });

  useEffect(() => {
    if (!settingsQuery.data) {
      return;
    }
    const map = settingsQuery.data.reduce<Record<string, ConfigValue>>((acc, config) => {
      const key = `${config.type}.${config.name}`;
      acc[key] = config;
      return acc;
    }, {});
    setConfigs(map);
  }, [settingsQuery.data]);

  const configGroups = useMemo(() => ({
    radius: {
      title: translate('pages.system_config.groups.radius.title'),
      description: translate('pages.system_config.groups.radius.description'),
      icon: <RouterIcon />,
      color: '#1976d2',
    },
    system: {
      title: translate('pages.system_config.groups.system.title'),
      description: translate('pages.system_config.groups.system.description'),
      icon: <SettingsIcon />,
      color: '#2e7d32',
    },
    security: {
      title: translate('pages.system_config.groups.security.title'),
      description: translate('pages.system_config.groups.security.description'),
      icon: <SecurityIcon />,
      color: '#d32f2f',
    },
    ldap: {
      title: translate('pages.system_config.groups.ldap.title'),
      description: translate('pages.system_config.groups.ldap.description'),
      icon: <AccountTreeIcon />,
      color: '#7b1fa2',
    },
    eap: {
      title: translate('pages.system_config.groups.eap.title'),
      description: translate('pages.system_config.groups.eap.description'),
      icon: <SecurityIcon />,
      color: '#0288d1',
    },
    eap_advanced: {
      title: translate('pages.system_config.groups.eap_advanced.title'),
      description: translate('pages.system_config.groups.eap_advanced.description'),
      icon: <SecurityIcon />,
      color: '#607d8b',
    },
  }), [translate]);

  const groupedSchemas = useMemo(() => {
    if (!schemaQuery.data) {
      return {} as Record<string, ConfigSchema[]>;
    }
    return schemaQuery.data.reduce<Record<string, ConfigSchema[]>>((groups, schema) => {
      const group = schema.group ?? schema.key.split('.')[0];
      if (!groups[group]) {
        groups[group] = [];
      }
      groups[group].push(schema);
      return groups;
    }, {});
  }, [schemaQuery.data]);

  const orderedGroupEntries = useMemo(() => {
    const entries = Object.entries(groupedSchemas);
    return entries.sort(([a], [b]) => {
      const ia = GROUP_ORDER.indexOf(a);
      const ib = GROUP_ORDER.indexOf(b);
      const ra = ia === -1 ? GROUP_ORDER.length : ia;
      const rb = ib === -1 ? GROUP_ORDER.length : ib;
      return ra - rb;
    });
  }, [groupedSchemas]);

  const resolveSchemaTitle = React.useCallback((schema: ConfigSchema) => {
    const fallback = schema.title || schema.key.split('.')[1];
    if (schema.title_i18n) {
      return translate(schema.title_i18n, { _: fallback });
    }
    return fallback;
  }, [translate]);

  const resolveSchemaDescription = React.useCallback((schema: ConfigSchema) => {
    const fallback = schema.description;
    if (schema.description_i18n) {
      return translate(schema.description_i18n, { _: fallback });
    }
    return fallback;
  }, [translate]);

  const isLoading = schemaQuery.isLoading || settingsQuery.isLoading;

  const updateConfigValue = (key: string, value: string) => {
    setConfigs(prev => ({
      ...prev,
      [key]: {
        ...prev[key],
        type: key.split('.')[0],
        name: key.split('.')[1],
        value,
      },
    }));
  };

  const getConfigValue = (schema: ConfigSchema): string => configs[schema.key]?.value ?? schema.default;

  const renderConfigInput = (schema: ConfigSchema, label: string, description: string) => {
    const value = getConfigValue(schema);

    // EAP certificate references are selected from the managed certificate store
    // (Certificates page) rather than typed as file paths, satisfying the
    // "select a certificate, don't write a path" requirement.
    if (schema.key === 'radius.EapTlsServerCert' || schema.key === 'radius.EapTlsClientCa') {
      const options = schema.key === 'radius.EapTlsServerCert' ? serverCertOptions : caCertOptions;
      const allOptions = value && !options.includes(value) ? [value, ...options] : options;
      return (
        <FormControl fullWidth>
          <InputLabel>{label}</InputLabel>
          <Select
            value={value}
            label={label}
            onChange={(e) => updateConfigValue(schema.key, e.target.value)}
          >
            <MenuItem value="">
              <em>{translate('pages.system_config.cert_select_none', { _: '未选择 (None)' })}</em>
            </MenuItem>
            {allOptions.map((name) => (
              <MenuItem key={name} value={name}>
                {name}
              </MenuItem>
            ))}
          </Select>
          {description && <FormHelperText>{description}</FormHelperText>}
        </FormControl>
      );
    }

    switch (schema.type) {
      case 'bool':
        return (
          <Box>
            <FormControlLabel
              control={
                <Switch
                  checked={value === 'true'}
                  onChange={(e) => updateConfigValue(schema.key, e.target.checked ? 'true' : 'false')}
                />
              }
              label={label}
            />
            {description && (
              <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mt: 0.5 }}>
                {description}
              </Typography>
            )}
          </Box>
        );
      case 'string':
        if (schema.enum) {
          return (
            <FormControl fullWidth>
              <InputLabel>{label}</InputLabel>
              <Select
                value={value}
                label={label}
                onChange={(e) => updateConfigValue(schema.key, e.target.value)}
              >
                {schema.enum.map(option => (
                  <MenuItem key={option} value={option}>
                    {option}
                  </MenuItem>
                ))}
              </Select>
              {description && (
                <FormHelperText>{description}</FormHelperText>
              )}
            </FormControl>
          );
        }
        return (
          <TextField
            fullWidth
            label={label}
            value={value}
            onChange={(e) => updateConfigValue(schema.key, e.target.value)}
            helperText={description || undefined}
          />
        );
      case 'int':
        return (
          <TextField
            fullWidth
            type="number"
            label={label}
            value={value}
            onChange={(e) => updateConfigValue(schema.key, e.target.value)}
            helperText={description || undefined}
            InputProps={{
              inputProps: {
                min: schema.min,
                max: schema.max,
              },
            }}
          />
        );
      default:
        return (
          <TextField
            fullWidth
            label={label}
            value={value}
            onChange={(e) => updateConfigValue(schema.key, e.target.value)}
            helperText={description || undefined}
          />
        );
    }
  };

  const saveMutation = useMutation({
    mutationFn: async (draft: Record<string, ConfigValue>) => {
      const schemaList = schemaQuery.data ?? [];
      await Promise.all(
        schemaList.map(async schema => {
          const [type, name] = schema.key.split('.');
          const currentConfig = draft[schema.key];
          const payload = {
            type,
            name,
            value: currentConfig?.value ?? schema.default,
            sort: currentConfig?.sort ?? 0,
            remark: currentConfig?.remark ?? schema.description,
          };

          if (currentConfig?.id) {
            await dataProvider.update('system/settings', {
              id: currentConfig.id,
              data: payload,
              previousData: currentConfig,
            });
            return;
          }

          const created = await dataProvider.create('system/settings', {
            data: payload,
          });
          const createdData = created.data as ConfigValue;
          draft[schema.key] = {
            ...payload,
            id: createdData?.id,
            updated_at: createdData?.updated_at,
          };
        }),
      );
    },
    onSuccess: () => {
      notify('配置保存成功', { type: 'success' });
      queryClient.invalidateQueries({ queryKey: SETTINGS_QUERY_KEY });
    },
    onError: (error: unknown) => {
      const message = error instanceof Error ? error.message : 'unknown error';
      notify(`配置保存失败: ${message}`, { type: 'error' });
    },
  });

  const handleResetConfigs = () => {
    if (!schemaQuery.data?.length) {
      setResetDialogOpen(false);
      return;
    }
    const nextConfigs = { ...configs };
    schemaQuery.data.forEach(schema => {
      nextConfigs[schema.key] = {
        ...nextConfigs[schema.key],
        type: schema.key.split('.')[0],
        name: schema.key.split('.')[1],
        value: schema.default,
      };
    });
    setConfigs(nextConfigs);
    setResetDialogOpen(false);
    notify('已重置为默认值', { type: 'info' });
  };

  const handleGroupToggle = (group: string) => {
    setExpandedGroups(prev =>
      prev.includes(group) ? prev.filter(g => g !== group) : [...prev, group],
    );
  };

  const handleReload = () => {
    queryClient.invalidateQueries({ queryKey: SCHEMA_QUERY_KEY });
    queryClient.invalidateQueries({ queryKey: SETTINGS_QUERY_KEY });
  };

  const handleBackup = async () => {
    setBackupLoading(true);
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${API_BASE}/system/backup`, {
        method: 'GET',
        headers: token ? { Authorization: 'Bearer ' + token } : undefined,
      });
      if (!response.ok) {
        const payload = await response.json().catch(() => ({}));
        throw new Error(payload?.message || translate('pages.system_config.backup.failed', { _: '备份失败' }));
      }
      const blob = await response.blob();
      const disposition = response.headers.get('Content-Disposition') || '';
      const match = disposition.match(/filename=([^;]+)/);
      const filename = match ? match[1].trim() : `toughradius-backup-${Date.now()}.json`;
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
      notify(translate('pages.system_config.backup.success', { _: '备份成功' }), { type: 'success' });
      notify(
        translate('pages.system_config.backup.notice', { _: '备份文件包含明文密码与凭据，请务必妥善保管。' }),
        { type: 'info', autoHideDuration: 8000 }
      );
    } catch (error) {
      notify((error as Error).message, { type: 'error' });
    } finally {
      setBackupLoading(false);
    }
  };

  // 选择恢复文件后先弹确认框，避免误覆盖现有配置（含管理员账号/密码）
  const handleRestoreSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0] ?? null;
    if (restoreInputRef.current) {
      restoreInputRef.current.value = '';
    }
    if (!file) {
      return;
    }
    setPendingRestoreFile(file);
    setRestoreDialogOpen(true);
  };

  const handleRestoreConfirm = () => {
    const file = pendingRestoreFile;
    setRestoreDialogOpen(false);
    setPendingRestoreFile(null);
    if (file) {
      void doRestore(file);
    }
  };

  const handleRestoreCancel = () => {
    setRestoreDialogOpen(false);
    setPendingRestoreFile(null);
  };

  const doRestore = async (file: File) => {
    setRestoreLoading(true);
    try {
      const formData = new FormData();
      formData.append('upload', file);
      const token = localStorage.getItem('token');
      const response = await fetch(`${API_BASE}/system/restore`, {
        method: 'POST',
        headers: token ? { Authorization: 'Bearer ' + token } : undefined,
        body: formData,
      });
      const payload = await response.json().catch(() => ({}));
      if (!response.ok) {
        throw new Error(payload?.message || translate('pages.system_config.restore.failed', { _: '恢复失败' }));
      }
      notify(translate('pages.system_config.restore.success', { _: '恢复成功' }), { type: 'success' });
      handleReload();
    } catch (error) {
      notify((error as Error).message, { type: 'error' });
    } finally {
      setRestoreLoading(false);
    }
  };

  const handleSave = () => {
    if (!schemaQuery.data?.length) {
      notify('暂无可保存的配置项', { type: 'warning' });
      return;
    }
    saveMutation.mutate({ ...configs });
  };

  return (
    <Box sx={{ p: 3 }}>
      {/* 页面标题 */}
      <Box sx={{ mb: 3 }}>
        <Typography variant="h4" gutterBottom>
          {translate('pages.system_config.title')}
        </Typography>
        <Typography variant="body1" color="textSecondary">
          {translate('pages.system_config.subtitle')}
        </Typography>
      </Box>

      {/* 操作按钮 */}
      <Box sx={{ mb: 3 }}>
        <Button
          variant="contained"
          startIcon={<SaveIcon />}
          onClick={handleSave}
          disabled={saveMutation.isPending || isLoading}
          sx={{ mr: 2 }}
        >
          {saveMutation.isPending ? translate('pages.system_config.saving') : translate('pages.system_config.save')}
        </Button>
        <Button
          variant="outlined"
          startIcon={<RefreshIcon />}
          onClick={() => setResetDialogOpen(true)}
          disabled={saveMutation.isPending || isLoading}
          sx={{ mr: 2 }}
        >
          {translate('pages.system_config.reset')}
        </Button>
        <Button
          variant="text"
          startIcon={<RefreshIcon />}
          onClick={handleReload}
          disabled={isLoading}
        >
          {isLoading ? translate('pages.system_config.loading') : translate('pages.system_config.reload')}
        </Button>
        <Button
          variant="outlined"
          startIcon={<BackupIcon />}
          onClick={handleBackup}
          disabled={backupLoading}
          sx={{ ml: 2, mr: 2 }}
        >
          {translate('pages.system_config.backup.button', { _: '系统备份' })}
        </Button>
        <Button
          variant="outlined"
          startIcon={<RestoreIcon />}
          onClick={() => restoreInputRef.current?.click()}
          disabled={restoreLoading}
        >
          {translate('pages.system_config.restore.button', { _: '系统恢复' })}
        </Button>
        <input
          ref={restoreInputRef}
          type="file"
          accept=".json"
          style={{ display: 'none' }}
          onChange={handleRestoreSelect}
        />
      </Box>

      {/* 配置分组 */}
      {!isLoading && (schemaQuery.data?.length ?? 0) > 0 && (
        <Box sx={{ mb: 2 }}>
          <Alert severity="info" sx={{ mb: 2 }}>
            {translate('pages.system_config.info_message')}
          </Alert>

          {orderedGroupEntries.map(([groupKey, groupSchemas]) => {
          const groupConfig = configGroups[groupKey as keyof typeof configGroups] || {
            title: groupKey,
            description: `${groupKey} 相关配置`,
            icon: <SettingsIcon />,
            color: '#666',
          };

          const isExpanded = expandedGroups.includes(groupKey);

          return (
            <Accordion 
              key={groupKey} 
              expanded={isExpanded}
              onChange={() => handleGroupToggle(groupKey)}
              sx={{ mb: 2 }}
            >
              <AccordionSummary 
                expandIcon={<ExpandMoreIcon />}
                sx={{ 
                  backgroundColor: `${groupConfig.color}15`,
                  '&:hover': { backgroundColor: `${groupConfig.color}25` }
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                  <Box sx={{ color: groupConfig.color }}>
                    {groupConfig.icon}
                  </Box>
                  <Box>
                    <Typography variant="h6" sx={{ color: groupConfig.color }}>
                      {groupConfig.title}
                    </Typography>
                    <Typography variant="body2" color="textSecondary">
                      {groupConfig.description} ({groupSchemas.length} {translate('pages.system_config.config_items')})
                    </Typography>
                  </Box>
                </Box>
              </AccordionSummary>
              
              <AccordionDetails>
                <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))', gap: 3 }}>
                  {groupSchemas.map((schema) => {
                    const schemaTitle = resolveSchemaTitle(schema);
                    const schemaDescription = resolveSchemaDescription(schema);
                    return (
                      <Box key={schema.key} sx={{ mb: 2 }}>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                          <Typography variant="subtitle2">
                            {schemaTitle}
                          </Typography>
                          <Tooltip title={schemaDescription}>
                            <IconButton size="small">
                              <InfoIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                          <Chip 
                            label={schema.type} 
                            size="small" 
                            variant="outlined"
                            sx={{ ml: 'auto' }}
                          />
                        </Box>
                        {renderConfigInput(schema, schemaTitle, schemaDescription)}
                        {schema.enum && (
                          <Typography variant="caption" color="textSecondary" sx={{ mt: 0.5, display: 'block' }}>
                            {translate('pages.system_config.available_values')}: {schema.enum.join(', ')}
                          </Typography>
                        )}
                        {(schema.min !== undefined || schema.max !== undefined) && (
                          <Typography variant="caption" color="textSecondary" sx={{ mt: 0.5, display: 'block' }}>
                            {translate('pages.system_config.value_range')}: {schema.min !== undefined ? `${translate('pages.system_config.min')} ${schema.min}` : ''} 
                            {schema.min !== undefined && schema.max !== undefined ? ', ' : ''}
                            {schema.max !== undefined ? `${translate('pages.system_config.max')} ${schema.max}` : ''}
                          </Typography>
                        )}
                      </Box>
                    );
                  })}
                </Box>
              </AccordionDetails>
            </Accordion>
          );
        })}
        </Box>
      )}

      {isLoading && (
        <Alert severity="info">
          {translate('pages.system_config.loading_message')}
          <br />
          {translate('pages.system_config.loading_detail')}
        </Alert>
      )}

      {!isLoading && (schemaQuery.data?.length ?? 0) === 0 && (
        <Alert severity="warning">
          {translate('pages.system_config.no_config_warning')}
          <br />
          <strong>调试信息：</strong>
          <br />
          - 配置schema数量：{schemaQuery.data?.length ?? 0}
          <br />
          - 配置值数量：{Object.keys(configs).length}
          <br />
          - API地址：/api/v1/system/config/schemas
          <br />
          请打开浏览器控制台查看详细日志。
        </Alert>
      )}

      {!isLoading && (schemaQuery.data?.length ?? 0) > 0 && (
        <Alert severity="success" sx={{ mb: 2 }}>
          ✓ {translate('pages.system_config.success_message', { schemaCount: schemaQuery.data?.length ?? 0, configCount: Object.keys(configs).length })}
        </Alert>
      )}

      {/* 重置确认对话框 */}
      <Dialog
      open={resetDialogOpen}
      onClose={() => setResetDialogOpen(false)}
      aria-labelledby="reset-dialog-title"
      aria-describedby="reset-dialog-description"
    >
      <DialogTitle id="reset-dialog-title">
        {translate('pages.system_config.confirm_reset')}
      </DialogTitle>
      <DialogContent>
        <DialogContentText id="reset-dialog-description">
          {translate('pages.system_config.reset_warning')}
          <br />
          <br />
          <strong>注意：</strong>此操作将清除您对以下配置项的自定义设置：
          <br />
          {schemaQuery.data?.map(schema => (
            <span key={schema.key}>
              • {schema.key.split('.')[1]} ({schema.description})
              <br />
            </span>
          ))}
          <br />
          {translate('pages.system_config.reset_notice')}
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button onClick={() => setResetDialogOpen(false)}>
          {translate('pages.system_config.cancel')}
        </Button>
        <Button onClick={handleResetConfigs} color="warning" variant="contained">
          {translate('pages.system_config.confirm')}
        </Button>
      </DialogActions>
    </Dialog>

      {/* 系统恢复确认对话框 */}
      <Dialog
        open={restoreDialogOpen}
        onClose={handleRestoreCancel}
        aria-labelledby="restore-dialog-title"
        aria-describedby="restore-dialog-description"
      >
        <DialogTitle id="restore-dialog-title">
          {translate('pages.system_config.restore.confirm_title', { _: '确认系统恢复' })}
        </DialogTitle>
        <DialogContent>
          <DialogContentText id="restore-dialog-description">
            {translate('pages.system_config.restore.confirm_warning', {
              _: '系统恢复将用备份文件中的数据覆盖现有配置（节点、NAS、套餐、用户、系统配置、操作员）。这会覆盖当前管理员账号及密码，恢复后可能需要使用备份中的凭据重新登录。确定要继续吗？',
            })}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleRestoreCancel}>
            {translate('pages.system_config.restore.cancel', { _: '取消' })}
          </Button>
          <Button onClick={handleRestoreConfirm} color="warning" variant="contained" disabled={restoreLoading}>
            {translate('pages.system_config.restore.confirm', { _: '确认恢复' })}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};