import SettingsOutlinedIcon from '@mui/icons-material/SettingsOutlined';
import { Box, IconButton, Stack, Tooltip, Typography } from '@mui/material';
import { AppBar, AppBarProps, TitlePortal, ToggleThemeButton } from 'react-admin';

export const CustomAppBar = (props: AppBarProps) => (
  <AppBar
    {...props}
    elevation={0}
    sx={{
      backgroundColor: '#ffffff',
      borderBottom: '1px solid rgba(15, 23, 42, 0.08)',
      boxShadow: 'none',
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
          <Typography variant="h6" sx={{ fontSize: 16, fontWeight: 600, color: 'text.primary' }}>
            TOUGHRADIUS
          </Typography>
        </Box>
      </Stack>

      <Stack direction="row" spacing={1.5} alignItems="center">
        <Tooltip title="主题切换">
          <Box sx={{ '& svg': { fontSize: 22 } }}>
            <ToggleThemeButton />
          </Box>
        </Tooltip>
        <Tooltip title="系统设置">
          <IconButton size="large">
            <SettingsOutlinedIcon />
          </IconButton>
        </Tooltip>
      </Stack>
    </Box>
  </AppBar>
);
