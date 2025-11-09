import SettingsOutlinedIcon from '@mui/icons-material/SettingsOutlined';
import { Box, IconButton, Stack, Tooltip, Typography, useTheme } from '@mui/material';
import { AppBar, AppBarProps, TitlePortal, ToggleThemeButton, useRedirect } from 'react-admin';

export const CustomAppBar = (props: AppBarProps) => {
  const redirect = useRedirect();
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  return (
    <AppBar
      {...props}
      elevation={0}
      sx={{
        // 浅色主题使用浅灰色背景，深色主题使用深色背景
        backgroundColor: isDark ? '#1e293b' : '#f1f5f9',
        color: isDark ? '#f1f5f9' : '#0f172a',
        borderBottom: isDark 
          ? '1px solid rgba(148, 163, 184, 0.2)'
          : '1px solid rgba(100, 116, 139, 0.12)',
        boxShadow: 'none',
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
                color: isDark ? '#f1f5f9' : '#0f172a',
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
                  color: isDark ? '#f1f5f9' : '#475569',
                },
                '& button': {
                  color: isDark ? '#f1f5f9' : '#475569',
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
          <Tooltip title="系统设置">
            <IconButton 
              size="large" 
              onClick={() => redirect('/system/settings')}
              sx={{
                color: isDark ? '#f1f5f9' : '#475569',
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
        </Stack>
      </Box>
    </AppBar>
  );
};
