import SettingsOutlinedIcon from '@mui/icons-material/SettingsOutlined';
import AccountCircleOutlinedIcon from '@mui/icons-material/AccountCircleOutlined';
import { Box, IconButton, Stack, Tooltip, Typography, useTheme } from '@mui/material';
import { AppBar, AppBarProps, TitlePortal, ToggleThemeButton, useRedirect, useGetIdentity } from 'react-admin';

export const CustomAppBar = (props: AppBarProps) => {
  const redirect = useRedirect();
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const { data: identity } = useGetIdentity();

  return (
    <AppBar
      {...props}
      toolbar={false}
      elevation={0}
      sx={{
        // 浅色主题使用白色背景，深色主题使用深色背景
        backgroundColor: isDark ? '#1e293b' : '#ffffff',
        color: isDark ? '#f1f5f9' : '#1f2937',
        borderBottom: isDark 
          ? '1px solid rgba(148, 163, 184, 0.2)'
          : '1px solid rgba(229, 231, 235, 0.8)',
        boxShadow: isDark 
          ? 'none' 
          : '0 1px 3px 0 rgba(0, 0, 0, 0.05)',
        transition: 'all 0.3s ease',
      }}
    >
      <TitlePortal />
      <Box
        sx={{
          width: '100%',
          px: 3,
          py: 1,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <Stack direction="row" spacing={2} alignItems="center">
          <Box>
            <Typography 
              variant="h6" 
              sx={{ 
                fontSize: 18, 
                fontWeight: 700, 
                color: isDark ? '#f1f5f9' : '#1f2937',
                letterSpacing: '0.5px',
              }}
            >
              ToughRADIUS
            </Typography>
          </Box>
        </Stack>

        <Stack direction="row" spacing={1.5} alignItems="center">
          <Tooltip title="切换主题">
            <Box 
              sx={{ 
                '& svg': { 
                  fontSize: 22, 
                  color: isDark ? '#f1f5f9' : '#6b7280',
                },
                '& button': {
                  color: isDark ? '#f1f5f9' : '#6b7280',
                  transition: 'all 0.2s ease',
                  '&:hover': {
                    transform: 'rotate(180deg)',
                    backgroundColor: isDark 
                      ? 'rgba(255, 255, 255, 0.1)'
                      : 'rgba(0, 0, 0, 0.05)',
                  },
                },
              }}
            >
              <ToggleThemeButton />
            </Box>
          </Tooltip>
          
          {/* 只对超级管理员和管理员显示系统设置按钮 */}
          {identity?.level === 'super' || identity?.level === 'admin' ? (
            <Tooltip title="系统设置">
              <IconButton 
                size="large" 
                onClick={() => redirect('/system/config')}
                sx={{
                  color: isDark ? '#f1f5f9' : '#6b7280',
                  transition: 'all 0.2s ease',
                  '&:hover': {
                    transform: 'scale(1.05)',
                    backgroundColor: isDark 
                      ? 'rgba(255, 255, 255, 0.1)'
                      : 'rgba(0, 0, 0, 0.05)',
                  },
                }}
              >
                <SettingsOutlinedIcon />
              </IconButton>
            </Tooltip>
          ) : null}

          {/* 账号设置按钮 - 所有用户都可见 */}
          <Tooltip title="账号设置">
            <IconButton 
              size="large" 
              onClick={() => redirect('/account/settings')}
              sx={{
                color: isDark ? '#f1f5f9' : '#6b7280',
                transition: 'all 0.2s ease',
                '&:hover': {
                  transform: 'scale(1.05)',
                  backgroundColor: isDark 
                    ? 'rgba(255, 255, 255, 0.1)'
                    : 'rgba(0, 0, 0, 0.05)',
                },
              }}
            >
              <AccountCircleOutlinedIcon />
            </IconButton>
          </Tooltip>
        </Stack>
      </Box>
    </AppBar>
  );
};
