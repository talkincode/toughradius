import { alpha, createTheme, PaletteMode } from '@mui/material/styles';

// 浅色主题配色
const lightPalette = {
  primary: {
    main: '#2563eb',
    light: '#60a5fa',
    dark: '#1e40af',
    contrastText: '#ffffff',
  },
  secondary: {
    main: '#8b5cf6',
    light: '#a78bfa',
    dark: '#7c3aed',
    contrastText: '#ffffff',
  },
  success: {
    main: '#10b981',
    light: '#34d399',
    dark: '#059669',
  },
  warning: {
    main: '#f59e0b',
    light: '#fbbf24',
    dark: '#d97706',
  },
  error: {
    main: '#ef4444',
    light: '#f87171',
    dark: '#dc2626',
  },
  info: {
    main: '#06b6d4',
    light: '#22d3ee',
    dark: '#0891b2',
  },
  background: {
    default: '#f8fafc',
    paper: '#ffffff',
  },
  text: {
    primary: '#0f172a',
    secondary: '#64748b',
    disabled: '#94a3b8',
  },
  divider: 'rgba(100, 116, 139, 0.12)',
};

// 深色主题配色
const darkPalette = {
  primary: {
    main: '#3b82f6',
    light: '#60a5fa',
    dark: '#2563eb',
    contrastText: '#ffffff',
  },
  secondary: {
    main: '#a78bfa',
    light: '#c4b5fd',
    dark: '#8b5cf6',
    contrastText: '#ffffff',
  },
  success: {
    main: '#22c55e',
    light: '#4ade80',
    dark: '#16a34a',
  },
  warning: {
    main: '#fbbf24',
    light: '#fcd34d',
    dark: '#f59e0b',
  },
  error: {
    main: '#f87171',
    light: '#fca5a5',
    dark: '#ef4444',
  },
  info: {
    main: '#22d3ee',
    light: '#67e8f9',
    dark: '#06b6d4',
  },
  background: {
    default: '#0f172a',
    paper: '#1e293b',
  },
  text: {
    primary: '#f1f5f9',
    secondary: '#cbd5e1',
    disabled: '#64748b',
  },
  divider: 'rgba(148, 163, 184, 0.12)',
};

/**
 * 创建主题配置
 * @param mode 主题模式：'light' | 'dark'
 */
export const createAppTheme = (mode: PaletteMode) => {
  const isDark = mode === 'dark';
  const palette = isDark ? darkPalette : lightPalette;

  return createTheme({
    palette: {
      mode,
      ...palette,
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
        '"Microsoft YaHei"',
        'sans-serif',
      ].join(','),
      h1: { 
        fontWeight: 700, 
        fontSize: '2.5rem',
        lineHeight: 1.2,
      },
      h2: { 
        fontWeight: 700, 
        fontSize: '2rem',
        lineHeight: 1.3,
      },
      h3: { 
        fontWeight: 600, 
        fontSize: '1.75rem',
        lineHeight: 1.3,
      },
      h4: { 
        fontWeight: 600, 
        fontSize: '1.5rem',
        lineHeight: 1.4,
      },
      h5: { 
        fontWeight: 600, 
        fontSize: '1.25rem',
        lineHeight: 1.4,
      },
      h6: { 
        fontWeight: 600, 
        fontSize: '1.125rem',
        lineHeight: 1.5,
      },
      button: {
        fontWeight: 600,
        textTransform: 'none',
      },
    },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          body: {
            backgroundColor: palette.background.default,
            color: palette.text.primary,
            transition: 'background-color 0.3s ease, color 0.3s ease',
          },
          '#root': {
            backgroundColor: palette.background.default,
          },
          '*::-webkit-scrollbar': {
            width: '8px',
            height: '8px',
          },
          '*::-webkit-scrollbar-track': {
            background: isDark ? '#1e293b' : '#f1f5f9',
          },
          '*::-webkit-scrollbar-thumb': {
            background: isDark ? '#475569' : '#cbd5e1',
            borderRadius: '4px',
            '&:hover': {
              background: isDark ? '#64748b' : '#94a3b8',
            },
          },
          // 强制菜单图标始终使用浅色
          '.RaMenuItemLink-icon': {
            color: 'rgba(255, 255, 255, 0.85) !important',
          },
          '.RaMenuItemLink-root:hover .RaMenuItemLink-icon': {
            color: '#ffffff !important',
          },
          '.RaMenuItemLink-active .RaMenuItemLink-icon': {
            color: '#ffffff !important',
          },
        },
      },
      MuiPaper: {
        styleOverrides: {
          root: {
            borderRadius: 12,
            border: `1px solid ${isDark ? 'rgba(148, 163, 184, 0.1)' : 'rgba(100, 116, 139, 0.12)'}`,
            boxShadow: isDark 
              ? '0 1px 3px 0 rgba(0, 0, 0, 0.3)' 
              : '0 1px 3px 0 rgba(0, 0, 0, 0.05)',
            backgroundImage: 'none',
            transition: 'all 0.3s ease',
          },
          elevation1: {
            boxShadow: isDark
              ? '0 2px 4px 0 rgba(0, 0, 0, 0.4)'
              : '0 2px 4px 0 rgba(0, 0, 0, 0.06)',
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
            border: `1px solid ${isDark ? 'rgba(148, 163, 184, 0.1)' : 'rgba(100, 116, 139, 0.12)'}`,
            boxShadow: isDark
              ? '0 1px 3px 0 rgba(0, 0, 0, 0.3)'
              : '0 1px 3px 0 rgba(0, 0, 0, 0.05)',
            transition: 'all 0.3s ease',
            '&:hover': {
              boxShadow: isDark
                ? '0 4px 12px 0 rgba(0, 0, 0, 0.4)'
                : '0 4px 12px 0 rgba(0, 0, 0, 0.08)',
            },
          },
        },
      },
      MuiButton: {
        defaultProps: {
          disableElevation: true,
        },
        styleOverrides: {
          root: {
            borderRadius: 8,
            textTransform: 'none',
            fontWeight: 600,
            paddingInline: 20,
            paddingBlock: 10,
            transition: 'all 0.2s ease',
          },
          contained: {
            boxShadow: 'none',
            '&:hover': {
              boxShadow: isDark
                ? '0 4px 8px 0 rgba(0, 0, 0, 0.3)'
                : '0 4px 8px 0 rgba(0, 0, 0, 0.1)',
            },
          },
        },
      },
      MuiIconButton: {
        styleOverrides: {
          root: {
            borderRadius: 8,
            transition: 'all 0.2s ease',
            '&:hover': {
              backgroundColor: alpha(palette.primary.main, 0.08),
            },
          },
        },
      },
      MuiTextField: {
        styleOverrides: {
          root: {
            '& .MuiOutlinedInput-root': {
              borderRadius: 8,
              transition: 'all 0.2s ease',
              '&:hover .MuiOutlinedInput-notchedOutline': {
                borderColor: palette.primary.main,
              },
            },
          },
        },
      },
      MuiAppBar: {
        styleOverrides: {
          root: {
            // 浅色主题使用浅灰色，深色主题使用深色
            backgroundColor: isDark ? '#1e293b' : '#f1f5f9',
            color: isDark ? '#f1f5f9' : '#0f172a',
            borderBottom: isDark 
              ? '1px solid rgba(148, 163, 184, 0.2)' 
              : '1px solid rgba(100, 116, 139, 0.12)',
            boxShadow: 'none',
            transition: 'all 0.3s ease',
          },
        },
      },
      MuiDrawer: {
        styleOverrides: {
          paper: {
            // 侧边栏根据主题使用不同的背景色
            backgroundColor: isDark ? '#1e293b' : '#1e40af',
            borderRight: 'none',
            transition: 'all 0.3s ease',
          },
        },
      },
      MuiTableCell: {
        styleOverrides: {
          root: {
            borderBottom: `1px solid ${palette.divider}`,
          },
          head: {
            fontWeight: 600,
            backgroundColor: isDark ? '#1e293b' : '#f8fafc',
            color: palette.text.primary,
          },
        },
      },
      MuiChip: {
        styleOverrides: {
          root: {
            borderRadius: 6,
            fontWeight: 500,
          },
        },
      },
      RaMenuItemLink: {
        styleOverrides: {
          root: {
            borderRadius: 8,
            marginInline: 8,
            marginBlock: 4,
            paddingBlock: 10,
            paddingInline: 12,
            // 菜单项始终使用浅色文字
            color: 'rgba(255, 255, 255, 0.9) !important',
            transition: 'all 0.2s ease',
            '& .RaMenuItemLink-label': {
              fontWeight: 600,
              fontSize: '0.95rem',
              color: 'rgba(255, 255, 255, 0.9) !important',
            },
            '& .RaMenuItemLink-icon': {
              color: 'rgba(255, 255, 255, 0.85) !important',
            },
            '&:hover': {
              backgroundColor: 'rgba(255, 255, 255, 0.12)',
              color: '#ffffff !important',
              '& .RaMenuItemLink-label': {
                color: '#ffffff !important',
              },
              '& .RaMenuItemLink-icon': {
                color: '#ffffff !important',
                transform: 'scale(1.05)',
              },
            },
            '&.RaMenuItemLink-active': {
              backgroundColor: 'rgba(59, 130, 246, 0.4)',
              color: '#ffffff !important',
              fontWeight: 700,
              '& .RaMenuItemLink-label': {
                color: '#ffffff !important',
              },
              '& .RaMenuItemLink-icon': {
                color: '#ffffff !important',
              },
            },
          },
          icon: {
            // 菜单图标始终使用浅色，添加 !important 确保优先级
            color: 'rgba(255, 255, 255, 0.85) !important',
            transition: 'all 0.2s ease',
          },
        },
      },
      RaLayout: {
        styleOverrides: {
          root: {
            backgroundColor: palette.background.default,
            transition: 'background-color 0.3s ease',
          },
        },
      },
      RaDatagrid: {
        styleOverrides: {
          root: {
            backgroundColor: palette.background.paper,
            '& .RaDatagrid-headerCell': {
              fontWeight: 700,
              backgroundColor: isDark ? '#1e293b' : '#f8fafc',
            },
            '& .RaDatagrid-row': {
              transition: 'background-color 0.15s ease',
              '&:hover': {
                backgroundColor: isDark 
                  ? alpha(palette.primary.main, 0.08)
                  : alpha(palette.primary.main, 0.04),
              },
            },
          },
        },
      },
    },
  });
};

// 默认导出浅色主题（向后兼容）
export const theme = createAppTheme('light');

// 深色主题
export const darkTheme = createAppTheme('dark');
