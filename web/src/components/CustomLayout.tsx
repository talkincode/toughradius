import type { SxProps, Theme } from '@mui/material';
import { Layout, LayoutProps } from 'react-admin';

import { CustomAppBar } from './CustomAppBar';
import { CustomMenu } from './CustomMenu';

type CustomLayoutProps = LayoutProps & { sx?: SxProps<Theme> };

export const CustomLayout = ({ sx, ...rest }: CustomLayoutProps) => (
  <Layout
    {...rest}
    appBar={CustomAppBar}
    menu={CustomMenu}
    sx={[
      {
        // 固定顶部 AppBar（滚动时不隐藏）
        '& .MuiAppBar-root': {
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          zIndex: 1200,
        },
        // 为固定的 AppBar 留出空间
        '& .RaLayout-appFrame': {
          marginTop: '48px',
        },
        // 固定左侧菜单（滚动时不跟随移动）
        '& .RaSidebar-fixed': {
          position: 'fixed',
          top: '48px',
          height: 'calc(100vh - 48px)',
          overflowY: 'auto',
        },
        // 内容区域样式
        '& .RaLayout-content': {
          padding: { xs: 2, md: 3, lg: 4 },
          minHeight: 'calc(100vh - 48px)',
          transition: 'background-color 0.3s ease',
        },
      },
      ...(Array.isArray(sx) ? sx : sx ? [sx] : []),
    ]}
  />
);
