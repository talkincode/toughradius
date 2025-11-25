import { useCallback } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Typography,
  Stack,
  Alert,
} from '@mui/material';
import {
  Error as ErrorIcon,
  Refresh as RefreshIcon,
  WifiOff as WifiOffIcon,
  Home as HomeIcon,
} from '@mui/icons-material';
import { useTranslate, useRefresh, useRedirect } from 'react-admin';

interface CustomErrorProps {
  error?: Error | string;
  errorInfo?: React.ErrorInfo;
  resetErrorBoundary?: (...args: unknown[]) => void;
  title?: string;
}

/**
 * CustomError displays a user-friendly error message with retry functionality.
 * It detects connectivity issues and provides appropriate guidance.
 * 
 * @param error - The error object or message string
 * @param errorInfo - React error info from error boundary
 * @param resetErrorBoundary - Function to reset the error boundary
 * @param title - Optional custom title for the error message
 */
export const CustomError = ({
  error,
  resetErrorBoundary,
  title,
}: CustomErrorProps) => {
  const translate = useTranslate();
  const refresh = useRefresh();
  const redirect = useRedirect();

  const errorMessage = error instanceof Error ? error.message : String(error || '');
  
  // Detect if this is a connectivity/network error
  const isConnectivityError = 
    errorMessage.toLowerCase().includes('connectivity') ||
    errorMessage.toLowerCase().includes('network') ||
    errorMessage.toLowerCase().includes('fetch') ||
    errorMessage.toLowerCase().includes('failed to fetch') ||
    errorMessage.toLowerCase().includes('no connectivity') ||
    errorMessage.toLowerCase().includes('econnrefused');

  const handleRetry = useCallback(() => {
    if (resetErrorBoundary) {
      resetErrorBoundary();
    }
    refresh();
  }, [resetErrorBoundary, refresh]);

  const handleGoHome = useCallback(() => {
    redirect('/');
  }, [redirect]);

  return (
    <Box
      sx={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '60vh',
        p: 2,
      }}
    >
      <Card
        elevation={3}
        sx={{
          maxWidth: 500,
          width: '100%',
          borderRadius: 2,
          overflow: 'hidden',
        }}
      >
        <Box
          sx={{
            backgroundColor: isConnectivityError ? 'warning.main' : 'error.main',
            color: 'white',
            p: 3,
            display: 'flex',
            alignItems: 'center',
            gap: 2,
          }}
        >
          {isConnectivityError ? (
            <WifiOffIcon sx={{ fontSize: 40 }} />
          ) : (
            <ErrorIcon sx={{ fontSize: 40 }} />
          )}
          <Typography variant="h5" fontWeight={600}>
            {title || (isConnectivityError 
              ? translate('error.connectivity_title', { _: '连接问题' })
              : translate('error.general_title', { _: '出错了' })
            )}
          </Typography>
        </Box>

        <CardContent sx={{ p: 3 }}>
          <Stack spacing={3}>
            {isConnectivityError ? (
              <Alert severity="warning">
                {translate('error.connectivity_message', {
                  _: '无法连接到服务器。请检查您的网络连接或后端服务是否正常运行。',
                })}
              </Alert>
            ) : (
              <Alert severity="error">
                {translate('error.general_message', {
                  _: '加载数据时发生错误。请稍后重试。',
                })}
              </Alert>
            )}

            {errorMessage && !isConnectivityError && (
              <Typography
                variant="body2"
                color="text.secondary"
                sx={{
                  p: 2,
                  backgroundColor: 'grey.100',
                  borderRadius: 1,
                  fontFamily: 'monospace',
                  wordBreak: 'break-all',
                }}
              >
                {errorMessage}
              </Typography>
            )}

            {isConnectivityError && (
              <Box>
                <Typography variant="subtitle2" fontWeight={600} gutterBottom>
                  {translate('error.troubleshooting_title', { _: '故障排除建议：' })}
                </Typography>
                <Typography variant="body2" color="text.secondary" component="ul" sx={{ pl: 2 }}>
                  <li>{translate('error.troubleshooting_network', { _: '检查您的网络连接是否正常' })}</li>
                  <li>{translate('error.troubleshooting_server', { _: '确认后端服务正在运行' })}</li>
                  <li>{translate('error.troubleshooting_refresh', { _: '刷新页面重试' })}</li>
                </Typography>
              </Box>
            )}

            <Stack direction="row" spacing={2} justifyContent="flex-end">
              <Button
                variant="outlined"
                startIcon={<HomeIcon />}
                onClick={handleGoHome}
              >
                {translate('error.go_home', { _: '返回首页' })}
              </Button>
              <Button
                variant="contained"
                startIcon={<RefreshIcon />}
                onClick={handleRetry}
                color={isConnectivityError ? 'warning' : 'primary'}
              >
                {translate('error.retry', { _: '重试' })}
              </Button>
            </Stack>
          </Stack>
        </CardContent>
      </Card>
    </Box>
  );
};

export default CustomError;
