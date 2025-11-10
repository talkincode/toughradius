import React, { useState, useEffect, useCallback } from 'react';
import {
  Box,
  Typography,
  TextField,
  Switch,
  FormControl,
  FormControlLabel,
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
import { useDataProvider, useNotify } from 'react-admin';

// 配置项类型定义
interface ConfigSchema {
  key: string;
  type: 'string' | 'int' | 'bool' | 'duration' | 'json';
  default: string;
  enum?: string[];
  min?: number;
  max?: number;
  description: string;
}

interface ConfigValue {
  type: string;
  name: string;
  value: string;
  updated_at?: string;
}

interface ApiSchemaData {
  key: string;
  type: string;
  default: string;
  enum?: string[];
  min?: number;
  max?: number;
  description: string;
}

// 配置分组定义
const CONFIG_GROUPS = {
  radius: {
    title: 'RADIUS 配置',
    description: 'RADIUS 认证和计费相关配置',
    icon: <RouterIcon />,
    color: '#1976d2',
  },
  system: {
    title: '系统配置',
    description: '系统基础设置和维护配置',
    icon: <SettingsIcon />,
    color: '#2e7d32',
  },
  security: {
    title: '安全配置',
    description: '安全策略和认证相关配置',
    icon: <SecurityIcon />,
    color: '#d32f2f',
  },
};

export const SystemConfigPage: React.FC = () => {
  const [configs, setConfigs] = useState<Record<string, ConfigValue>>({});
  const [schemas, setSchemas] = useState<ConfigSchema[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [expandedGroups, setExpandedGroups] = useState<string[]>(['radius']);
  const [resetDialogOpen, setResetDialogOpen] = useState(false);
  
  const dataProvider = useDataProvider();
  const notify = useNotify();

  // 加载配置数据
  const loadConfigs = useCallback(async () => {
    setLoading(true);
    try {
      console.log('开始加载配置...');
      
      // 1. 首先直接调用API加载配置 schema
      console.log('开始加载配置 schemas...');
      const token = localStorage.getItem('token');
      
      if (!token) {
        throw new Error('未找到认证token，请重新登录');
      }
      
      const schemaResponse = await fetch('/api/v1/system/config/schemas', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!schemaResponse.ok) {
        throw new Error(`Schema API 调用失败: ${schemaResponse.status} ${schemaResponse.statusText}`);
      }

      const schemaJson = await schemaResponse.json();
      console.log('Schema 响应:', schemaJson);
      
      // 提取数据 - 根据后端API格式调整
      const schemaData = schemaJson.data || schemaJson;
      console.log('Schema 数据:', schemaData);

      if (!Array.isArray(schemaData) || schemaData.length === 0) {
        throw new Error('没有找到配置定义数据');
      }

      const loadedSchemas: ConfigSchema[] = schemaData.map((schema: ApiSchemaData) => ({
        key: schema.key,
        type: schema.type as ConfigSchema['type'],
        default: schema.default,
        enum: schema.enum,
        min: schema.min,
        max: schema.max,
        description: schema.description,
      }));

      setSchemas(loadedSchemas);
      console.log('处理后的 schemas:', loadedSchemas);

      // 2. 然后加载配置值（使用dataProvider）
      const configResponse = await dataProvider.getList('system/settings', {
        pagination: { page: 1, perPage: 1000 },
        sort: { field: 'type', order: 'ASC' },
        filter: {},
      });

      console.log('配置响应:', configResponse);
      const configData = configResponse.data;
      console.log('配置数据:', configData);

      const configMap: Record<string, ConfigValue> = {};
      if (Array.isArray(configData)) {
        configData.forEach((config: ConfigValue) => {
          const key = `${config.type}.${config.name}`;
          configMap[key] = config;
        });
      }

      setConfigs(configMap);
      console.log('处理后的配置映射:', configMap);

    } catch (error) {
      console.error('加载配置失败:', error);
      notify('加载配置失败: ' + (error as Error).message, { type: 'error' });
      
      // 如果加载失败，显示错误信息但不使用硬编码配置
      setSchemas([]);
      setConfigs({});
    } finally {
      setLoading(false);
    }
  }, [dataProvider, notify]);

  // 保存配置
  const saveConfigs = async () => {
    setSaving(true);
    try {
      const updates = [];

      for (const schema of schemas) {
        const [type, name] = schema.key.split('.');
        const currentConfig = configs[schema.key];
        const currentValue = currentConfig?.value || schema.default;

        updates.push({
          type,
          name,
          value: currentValue,
        });
      }

      // 批量更新配置
      await Promise.all(
        updates.map(async (update) => {
          const existing = configs[`${update.type}.${update.name}`];
          
          if (existing) {
            // 更新现有配置
            return dataProvider.update('system/settings', {
              id: existing.type + '.' + existing.name,
              data: update,
              previousData: existing,
            });
          } else {
            // 创建新配置
            return dataProvider.create('system/settings', {
              data: update,
            });
          }
        })
      );

      notify('配置保存成功', { type: 'success' });
      await loadConfigs();
    } catch (error) {
      console.error('Save configs failed:', error);
      notify('配置保存失败', { type: 'error' });
    } finally {
      setSaving(false);
    }
  };

  // 重置配置
  const handleResetConfigs = () => {
    const resetConfigs = { ...configs };
    schemas.forEach(schema => {
      const config = resetConfigs[schema.key];
      if (config) {
        config.value = schema.default;
      }
    });
    setConfigs(resetConfigs);
    setResetDialogOpen(false);
    notify('已重置为默认值', { type: 'info' });
  };

  // 更新配置值
  const updateConfigValue = (key: string, value: string) => {
    setConfigs(prev => ({
      ...prev,
      [key]: {
        ...prev[key],
        type: key.split('.')[0],
        name: key.split('.')[1],
        value: value,
      }
    }));
  };

  // 获取配置值
  const getConfigValue = (schema: ConfigSchema): string => {
    return configs[schema.key]?.value || schema.default;
  };

  // 渲染配置输入组件
  const renderConfigInput = (schema: ConfigSchema) => {
    const value = getConfigValue(schema);

    switch (schema.type) {
      case 'bool':
        return (
          <FormControlLabel
            control={
              <Switch
                checked={value === 'true'}
                onChange={(e) => updateConfigValue(schema.key, e.target.checked ? 'true' : 'false')}
              />
            }
            label={schema.description}
          />
        );

      case 'string':
        if (schema.enum) {
          return (
            <FormControl fullWidth>
              <InputLabel>{schema.key.split('.')[1]}</InputLabel>
              <Select
                value={value}
                label={schema.key.split('.')[1]}
                onChange={(e) => updateConfigValue(schema.key, e.target.value)}
              >
                {schema.enum.map(option => (
                  <MenuItem key={option} value={option}>{option}</MenuItem>
                ))}
              </Select>
            </FormControl>
          );
        }
        return (
          <TextField
            fullWidth
            label={schema.key.split('.')[1]}
            value={value}
            onChange={(e) => updateConfigValue(schema.key, e.target.value)}
            helperText={schema.description}
          />
        );

      case 'int':
        return (
          <TextField
            fullWidth
            type="number"
            label={schema.key.split('.')[1]}
            value={value}
            onChange={(e) => updateConfigValue(schema.key, e.target.value)}
            helperText={schema.description}
            InputProps={{
              inputProps: {
                min: schema.min,
                max: schema.max,
              }
            }}
          />
        );

      default:
        return (
          <TextField
            fullWidth
            label={schema.key.split('.')[1]}
            value={value}
            onChange={(e) => updateConfigValue(schema.key, e.target.value)}
            helperText={schema.description}
          />
        );
    }
  };

  // 分组配置项
  const groupedSchemas = schemas.reduce((groups, schema) => {
    const group = schema.key.split('.')[0];
    if (!groups[group]) {
      groups[group] = [];
    }
    groups[group].push(schema);
    return groups;
  }, {} as Record<string, ConfigSchema[]>);

  console.log('分组后的配置:', groupedSchemas);

  useEffect(() => {
    loadConfigs();
  }, [loadConfigs]);

  const handleGroupToggle = (group: string) => {
    setExpandedGroups(prev => 
      prev.includes(group) 
        ? prev.filter(g => g !== group)
        : [...prev, group]
    );
  };

  return (
    <Box sx={{ p: 3 }}>
      {/* 页面标题 */}
      <Box sx={{ mb: 3 }}>
        <Typography variant="h4" gutterBottom>
          系统配置
        </Typography>
        <Typography variant="body1" color="textSecondary">
          管理系统的各项配置参数，配置修改后立即生效
        </Typography>
      </Box>

      {/* 操作按钮 */}
      <Box sx={{ mb: 3 }}>
        <Button
          variant="contained"
          startIcon={<SaveIcon />}
          onClick={saveConfigs}
          disabled={saving || loading}
          sx={{ mr: 2 }}
        >
          {saving ? '保存中...' : '保存配置'}
        </Button>
        <Button
          variant="outlined"
          startIcon={<RefreshIcon />}
          onClick={() => setResetDialogOpen(true)}
          disabled={saving || loading}
          sx={{ mr: 2 }}
        >
          重置默认值
        </Button>
        <Button
          variant="text"
          startIcon={<RefreshIcon />}
          onClick={loadConfigs}
          disabled={loading}
        >
          {loading ? '加载中...' : '重新加载'}
        </Button>
      </Box>

      {/* 配置分组 */}
      {!loading && schemas.length > 0 && (
        <Box sx={{ mb: 2 }}>
          <Alert severity="info" sx={{ mb: 2 }}>
            配置项按功能模块分组显示，点击展开/收起分组。修改配置后请点击"保存配置"按钮。
          </Alert>

          {Object.entries(groupedSchemas).map(([groupKey, groupSchemas]) => {
          const groupConfig = CONFIG_GROUPS[groupKey as keyof typeof CONFIG_GROUPS] || {
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
                      {groupConfig.description} ({groupSchemas.length} 项配置)
                    </Typography>
                  </Box>
                </Box>
              </AccordionSummary>
              
              <AccordionDetails>
                <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))', gap: 3 }}>
                  {groupSchemas.map((schema) => (
                    <Box key={schema.key} sx={{ mb: 2 }}>
                      <Box sx={{ mb: 2 }}>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                          <Typography variant="subtitle2">
                            {schema.key.split('.')[1]}
                          </Typography>
                          <Tooltip title={schema.description}>
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
                        {renderConfigInput(schema)}
                        {schema.enum && (
                          <Typography variant="caption" color="textSecondary" sx={{ mt: 0.5, display: 'block' }}>
                            可选值: {schema.enum.join(', ')}
                          </Typography>
                        )}
                        {(schema.min !== undefined || schema.max !== undefined) && (
                          <Typography variant="caption" color="textSecondary" sx={{ mt: 0.5, display: 'block' }}>
                            范围: {schema.min !== undefined ? `最小 ${schema.min}` : ''} 
                            {schema.min !== undefined && schema.max !== undefined ? ', ' : ''}
                            {schema.max !== undefined ? `最大 ${schema.max}` : ''}
                          </Typography>
                        )}
                      </Box>
                    </Box>
                  ))}
                </Box>
              </AccordionDetails>
            </Accordion>
          );
        })}
        </Box>
      )}

      {loading && (
        <Alert severity="info">
          加载配置中，请稍候...
          <br />
          正在从后端API获取配置定义和当前配置值。
        </Alert>
      )}

      {!loading && schemas.length === 0 && (
        <Alert severity="warning">
          没有找到配置项，请检查后端API是否正常工作。
          <br />
          <strong>调试信息：</strong>
          <br />
          - 配置schema数量：{schemas.length}
          <br />
          - 配置值数量：{Object.keys(configs).length}
          <br />
          - API地址：/api/v1/system/config/schemas
          <br />
          请打开浏览器控制台查看详细日志。
        </Alert>
      )}

      {!loading && schemas.length > 0 && (
        <Alert severity="success" sx={{ mb: 2 }}>
          ✓ 成功加载 {schemas.length} 个配置定义，{Object.keys(configs).length} 个配置值
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
        确认重置配置
      </DialogTitle>
      <DialogContent>
        <DialogContentText id="reset-dialog-description">
          确定要将所有配置项重置为默认值吗？
          <br />
          <br />
          <strong>注意：</strong>此操作将清除您对以下配置项的自定义设置：
          <br />
          {schemas.map(schema => (
            <span key={schema.key}>
              • {schema.key.split('.')[1]} ({schema.description})
              <br />
            </span>
          ))}
          <br />
          重置后需要点击"保存配置"才会生效。
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button onClick={() => setResetDialogOpen(false)}>
          取消
        </Button>
        <Button onClick={handleResetConfigs} color="warning" variant="contained">
          确认重置
        </Button>
      </DialogActions>
    </Dialog>
    </Box>
  );
};