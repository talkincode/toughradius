import { useEffect, useState } from 'react';
import { useNotify } from 'react-admin';
import {
  Card,
  CardContent,
  Typography,
  Box,
  Button,
  Alert,
} from '@mui/material';

interface DebugInfo {
  hasToken?: boolean;
  tokenLength?: number;
  tokenPreview?: string | null;
  hasUser?: boolean;
  userData?: Record<string, unknown> | null;
  username?: string | null;
  timestamp?: string;
  apiResponse?: Record<string, unknown>;
  lastApiCall?: string;
}

export default function DebugAccountSettings() {
  const notify = useNotify();
  const [debugInfo, setDebugInfo] = useState<DebugInfo>({});
  const [loading, setLoading] = useState(false);

  const checkAuth = () => {
    const token = localStorage.getItem('token');
    const user = localStorage.getItem('user');
    const username = localStorage.getItem('username');
    
    const info = {
      hasToken: !!token,
      tokenLength: token ? token.length : 0,
      tokenPreview: token ? token.substring(0, 20) + '...' : null,
      hasUser: !!user,
      userData: user ? JSON.parse(user) : null,
      username: username,
      timestamp: new Date().toISOString()
    };
    
    setDebugInfo(info);
    console.log('Auth Debug Info:', info);
  };

  const testAPI = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      
      if (!token) {
        notify('没有找到认证令牌', { type: 'error' });
        return;
      }
      
      console.log('Testing API call...');
      const response = await fetch('/api/v1/system/operators/me', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      
      console.log('Response status:', response.status);
      console.log('Response headers:', Object.fromEntries(response.headers.entries()));
      
      if (!response.ok) {
        const errorText = await response.text();
        console.error('API Error:', errorText);
        notify(`API调用失败: ${response.status} ${response.statusText}`, { type: 'error' });
        return;
      }
      
      const result = await response.json();
      console.log('API Response:', result);
      
      if (result.success) {
        notify('API调用成功');
        setDebugInfo({
          ...debugInfo,
          apiResponse: result,
          lastApiCall: new Date().toISOString()
        });
      } else {
        notify(`API返回错误: ${result.message}`, { type: 'error' });
      }
      
    } catch (error) {
      console.error('API Call Error:', error);
      const errorMessage = error instanceof Error ? error.message : '未知错误';
      notify(`API调用异常: ${errorMessage}`, { type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    checkAuth();
  }, []);

  return (
    <Box sx={{ maxWidth: 800, mx: 'auto', mt: 2, p: 2 }}>
      <Card>
        <CardContent>
          <Typography variant="h5" gutterBottom>
            账号设置调试页面
          </Typography>
          
          <Alert severity="info" sx={{ mb: 3 }}>
            此页面用于调试账号设置功能的问题
          </Alert>
          
          <Box sx={{ display: 'flex', gap: 2, mb: 3 }}>
            <Button variant="outlined" onClick={checkAuth}>
              检查认证状态
            </Button>
            <Button 
              variant="contained" 
              onClick={testAPI} 
              disabled={loading}
            >
              {loading ? '测试中...' : '测试API调用'}
            </Button>
          </Box>
          
          <Typography variant="h6" sx={{ mb: 2 }}>
            调试信息:
          </Typography>
          
          <Box 
            component="pre" 
            sx={{ 
              bgcolor: '#f5f5f5', 
              p: 2, 
              borderRadius: 1, 
              overflow: 'auto',
              fontSize: '0.875rem'
            }}
          >
            {JSON.stringify(debugInfo, null, 2)}
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
}