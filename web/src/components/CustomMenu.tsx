import DashboardOutlinedIcon from '@mui/icons-material/DashboardOutlined';
import PeopleAltOutlinedIcon from '@mui/icons-material/PeopleAltOutlined';
import SensorsOutlinedIcon from '@mui/icons-material/SensorsOutlined';
import ReceiptLongOutlinedIcon from '@mui/icons-material/ReceiptLongOutlined';
import SettingsSuggestOutlinedIcon from '@mui/icons-material/SettingsSuggestOutlined';
import SettingsOutlinedIcon from '@mui/icons-material/SettingsOutlined';
import RouterOutlinedIcon from '@mui/icons-material/RouterOutlined';
import AccountTreeOutlinedIcon from '@mui/icons-material/AccountTreeOutlined';
import AdminPanelSettingsOutlinedIcon from '@mui/icons-material/AdminPanelSettingsOutlined';
import { Box, useTheme } from '@mui/material';
import { MenuItemLink, MenuProps, useGetIdentity, useTranslate } from 'react-admin';

const menuItems = [
  { to: '/', labelKey: 'menu.dashboard', icon: <DashboardOutlinedIcon /> },
  { to: '/network/nodes', labelKey: 'menu.network_nodes', icon: <AccountTreeOutlinedIcon /> },
  { to: '/network/nas', labelKey: 'menu.nas_devices', icon: <RouterOutlinedIcon /> },
  { to: '/radius/users', labelKey: 'menu.radius_users', icon: <PeopleAltOutlinedIcon /> },
  { to: '/radius/profiles', labelKey: 'menu.radius_profiles', icon: <SettingsSuggestOutlinedIcon /> },
  { to: '/radius/online', labelKey: 'menu.online_sessions', icon: <SensorsOutlinedIcon /> },
  { to: '/radius/accounting', labelKey: 'menu.accounting', icon: <ReceiptLongOutlinedIcon /> },
  { to: '/system/config', labelKey: 'menu.system_config', icon: <SettingsOutlinedIcon />, permissions: ['super', 'admin'] },
  { to: '/system/operators', labelKey: 'menu.operators', icon: <AdminPanelSettingsOutlinedIcon />, permissions: ['super', 'admin'] },
];

export const CustomMenu = ({ dense, onMenuClick, logout }: MenuProps) => {
  const currentYear = new Date().getFullYear();
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const { data: identity } = useGetIdentity();
  const translate = useTranslate();

  // 根据用户权限过滤菜单项
  const filteredMenuItems = menuItems.filter(item => {
    if (!item.permissions) return true; // 无权限限制的菜单项对所有人可见
    if (!identity?.level) return false; // 未登录用户不显示需要权限的菜单
    return item.permissions.includes(identity.level); // 检查用户权限是否在允许列表中
  });

  return (
    <Box
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        // 侧边栏根据主题使用不同背景色
        backgroundColor: isDark ? '#1e293b' : '#1e40af',
        color: '#ffffff',
        pt: 0,
        transition: 'background-color 0.3s ease',
      }}
    >
      <Box sx={{ flexGrow: 1, overflowY: 'auto', pt: 1, marginTop: 2 }}>
        {filteredMenuItems.map((item) => (
          <MenuItemLink
            key={item.to}
            to={item.to}
            primaryText={translate(item.labelKey)}
            leftIcon={item.icon}
            dense={dense}
            onClick={onMenuClick}
          />
        ))}
      </Box>

      <Box
        sx={{
          borderTop: '1px solid rgba(255, 255, 255, 0.1)',
          textAlign: 'center',
          px: 2,
          py: 3,
          fontSize: 12,
          color: 'rgba(255, 255, 255, 0.6)',
          transition: 'all 0.3s ease',
        }}
      >
        <div style={{ fontWeight: 600, marginBottom: 4 }}>ToughRADIUS v9</div>
        <div>© {currentYear} ALL RIGHTS RESERVED</div>
        {logout && <Box sx={{ mt: 2 }}>{logout}</Box>}
      </Box>
    </Box>
  );
};
