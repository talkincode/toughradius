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
} from '@mui/icons-material';
import { useDataProvider, useNotify, useTranslate } from 'react-admin';
import { useApiQuery } from '../hooks/useApiQuery';

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

export const SystemConfigPage: React.FC = () => {
  const [configs, setConfigs] = useState<Record<string, ConfigValue>>({});
  const [expandedGroups, setExpandedGroups] = useState<string[]>(['radius']);
  const [resetDialogOpen, setResetDialogOpen] = useState(false);

  const dataProvider = useDataProvider();
  const notify = useNotify();
  const translate = useTranslate();
  const queryClient = useQueryClient();

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
  }), [translate]);

  const groupedSchemas = useMemo(() => {
    if (!schemaQuery.data) {
      return {} as Record<string, ConfigSchema[]>;
    }
    return schemaQuery.data.reduce<Record<string, ConfigSchema[]>>((groups, schema) => {
      const group = schema.key.split('.')[0];
      if (!groups[group]) {
        groups[group] = [];
      }
      groups[group].push(schema);
      return groups;
    }, {});
  }, [schemaQuery.data]);

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
      </Box>

      {/* 配置分组 */}
      {!isLoading && (schemaQuery.data?.length ?? 0) > 0 && (
        <Box sx={{ mb: 2 }}>
          <Alert severity="info" sx={{ mb: 2 }}>
            {translate('pages.system_config.info_message')}
          </Alert>

          {Object.entries(groupedSchemas).map(([groupKey, groupSchemas]) => {
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
    </Box>
  );
};