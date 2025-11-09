import DashboardOutlinedIcon from '@mui/icons-material/DashboardOutlined';
import PeopleAltOutlinedIcon from '@mui/icons-material/PeopleAltOutlined';
import SensorsOutlinedIcon from '@mui/icons-material/SensorsOutlined';
import ReceiptLongOutlinedIcon from '@mui/icons-material/ReceiptLongOutlined';
import SettingsSuggestOutlinedIcon from '@mui/icons-material/SettingsSuggestOutlined';
import { Box } from '@mui/material';
import { MenuItemLink, MenuProps } from 'react-admin';

const menuItems = [
  { to: '/', label: '控制台', icon: <DashboardOutlinedIcon /> },
  { to: '/radius/users', label: '用户管理', icon: <PeopleAltOutlinedIcon /> },
  { to: '/radius/profiles', label: '策略管理', icon: <SettingsSuggestOutlinedIcon /> },
  { to: '/radius/online', label: '在线会话', icon: <SensorsOutlinedIcon /> },
  { to: '/radius/accounting', label: '计费日志', icon: <ReceiptLongOutlinedIcon /> },
];

export const CustomMenu = ({ dense, onMenuClick, logout }: MenuProps) => {
  const currentYear = new Date().getFullYear();

  return (
    <Box
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        backgroundColor: '#efefefff',
        color: '#373737ff',
        pt: 0,
      }}
    >

      <Box sx={{ flexGrow: 1, overflowY: 'auto', pt: 1, marginTop: 2 }}>
        {menuItems.map((item) => (
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
          borderTop: '1px solid rgba(255, 255, 255, 0.08)',
          textAlign: 'center',
          px: 2,
          py: 3,
          fontSize: 12,
          color: 'rgba(72, 72, 72, 0.65)',
        }}
      >
        <div>ToughRADIUS v9</div>
        <div>© {currentYear} ALL RIGHTS RESERVED</div>
        {logout && <Box sx={{ mt: 2 }}>{logout}</Box>}
      </Box>
    </Box>
  );
};
