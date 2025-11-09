import DashboardOutlinedIcon from '@mui/icons-material/DashboardOutlined';
import PeopleAltOutlinedIcon from '@mui/icons-material/PeopleAltOutlined';
import SensorsOutlinedIcon from '@mui/icons-material/SensorsOutlined';
import ReceiptLongOutlinedIcon from '@mui/icons-material/ReceiptLongOutlined';
import SettingsSuggestOutlinedIcon from '@mui/icons-material/SettingsSuggestOutlined';
import RouterOutlinedIcon from '@mui/icons-material/RouterOutlined';
import AccountTreeOutlinedIcon from '@mui/icons-material/AccountTreeOutlined';
import AdminPanelSettingsOutlinedIcon from '@mui/icons-material/AdminPanelSettingsOutlined';
import { Box, useTheme } from '@mui/material';
import { MenuItemLink, MenuProps, useGetIdentity } from 'react-admin';

const menuItems = [
  { to: '/', label: '控制台', icon: <DashboardOutlinedIcon /> },
  { to: '/network/nodes', label: '网络节点', icon: <AccountTreeOutlinedIcon /> },
  { to: '/network/nas', label: 'NAS设备', icon: <RouterOutlinedIcon /> },
  { to: '/radius/users', label: '用户管理', icon: <PeopleAltOutlinedIcon /> },
  { to: '/radius/profiles', label: '策略管理', icon: <SettingsSuggestOutlinedIcon /> },
  { to: '/radius/online', label: '在线会话', icon: <SensorsOutlinedIcon /> },
  { to: '/radius/accounting', label: '计费日志', icon: <ReceiptLongOutlinedIcon /> },
  { to: '/system/operators', label: '操作员管理', icon: <AdminPanelSettingsOutlinedIcon />, permissions: ['super', 'admin'] },
];

export const CustomMenu = ({ dense, onMenuClick, logout }: MenuProps) => {
  const currentYear = new Date().getFullYear();
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const { data: identity } = useGetIdentity();

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
            primaryText={item.label}
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
