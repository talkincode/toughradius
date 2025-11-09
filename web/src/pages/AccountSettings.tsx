import { useEffect, useState } from 'react';
import { useGetIdentity, useNotify, Loading } from 'react-admin';
import {
  Card,
  CardContent,
  Typography,
  Box,
  TextField,
  Button,
  CircularProgress,
} from '@mui/material';

// 表单区域组件
const FormSection = ({ title, children }: { title: string; children: React.ReactNode }) => (
  <Box sx={{ mb: 3 }}>
    <Typography variant="subtitle1" sx={{ mb: 2, fontWeight: 600, color: 'primary.main' }}>
      {title}
    </Typography>
    {children}
  </Box>
);

interface FormData {
  username: string;
  password: string;
  realname: string;
  mobile: string;
  email: string;
  remark: string;
}

export const AccountSettings = () => {
  const { data: identity, isLoading: identityLoading } = useGetIdentity();
  const notify = useNotify();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [formData, setFormData] = useState<FormData>({
    username: '',
    password: '',
    realname: '',
    mobile: '',
    email: '',
    remark: '',
  });
  const [currentLevel, setCurrentLevel] = useState('');
  const [currentStatus, setCurrentStatus] = useState('');

  useEffect(() => {
    if (!identity?.id) return;

    // 使用专门的 /me 接口获取当前用户数据
    const fetchCurrentUser = async () => {
      try {
        const token = localStorage.getItem('token');
        const response = await fetch('/api/v1/system/operators/me', {
          headers: {
            Authorization: `Bearer ${token}`,
            Accept: 'application/json',
          },
        });

        if (!response.ok) {
          throw new Error('获取用户信息失败');
        }

        const result = await response.json();
        const data = result.data || result;

        setFormData({
          username: data.username || '',
          password: '', // 密码始终为空
          realname: data.realname || '',
          mobile: data.mobile || '',
          email: data.email || '',
          remark: data.remark || '',
        });
        setCurrentLevel(data.level || '');
        setCurrentStatus(data.status || '');
        setLoading(false);
      } catch (error) {
        notify('加载账号信息失败', { type: 'error' });
        console.error('Failed to load operator data:', error);
        setLoading(false);
      }
    };

    fetchCurrentUser();
  }, [identity?.id, notify]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);

    try {
      const token = localStorage.getItem('token');

      // 只提交修改的字段
      const updateData: Record<string, string> = {
        username: formData.username,
        realname: formData.realname,
      };

      if (formData.password) {
        updateData.password = formData.password;
      }
      if (formData.mobile) {
        updateData.mobile = formData.mobile;
      }
      if (formData.email) {
        updateData.email = formData.email;
      }
      if (formData.remark) {
        updateData.remark = formData.remark;
      }

      const response = await fetch('/api/v1/system/operators/me', {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
          Accept: 'application/json',
        },
        body: JSON.stringify(updateData),
      });

      if (!response.ok) {
        const result = await response.json();
        throw new Error(result.message || '更新失败');
      }

      notify('账号信息更新成功', { type: 'success' });

      // 清空密码字段
      setFormData((prev) => ({ ...prev, password: '' }));
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '账号信息更新失败';
      notify(errorMessage, { type: 'error' });
      console.error('Failed to update operator data:', error);
    } finally {
      setSaving(false);
    }
  };

  if (identityLoading || loading) {
    return <Loading />;
  }

  const getLevelLabel = (level: string) => {
    switch (level) {
      case 'super':
        return '超级管理员';
      case 'admin':
        return '管理员';
      case 'operator':
        return '普通操作员';
      default:
        return level;
    }
  };

  const getStatusLabel = (status: string) => {
    return status === 'enabled' ? '启用' : '禁用';
  };

  return (
    <Box sx={{ maxWidth: 900, mx: 'auto', mt: 2 }}>
      <Card>
        <CardContent>
          <Typography variant="h5" gutterBottom sx={{ mb: 3, fontWeight: 600 }}>
            账号设置
          </Typography>

          <form onSubmit={handleSubmit}>
            <FormSection title="登录信息">
              <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2 }}>
                <TextField
                  fullWidth
                  label="用户名"
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  required
                  helperText="3-30个字符，只能包含字母、数字和下划线"
                />
                <TextField
                  fullWidth
                  label="新密码"
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  helperText="留空表示不修改密码。6-50个字符，必须包含字母和数字"
                />
              </Box>
            </FormSection>

            <FormSection title="个人信息">
              <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, gap: 2 }}>
                <TextField
                  fullWidth
                  label="真实姓名"
                  value={formData.realname}
                  onChange={(e) => setFormData({ ...formData, realname: e.target.value })}
                  required
                />
                <TextField
                  fullWidth
                  label="手机号"
                  value={formData.mobile}
                  onChange={(e) => setFormData({ ...formData, mobile: e.target.value })}
                  helperText="请输入11位中国大陆手机号"
                />
                <TextField
                  fullWidth
                  label="邮箱"
                  type="email"
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  helperText="用于接收通知和找回密码"
                />
              </Box>
            </FormSection>

            <FormSection title="备注说明">
              <TextField
                fullWidth
                label="备注"
                multiline
                rows={3}
                value={formData.remark}
                onChange={(e) => setFormData({ ...formData, remark: e.target.value })}
              />
            </FormSection>

            {/* 只读信息 */}
            <Box
              sx={{
                mt: 2,
                p: 2,
                bgcolor: 'background.paper',
                borderRadius: 1,
                border: '1px solid',
                borderColor: 'divider',
              }}
            >
              <Typography variant="subtitle2" sx={{ mb: 1, color: 'text.secondary' }}>
                当前权限信息（不可修改）
              </Typography>
              <Box sx={{ display: 'flex', gap: 3 }}>
                <Typography variant="body2">
                  <strong>权限级别：</strong>
                  <span
                    style={{
                      color:
                        currentLevel === 'super'
                          ? '#d32f2f'
                          : currentLevel === 'admin'
                          ? '#ed6c02'
                          : '#0288d1',
                    }}
                  >
                    {getLevelLabel(currentLevel)}
                  </span>
                </Typography>
                <Typography variant="body2">
                  <strong>账号状态：</strong>
                  <span style={{ color: currentStatus === 'enabled' ? '#2e7d32' : '#757575' }}>
                    {getStatusLabel(currentStatus)}
                  </span>
                </Typography>
              </Box>
            </Box>

            {/* 提交按钮 */}
            <Box sx={{ mt: 3, display: 'flex', justifyContent: 'flex-end' }}>
              <Button
                type="submit"
                variant="contained"
                disabled={saving}
                startIcon={saving ? <CircularProgress size={20} /> : null}
              >
                {saving ? '保存中...' : '保存'}
              </Button>
            </Box>
          </form>
        </CardContent>
      </Card>
    </Box>
  );
};
