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
        '& .RaLayout-content': {
          padding: { xs: 2, md: 3, lg: 4 },
          minHeight: 'calc(100vh - 64px)',
          transition: 'background-color 0.3s ease',
        },
      },
      ...(Array.isArray(sx) ? sx : sx ? [sx] : []),
    ]}
  />
);
