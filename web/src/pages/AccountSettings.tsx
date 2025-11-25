import { useEffect, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useNotify, useTranslate } from 'react-admin';
import {
  Card,
  CardContent,
  Typography,
  Box,
  TextField,
  Button,
  CircularProgress,
  Alert,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material';
import {
  Person as PersonIcon,
  Lock as LockIcon,
  ExpandMore as ExpandMoreIcon,
  Info as InfoIcon,
} from '@mui/icons-material';
import { clearAuthStorage } from '../utils/storage';
import { useApiQuery } from '../hooks/useApiQuery';
import { useApiMutation } from '../hooks/useApiMutation';
import { ApiError } from '../utils/apiClient';

interface OperatorProfile {
  username?: string;
  realname?: string;
  mobile?: string;
  email?: string;
  remark?: string;
  level?: string;
  status?: string;
}

const ACCOUNT_QUERY_KEY = ['operators', 'me'] as const;

export default function AccountSettings() {
  const notify = useNotify();
  const translate = useTranslate();
  const queryClient = useQueryClient();
  const [expandedGroups, setExpandedGroups] = useState<string[]>(['profile']);

  const [profileForm, setProfileForm] = useState({
    username: '',
    realname: '',
    mobile: '',
    email: '',
    remark: '',
  });

  const [passwordForm, setPasswordForm] = useState({
    newPassword: '',
    confirmPassword: '',
  });

  const [userInfo, setUserInfo] = useState({
    level: '',
    status: '',
  });

  const ACCOUNT_GROUPS = {
    profile: {
      title: translate('pages.account_settings.basic_info'),
      description: translate('pages.account_settings.basic_info_desc'),
      icon: <PersonIcon />,
      color: '#1976d2',
    },
    password: {
      title: translate('pages.account_settings.security'),
      description: translate('pages.account_settings.security_desc'),
      icon: <LockIcon />,
      color: '#d32f2f',
    },
  };

  useEffect(() => {
    const cachedUser = localStorage.getItem('user');
    if (!cachedUser) {
      return;
    }
    try {
      const parsed = JSON.parse(cachedUser) as OperatorProfile;
      setProfileForm({
        username: parsed.username ?? '',
        realname: parsed.realname ?? '',
        mobile: parsed.mobile ?? '',
        email: parsed.email ?? '',
        remark: parsed.remark ?? '',
      });
      setUserInfo({
        level: parsed.level ?? '',
        status: parsed.status ?? '',
      });
    } catch {
      // ignore malformed cache
    }
  }, []);

  const profileQuery = useApiQuery<OperatorProfile>({
    path: '/system/operators/me',
    queryKey: ACCOUNT_QUERY_KEY,
    staleTime: 5 * 60 * 1000,
    retry: 1,
  });

  const handleAuthError = (error: unknown) => {
    if (error instanceof ApiError && error.status === 401) {
      notify('认证已过期，请重新登录', { type: 'error' });
      clearAuthStorage();
      window.location.href = '/login';
      return true;
    }
    return false;
  };

  useEffect(() => {
    if (profileQuery.data) {
      const data = profileQuery.data;
      setProfileForm({
        username: data.username ?? '',
        realname: data.realname ?? '',
        mobile: data.mobile ?? '',
        email: data.email ?? '',
        remark: data.remark ?? '',
      });
      setUserInfo({
        level: data.level ?? '',
        status: data.status ?? '',
      });
      localStorage.setItem('user', JSON.stringify(data));
    }
  }, [profileQuery.data]);

  useEffect(() => {
    if (profileQuery.error) {
      handleAuthError(profileQuery.error);
    }
  }, [profileQuery.error]);

  const handleGroupToggle = (group: string) => {
    setExpandedGroups(prev =>
      prev.includes(group)
        ? prev.filter(g => g !== group)
        : [...prev, group],
    );
  };

  const profileMutation = useApiMutation<OperatorProfile>({
    onSuccess: data => {
      setProfileForm({
        username: data.username ?? '',
        realname: data.realname ?? '',
        mobile: data.mobile ?? '',
        email: data.email ?? '',
        remark: data.remark ?? '',
      });
      setUserInfo({
        level: data.level ?? '',
        status: data.status ?? '',
      });
      localStorage.setItem('user', JSON.stringify(data));
      queryClient.setQueryData(ACCOUNT_QUERY_KEY, data);
      notify('用户信息更新成功');
    },
    onError: error => {
      if (handleAuthError(error)) {
        return;
      }
      const message = error instanceof Error ? error.message : '未知错误';
      notify(`更新用户信息失败: ${message}`, { type: 'error' });
    },
  });

  const passwordMutation = useApiMutation<{ message?: string }>({
    onSuccess: () => {
      notify('密码修改成功');
      setPasswordForm({ newPassword: '', confirmPassword: '' });
    },
    onError: error => {
      if (handleAuthError(error)) {
        return;
      }
      const message = error instanceof Error ? error.message : '未知错误';
      notify(`密码修改失败: ${message}`, { type: 'error' });
    },
  });

  const handleProfileSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!profileForm.realname || !profileForm.realname.trim()) {
      notify('真实姓名不能为空', { type: 'error' });
      return;
    }
    const requestData: Record<string, string> = {
      realname: profileForm.realname.trim(),
    };

    if (profileForm.username?.trim()) {
      requestData.username = profileForm.username.trim();
    }
    if (profileForm.mobile?.trim()) {
      requestData.mobile = profileForm.mobile.trim();
    }
    if (profileForm.email?.trim()) {
      requestData.email = profileForm.email.trim();
    }
    if (profileForm.remark?.trim()) {
      requestData.remark = profileForm.remark.trim();
    }

    profileMutation.mutate({
      path: '/system/operators/me',
      method: 'PUT',
      body: requestData,
    });
  };

  const handlePasswordSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      notify('新密码和确认密码不一致', { type: 'error' });
      return;
    }

    if (passwordForm.newPassword.length < 6) {
      notify('新密码长度至少6位', { type: 'error' });
      return;
    }

    passwordMutation.mutate({
      path: '/system/operators/me',
      method: 'PUT',
      body: { password: passwordForm.newPassword },
    });
  };

  const getLevelLabel = (level: string) => {
    const key = `pages.account_settings.level.${level}` as const;
    return translate(key, { _: level });
  };

  const getStatusLabel = (status: string) => {
    const key = status === 'enabled' ? 'pages.account_settings.status.enabled' : 'pages.account_settings.status.disabled';
    return translate(key);
  };

  return (
    <Box sx={{ p: 3 }}>
      <Card>
        <CardContent sx={{ p: { xs: 2, sm: 3 } }}>
          <Typography variant="h5" gutterBottom sx={{ mb: 3, fontWeight: 600 }}>
            {translate('pages.account_settings.title')}
          </Typography>

          {/* 个人信息组 */}
          <Accordion
            expanded={expandedGroups.includes('profile')}
            onChange={() => handleGroupToggle('profile')}
            sx={{
              mb: 2,
              '&:before': {
                display: 'none',
              },
            }}
          >
            <AccordionSummary
              expandIcon={<ExpandMoreIcon />}
              sx={{
                backgroundColor: `${ACCOUNT_GROUPS.profile.color}08`,
                borderLeft: `4px solid ${ACCOUNT_GROUPS.profile.color}`,
                minHeight: 64,
                '& .MuiAccordionSummary-content': {
                  alignItems: 'center',
                  my: 1,
                },
              }}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                <Box sx={{ color: ACCOUNT_GROUPS.profile.color, display: 'flex', alignItems: 'center' }}>
                  {ACCOUNT_GROUPS.profile.icon}
                </Box>
                <Box>
                  <Typography variant="h6" sx={{ fontWeight: 600, color: ACCOUNT_GROUPS.profile.color }}>
                    {ACCOUNT_GROUPS.profile.title}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    {ACCOUNT_GROUPS.profile.description}
                  </Typography>
                </Box>
              </Box>
            </AccordionSummary>
            <AccordionDetails sx={{ px: 3, py: 3 }}>
              <form onSubmit={handleProfileSubmit}>
                <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))', gap: 3 }}>
                  <TextField
                    fullWidth
                    label={translate('pages.account_settings.fields.username')}
                    value={profileForm.username}
                    onChange={(e) => setProfileForm({ ...profileForm, username: e.target.value })}
                    required
                    helperText={translate('pages.account_settings.fields.username_helper')}
                    size="medium"
                  />
                  <TextField
                    fullWidth
                    label={translate('pages.account_settings.fields.realname')}
                    value={profileForm.realname}
                    onChange={(e) => setProfileForm({ ...profileForm, realname: e.target.value })}
                    required
                    size="medium"
                  />
                  <TextField
                    fullWidth
                    label={translate('pages.account_settings.fields.mobile')}
                    value={profileForm.mobile}
                    onChange={(e) => setProfileForm({ ...profileForm, mobile: e.target.value })}
                    helperText={translate('pages.account_settings.fields.mobile_helper')}
                    size="medium"
                  />
                  <TextField
                    fullWidth
                    label={translate('pages.account_settings.fields.email')}
                    type="email"
                    value={profileForm.email}
                    onChange={(e) => setProfileForm({ ...profileForm, email: e.target.value })}
                    helperText={translate('pages.account_settings.fields.email_helper')}
                    size="medium"
                  />
                  <TextField
                    fullWidth
                    label={translate('pages.account_settings.fields.remark')}
                    multiline
                    rows={3}
                    value={profileForm.remark}
                    onChange={(e) => setProfileForm({ ...profileForm, remark: e.target.value })}
                    size="medium"
                  />

                  {/* 只读信息 */}
                  <Alert 
                    severity="info" 
                    sx={{ mt: 2 }}
                    icon={<InfoIcon />}
                  >
                    <Typography variant="subtitle2" sx={{ mb: 1, fontSize: { xs: '0.875rem', sm: '1rem' } }}>
                      {translate('pages.account_settings.permission_info')}
                    </Typography>
                    <Box sx={{ 
                      display: 'flex', 
                      flexDirection: { xs: 'column', sm: 'row' },
                      gap: { xs: 1, sm: 3 } 
                    }}>
                      <Typography variant="body2" sx={{ fontSize: { xs: '0.875rem', sm: '1rem' } }}>
                        <strong>{translate('pages.account_settings.permission_level')}：</strong>
                        <span
                          style={{
                            color:
                              userInfo.level === 'super'
                                ? '#d32f2f'
                                : userInfo.level === 'admin'
                                ? '#ed6c02'
                                : '#0288d1',
                          }}
                        >
                          {getLevelLabel(userInfo.level)}
                        </span>
                      </Typography>
                      <Typography variant="body2" sx={{ fontSize: { xs: '0.875rem', sm: '1rem' } }}>
                        <strong>{translate('pages.account_settings.account_status')}：</strong>
                        <span style={{ color: userInfo.status === 'enabled' ? '#2e7d32' : '#757575' }}>
                          {getStatusLabel(userInfo.status)}
                        </span>
                      </Typography>
                    </Box>
                  </Alert>

                  <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: { xs: 2, sm: 3 } }}>
                    <Button
                      type="submit"
                      variant="contained"
                      disabled={profileMutation.isPending}
                      startIcon={profileMutation.isPending ? <CircularProgress size={20} /> : null}
                      size="large"
                      sx={{ 
                        minWidth: { xs: '100%', sm: 'auto' },
                        py: { xs: 1.5, sm: 1 }
                      }}
                    >
                      {profileMutation.isPending ? translate('pages.account_settings.saving') : translate('pages.account_settings.save_profile')}
                    </Button>
                  </Box>
                </Box>
              </form>
            </AccordionDetails>
          </Accordion>

          {/* 密码修改组 */}
          <Accordion
            expanded={expandedGroups.includes('password')}
            onChange={() => handleGroupToggle('password')}
            sx={{
              '&:before': {
                display: 'none',
              },
            }}
          >
            <AccordionSummary
              expandIcon={<ExpandMoreIcon />}
              sx={{
                backgroundColor: `${ACCOUNT_GROUPS.password.color}08`,
                borderLeft: `4px solid ${ACCOUNT_GROUPS.password.color}`,
                minHeight: 64,
                '& .MuiAccordionSummary-content': {
                  alignItems: 'center',
                  my: 1,
                },
              }}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                <Box sx={{ color: ACCOUNT_GROUPS.password.color, display: 'flex', alignItems: 'center' }}>
                  {ACCOUNT_GROUPS.password.icon}
                </Box>
                <Box>
                  <Typography variant="h6" sx={{ fontWeight: 600, color: ACCOUNT_GROUPS.password.color }}>
                    {ACCOUNT_GROUPS.password.title}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    {ACCOUNT_GROUPS.password.description}
                  </Typography>
                </Box>
              </Box>
            </AccordionSummary>
            <AccordionDetails sx={{ px: 3, py: 3 }}>
              <form onSubmit={handlePasswordSubmit}>
                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
                  {/* 提示块独占一行 */}
                  <Alert severity="warning">
                    <Typography variant="body2" sx={{ fontSize: { xs: '0.875rem', sm: '1rem' } }}>
                      {translate('pages.account_settings.password_requirement')}
                    </Typography>
                  </Alert>

                  {/* 密码输入框区域 */}
                  <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))', gap: 3 }}>
                    <TextField
                      fullWidth
                      label={translate('pages.account_settings.new_password')}
                      type="password"
                      value={passwordForm.newPassword}
                      onChange={(e) => setPasswordForm({ ...passwordForm, newPassword: e.target.value })}
                      required
                      size="medium"
                    />

                    <TextField
                      fullWidth
                      label={translate('pages.account_settings.confirm_password')}
                      type="password"
                      value={passwordForm.confirmPassword}
                      onChange={(e) => setPasswordForm({ ...passwordForm, confirmPassword: e.target.value })}
                      required
                      error={passwordForm.confirmPassword !== '' && passwordForm.newPassword !== passwordForm.confirmPassword}
                      helperText={
                        passwordForm.confirmPassword !== '' && passwordForm.newPassword !== passwordForm.confirmPassword
                          ? translate('pages.account_settings.password_mismatch')
                          : ''
                      }
                      size="medium"
                    />
                  </Box>

                  <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: { xs: 2, sm: 3 } }}>
                    <Button
                      type="submit"
                      variant="contained"
                      disabled={passwordMutation.isPending}
                      startIcon={passwordMutation.isPending ? <CircularProgress size={20} /> : null}
                      size="large"
                      sx={{ 
                        minWidth: { xs: '100%', sm: 'auto' },
                        py: { xs: 1.5, sm: 1 }
                      }}
                    >
                      {passwordMutation.isPending ? translate('pages.account_settings.changing') : translate('pages.account_settings.change_password')}
                    </Button>
                  </Box>
                </Box>
              </form>
            </AccordionDetails>
          </Accordion>
        </CardContent>
      </Card>
    </Box>
  );
}
