import { alpha, createTheme } from '@mui/material/styles';

const primaryMain = '#1976d2';
const secondaryMain = '#9c27b0';
const backgroundDefault = '#f5f7fa';
const textPrimary = '#2c3e50';

export const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: primaryMain,
      light: '#63a4ff',
      dark: '#004ba0',
      contrastText: '#ffffff',
    },
    secondary: {
      main: secondaryMain,
      light: '#d05ce3',
      dark: '#6a0080',
      contrastText: '#ffffff',
    },
    background: {
      default: backgroundDefault,
      paper: '#ffffff',
    },
    text: {
      primary: textPrimary,
      secondary: '#6b7280',
    },
  },
  shape: {
    borderRadius: 8,
  },
  typography: {
    fontFamily: [
      '-apple-system',
      'BlinkMacSystemFont',
      '"Segoe UI"',
      'Roboto',
      '"Helvetica Neue"',
      'Arial',
      'sans-serif',
    ].join(','),
    h1: { fontWeight: 600, fontSize: '2.25rem' },
    h2: { fontWeight: 600, fontSize: '1.75rem' },
    h3: { fontWeight: 600, fontSize: '1.5rem' },
    h4: { fontWeight: 600, fontSize: '1.25rem' },
    h5: { fontWeight: 600, fontSize: '1.125rem' },
    h6: { fontWeight: 600, fontSize: '1rem' },
    subtitle1: { color: '#4b5563' },
    subtitle2: { color: '#64748b' },
    body1: { color: textPrimary },
    body2: { color: '#475569' },
  },
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        body: {
          backgroundColor: backgroundDefault,
          color: textPrimary,
        },
        '#root': {
          backgroundColor: backgroundDefault,
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          borderRadius: 12,
          border: '1px solid rgba(148, 163, 184, 0.15)',
          boxShadow: 'none',
          backgroundImage: 'none',
        },
      },
    },
    MuiCard: {
      defaultProps: {
        elevation: 0,
      },
      styleOverrides: {
        root: {
          borderRadius: 16,
          border: '1px solid rgba(148, 163, 184, 0.2)',
          boxShadow: 'none',
        },
      },
    },
    MuiButton: {
      defaultProps: {
        disableElevation: true,
      },
      styleOverrides: {
        root: {
          borderRadius: 999,
          textTransform: 'none',
          fontWeight: 600,
          paddingInline: 24,
          paddingBlock: 10,
        },
      },
    },
    MuiIconButton: {
      styleOverrides: {
        root: {
          borderRadius: 10,
          backgroundColor: 'transparent',
          '&:hover': {
            backgroundColor: alpha(primaryMain, 0.08),
          },
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          backgroundColor: '#ffffff',
          color: textPrimary,
          borderBottom: '1px solid rgba(15, 23, 42, 0.08)',
          boxShadow: 'none',
        },
      },
    },
    RaMenuItemLink: {
      styleOverrides: {
        root: {
          borderRadius: 7,
          marginInline: 7,
          marginBlock: 4,
          paddingBlock: 7,
          paddingInline: 9,
          color: 'rgba(25, 25, 25, 0.85)',
          transition: 'all 0.3s ease',
          '& .RaMenuItemLink-label': {
            fontWeight: 600,
          },
          '&:hover': {
            backgroundColor: 'rgba(255, 255, 255, 0.16)',
            color: '#494949ff',
          },
          '&.RaMenuItemLink-active': {
            backgroundColor: '#1890ff',
            color: '#ddddddff',
          },
        },
        icon: {
          color: 'inherit',
        },
      },
    },
    RaLayout: {
      styleOverrides: {
        root: {
          backgroundColor: backgroundDefault,
        },
      },
    },
  },
});
