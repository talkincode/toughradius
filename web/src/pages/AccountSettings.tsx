import { useEffect, useState } from 'react';
import { useNotify } from 'react-admin';
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

// 账号设置分组配置
const ACCOUNT_GROUPS = {
  profile: {
    title: '个人信息',
    description: '基本个人信息和账号详情',
    icon: <PersonIcon />,
    color: '#1976d2',
  },
  password: {
    title: '密码修改',
    description: '修改登录密码',
    icon: <LockIcon />,
    color: '#d32f2f',
  },
};

export default function AccountSettings() {
  const notify = useNotify();
  const [profileLoading, setProfileLoading] = useState(false);
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [expandedGroups, setExpandedGroups] = useState<string[]>(['profile']);
  
  // 用户信息表单状态
  const [profileForm, setProfileForm] = useState({
    username: '',
    realname: '',
    mobile: '',
    email: '',
    remark: ''
  });

  // 密码修改表单状态
  const [passwordForm, setPasswordForm] = useState({
    newPassword: '',
    confirmPassword: ''
  });

  // 只读信息
  const [userInfo, setUserInfo] = useState({
    level: '',
    status: ''
  });

  useEffect(() => {
    const loadUserData = async () => {
      try {
        const token = localStorage.getItem('token');
        const user = localStorage.getItem('user');
        
        console.log('=== Debug Info ===');
        console.log('Token exists:', token ? 'YES' : 'NO');
        console.log('User in localStorage:', user);
        
        if (!token) {
          notify('请先登录', { type: 'error' });
          return;
        }
        
        // 如果localStorage中有用户信息，先使用它
        if (user) {
          try {
            const userData = JSON.parse(user);
            console.log('Parsed user data from localStorage:', userData);
            setProfileForm({
              username: userData.username || '',
              realname: userData.realname || '',
              mobile: userData.mobile || '',
              email: userData.email || '',
              remark: userData.remark || ''
            });
            setUserInfo({
              level: userData.level || '',
              status: userData.status || ''
            });
          } catch (parseError) {
            console.error('Failed to parse user data from localStorage:', parseError);
          }
        }
        
        // 然后从API获取最新数据
        console.log('Fetching latest user data from API...');
        const response = await fetch('/api/v1/system/operators/me', {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        });

        console.log('API Response status:', response.status);
        
        if (!response.ok) {
          if (response.status === 401) {
            notify('认证已过期，请重新登录', { type: 'error' });
            localStorage.clear();
            window.location.href = '/login';
            return;
          }
          
          // 尝试解析错误响应中的具体错误消息
          const errorText = await response.text();
          console.error('API Error response body:', errorText);
          
          try {
            const errorData = JSON.parse(errorText);
            throw new Error(errorData.message || `HTTP ${response.status}: ${response.statusText}`);
          } catch (parseError) {
            // 如果无法解析JSON，使用原始错误文本或状态信息
            throw new Error(errorText || `HTTP ${response.status}: ${response.statusText}`);
          }
        }
        
        const result = await response.json();
        console.log('API Response data:', result);
        
        // 后端成功响应格式: { "data": {...} }
        // 错误响应格式: { "error": "code", "message": "msg" }
        if (result.data && !result.error) {
          const data = result.data;
          console.log('Updating form with API data:', data);
          
          setProfileForm({
            username: data.username || '',
            realname: data.realname || '',
            mobile: data.mobile || '',
            email: data.email || '',
            remark: data.remark || ''
          });
          setUserInfo({
            level: data.level || '',
            status: data.status || ''
          });
          
          // 更新localStorage中的用户信息
          localStorage.setItem('user', JSON.stringify(data));
          console.log('User data loaded and updated successfully');
        } else {
          console.error('API response not successful:', result);
          notify(result.message || '加载用户信息失败', { type: 'error' });
        }
      } catch (error) {
        console.error('Failed to load user data:', error);
        const errorMessage = error instanceof Error ? error.message : '未知错误';
        notify(`加载用户信息失败: ${errorMessage}`, { type: 'error' });
      }
    };

    // 检查认证状态后加载数据
    const token = localStorage.getItem('token');
    if (token) {
      loadUserData();
    } else {
      notify('未找到认证信息，请重新登录', { type: 'error' });
    }
    
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [notify]); // 只依赖notify

  const handleGroupToggle = (group: string) => {
    setExpandedGroups(prev => 
      prev.includes(group) 
        ? prev.filter(g => g !== group)
        : [...prev, group]
    );
  };

  const handleProfileSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    console.log('=== Profile Submit Debug ===');
    console.log('Form data:', profileForm);
    
    // 基本验证
    if (!profileForm.realname || !profileForm.realname.trim()) {
      notify('真实姓名不能为空', { type: 'error' });
      return;
    }
    
    setProfileLoading(true);
    
    try {
      const token = localStorage.getItem('token');
      console.log('Token for update:', token ? 'present' : 'missing');
      
      if (!token) {
        notify('请先登录', { type: 'error' });
        return;
      }
      
      // 准备请求数据，只包含有值的字段
      const requestData: Record<string, string> = {};
      
      // realname是必需的
      requestData.realname = profileForm.realname.trim();
      
      // 其他可选字段
      if (profileForm.username && profileForm.username.trim()) {
        requestData.username = profileForm.username.trim();
      }
      
      if (profileForm.mobile && profileForm.mobile.trim()) {
        requestData.mobile = profileForm.mobile.trim();
      }
      
      if (profileForm.email && profileForm.email.trim()) {
        requestData.email = profileForm.email.trim();
      }
      
      if (profileForm.remark && profileForm.remark.trim()) {
        requestData.remark = profileForm.remark.trim();
      }
      
      console.log('Request data:', requestData);
      console.log('Sending PUT request to /api/v1/system/operators/me');
      
      const response = await fetch('/api/v1/system/operators/me', {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestData),
      });

      console.log('Update response status:', response.status);
      
      if (!response.ok) {
        const errorText = await response.text();
        console.error('Update error response body:', errorText);
        
        if (response.status === 401) {
          notify('认证已过期，请重新登录', { type: 'error' });
          localStorage.clear();
          window.location.href = '/login';
          return;
        }
        
        // 尝试解析错误响应
        try {
          const errorData = JSON.parse(errorText);
          throw new Error(errorData.message || `HTTP ${response.status}: ${response.statusText}`);
        } catch (parseError) {
          // 如果无法解析JSON，使用原始错误文本或状态信息
          throw new Error(errorText || `HTTP ${response.status}: ${response.statusText}`);
        }
      }
      
      const result = await response.json();
      console.log('Update response data:', result);
      
      // 后端成功响应格式: { "data": {...} }
      // 错误响应格式: { "error": "code", "message": "msg" }
      if (result.data && !result.error) {
        notify('用户信息更新成功');
        
        // 更新localStorage中的用户信息
        localStorage.setItem('user', JSON.stringify(result.data));
        
        console.log('Profile updated successfully');
      } else {
        console.error('Update failed:', result);
        notify(result.message || '更新用户信息失败', { type: 'error' });
      }
    } catch (error) {
      console.error('Failed to update profile:', error);
      const errorMessage = error instanceof Error ? error.message : '未知错误';
      notify(`更新用户信息失败: ${errorMessage}`, { type: 'error' });
    } finally {
      setProfileLoading(false);
    }
  };  useEffect(() => {
    const loadUserData = async () => {
      try {
        const token = localStorage.getItem('token');
        const user = localStorage.getItem('user');
        
        console.log('=== Debug Info ===');
        console.log('Token exists:', token ? 'YES' : 'NO');
        console.log('User in localStorage:', user);
        
        if (!token) {
          notify('请先登录', { type: 'error' });
          return;
        }
        
        // 如果localStorage中有用户信息，先使用它
        if (user) {
          try {
            const userData = JSON.parse(user);
            console.log('Parsed user data from localStorage:', userData);
            setProfileForm({
              username: userData.username || '',
              realname: userData.realname || '',
              mobile: userData.mobile || '',
              email: userData.email || '',
              remark: userData.remark || ''
            });
            setUserInfo({
              level: userData.level || '',
              status: userData.status || ''
            });
          } catch (parseError) {
            console.error('Failed to parse user data from localStorage:', parseError);
          }
        }
        
        // 然后从API获取最新数据
        console.log('Fetching latest user data from API...');
        const response = await fetch('/api/v1/system/operators/me', {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        });

        console.log('API Response status:', response.status);
        console.log('API Response headers:', Object.fromEntries(response.headers.entries()));
        
        if (!response.ok) {
          if (response.status === 401) {
            notify('认证已过期，请重新登录', { type: 'error' });
            localStorage.clear();
            window.location.href = '/login';
            return;
          }
          const errorText = await response.text();
          console.error('API Error response body:', errorText);
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const result = await response.json();
        console.log('API Response data:', result);
        
        // 后端成功响应格式: { "data": {...} }
        // 错误响应格式: { "error": "code", "message": "msg" }
        if (result.data && !result.error) {
          const data = result.data;
          console.log('Updating form with API data:', data);
          
          setProfileForm({
            username: data.username || '',
            realname: data.realname || '',
            mobile: data.mobile || '',
            email: data.email || '',
            remark: data.remark || ''
          });
          setUserInfo({
            level: data.level || '',
            status: data.status || ''
          });
          
          // 更新localStorage中的用户信息
          localStorage.setItem('user', JSON.stringify(data));
          console.log('User data loaded and updated successfully');
        } else {
          console.error('API response not successful:', result);
          notify(result.message || '加载用户信息失败', { type: 'error' });
        }
      } catch (error) {
        console.error('Failed to load user data:', error);
        const errorMessage = error instanceof Error ? error.message : '未知错误';
        notify(`加载用户信息失败: ${errorMessage}`, { type: 'error' });
      }
    };

    // 检查认证状态后加载数据
    const token = localStorage.getItem('token');
    if (token) {
      loadUserData();
    } else {
      notify('未找到认证信息，请重新登录', { type: 'error' });
    }
    
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [notify]); // 只依赖notify，忽略identity和isLoading

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

    setPasswordLoading(true);
    
    try {
      const token = localStorage.getItem('token');
      
      if (!token) {
        notify('请先登录', { type: 'error' });
        return;
      }
      
      // 使用现有的 updateCurrentOperator API，只传递 password 字段
      const response = await fetch('/api/v1/system/operators/me', {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          password: passwordForm.newPassword,
        }),
      });

      if (!response.ok) {
        if (response.status === 401) {
          notify('认证已过期，请重新登录', { type: 'error' });
          localStorage.clear();
          window.location.href = '/login';
          return;
        }
        
        // 尝试解析错误响应中的具体错误消息
        const errorText = await response.text();
        console.error('Password update error response body:', errorText);
        
        try {
          const errorData = JSON.parse(errorText);
          throw new Error(errorData.message || `HTTP ${response.status}: ${response.statusText}`);
        } catch (parseError) {
          // 如果无法解析JSON，使用原始错误文本或状态信息
          throw new Error(errorText || `HTTP ${response.status}: ${response.statusText}`);
        }
      }
      
      const result = await response.json();
      if (result.data && !result.error) {
        notify('密码修改成功');
        setPasswordForm({
          newPassword: '',
          confirmPassword: ''
        });
      } else {
        notify(result.message || '密码修改失败', { type: 'error' });
      }
    } catch (error) {
      console.error('Failed to change password:', error);
      const errorMessage = error instanceof Error ? error.message : '未知错误';
      notify(`密码修改失败: ${errorMessage}`, { type: 'error' });
    } finally {
      setPasswordLoading(false);
    }
  };

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
    <Box sx={{ p: 3 }}>
      <Card>
        <CardContent sx={{ p: { xs: 2, sm: 3 } }}>
          <Typography variant="h5" gutterBottom sx={{ mb: 3, fontWeight: 600 }}>
            账号设置
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
                    label="用户名"
                    value={profileForm.username}
                    onChange={(e) => setProfileForm({ ...profileForm, username: e.target.value })}
                    required
                    helperText="3-30个字符，只能包含字母、数字和下划线"
                    size="medium"
                  />
                  <TextField
                    fullWidth
                    label="真实姓名"
                    value={profileForm.realname}
                    onChange={(e) => setProfileForm({ ...profileForm, realname: e.target.value })}
                    required
                    size="medium"
                  />
                  <TextField
                    fullWidth
                    label="手机号"
                    value={profileForm.mobile}
                    onChange={(e) => setProfileForm({ ...profileForm, mobile: e.target.value })}
                    helperText="请输入11位中国大陆手机号"
                    size="medium"
                  />
                  <TextField
                    fullWidth
                    label="邮箱"
                    type="email"
                    value={profileForm.email}
                    onChange={(e) => setProfileForm({ ...profileForm, email: e.target.value })}
                    helperText="用于接收通知和找回密码"
                    size="medium"
                  />
                  <TextField
                    fullWidth
                    label="备注"
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
                      当前权限信息（不可修改）
                    </Typography>
                    <Box sx={{ 
                      display: 'flex', 
                      flexDirection: { xs: 'column', sm: 'row' },
                      gap: { xs: 1, sm: 3 } 
                    }}>
                      <Typography variant="body2" sx={{ fontSize: { xs: '0.875rem', sm: '1rem' } }}>
                        <strong>权限级别：</strong>
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
                        <strong>账号状态：</strong>
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
                      disabled={profileLoading}
                      startIcon={profileLoading ? <CircularProgress size={20} /> : null}
                      size="large"
                      sx={{ 
                        minWidth: { xs: '100%', sm: 'auto' },
                        py: { xs: 1.5, sm: 1 }
                      }}
                    >
                      {profileLoading ? '保存中...' : '保存个人信息'}
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
                      密码长度至少6位，建议包含字母、数字和特殊字符
                    </Typography>
                  </Alert>

                  {/* 密码输入框区域 */}
                  <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))', gap: 3 }}>
                    <TextField
                      fullWidth
                      label="新密码"
                      type="password"
                      value={passwordForm.newPassword}
                      onChange={(e) => setPasswordForm({ ...passwordForm, newPassword: e.target.value })}
                      required
                      size="medium"
                    />

                    <TextField
                      fullWidth
                      label="确认新密码"
                      type="password"
                      value={passwordForm.confirmPassword}
                      onChange={(e) => setPasswordForm({ ...passwordForm, confirmPassword: e.target.value })}
                      required
                      error={passwordForm.confirmPassword !== '' && passwordForm.newPassword !== passwordForm.confirmPassword}
                      helperText={
                        passwordForm.confirmPassword !== '' && passwordForm.newPassword !== passwordForm.confirmPassword
                          ? '密码不一致'
                          : ''
                      }
                      size="medium"
                    />
                  </Box>

                  <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: { xs: 2, sm: 3 } }}>
                    <Button
                      type="submit"
                      variant="contained"
                      disabled={passwordLoading}
                      startIcon={passwordLoading ? <CircularProgress size={20} /> : null}
                      size="large"
                      sx={{ 
                        minWidth: { xs: '100%', sm: 'auto' },
                        py: { xs: 1.5, sm: 1 }
                      }}
                    >
                      {passwordLoading ? '修改中...' : '修改密码'}
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
