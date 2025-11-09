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
        backgroundColor: '#f5f7fa',
        '& .RaLayout-content': {
          backgroundColor: '#f5f7fa',
          padding: { xs: 2, md: 3, lg: 4 },
          minHeight: 'calc(100vh - 64px)',
        },
      },
      ...(Array.isArray(sx) ? sx : sx ? [sx] : []),
    ]}
  />
);
