import { useEffect, useState } from 'react';
import { useNotify } from 'react-admin';
import {
  Card,
  CardContent,
  Typography,
  Box,
  Button,
  TextField,
  Alert,
} from '@mui/material';

export default function AuthTest() {
  const notify = useNotify();
  const [token, setToken] = useState('');
  const [result, setResult] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const savedToken = localStorage.getItem('token') || '';
    setToken(savedToken);
  }, []);

  const testLogin = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: 'admin',
          password: 'admin',
        }),
      });

      const data = await response.json();
      console.log('Login response:', data);
      
      if (response.ok && data.data && !data.error) {
        const newToken = data.data.token;
        localStorage.setItem('token', newToken);
        localStorage.setItem('user', JSON.stringify(data.data.user));
        setToken(newToken);
        setResult(`登录成功，获得token: ${newToken.substring(0, 20)}...`);
        notify('登录成功', { type: 'success' });
      } else {
        setResult(`登录失败: ${data.message || '未知错误'}`);
        notify('登录失败', { type: 'error' });
      }
    } catch (error) {
      console.error('Login error:', error);
      setResult(`登录异常: ${error instanceof Error ? error.message : '未知错误'}`);
    } finally {
      setLoading(false);
    }
  };

  const testApiCall = async () => {
    if (!token) {
      notify('请先登录获取token', { type: 'error' });
      return;
    }

    setLoading(true);
    try {
      console.log('Testing with token:', token.substring(0, 20) + '...');
      
      const response = await fetch('/api/v1/system/operators/me', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      console.log('API Response status:', response.status);
      console.log('API Response headers:', Object.fromEntries(response.headers.entries()));
      
      const data = await response.json();
      console.log('API Response data:', data);
      
      if (response.ok && data.data && !data.error) {
        setResult(`API调用成功: ${JSON.stringify(data.data, null, 2)}`);
        notify('API调用成功', { type: 'success' });
      } else {
        setResult(`API调用失败: ${data.message || '未知错误'}`);
        notify('API调用失败', { type: 'error' });
      }
    } catch (error) {
      console.error('API error:', error);
      setResult(`API异常: ${error instanceof Error ? error.message : '未知错误'}`);
    } finally {
      setLoading(false);
    }
  };

  const testCurl = () => {
    if (!token) {
      notify('请先登录获取token', { type: 'error' });
      return;
    }
    
    const curlCommand = `curl -H "Authorization: Bearer ${token}" -H "Content-Type: application/json" http://localhost:1816/api/v1/system/operators/me`;
    navigator.clipboard.writeText(curlCommand);
    notify('curl命令已复制到剪贴板', { type: 'info' });
    setResult(`curl命令:\n${curlCommand}`);
  };

  return (
    <Box sx={{ maxWidth: 800, mx: 'auto', mt: 2, p: 2 }}>
      <Card>
        <CardContent>
          <Typography variant="h5" gutterBottom>
            认证测试页面
          </Typography>
          
          <Alert severity="info" sx={{ mb: 3 }}>
            用于测试登录和API调用的认证流程
          </Alert>
          
          <Box sx={{ mb: 3 }}>
            <TextField
              fullWidth
              label="当前Token"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              multiline
              rows={3}
              helperText="当前存储的认证token"
            />
          </Box>
          
          <Box sx={{ display: 'flex', gap: 2, mb: 3, flexWrap: 'wrap' }}>
            <Button 
              variant="contained" 
              onClick={testLogin}
              disabled={loading}
            >
              {loading ? '登录中...' : '测试登录'}
            </Button>
            <Button 
              variant="outlined" 
              onClick={testApiCall}
              disabled={loading}
            >
              {loading ? '调用中...' : '测试API调用'}
            </Button>
            <Button 
              variant="text" 
              onClick={testCurl}
            >
              生成curl命令
            </Button>
          </Box>
          
          <Typography variant="h6" sx={{ mb: 2 }}>
            测试结果:
          </Typography>
          
          <Box 
            component="pre" 
            sx={{ 
              bgcolor: '#f5f5f5', 
              p: 2, 
              borderRadius: 1, 
              overflow: 'auto',
              fontSize: '0.875rem',
              whiteSpace: 'pre-wrap'
            }}
          >
            {result || '等待测试...'}
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
}