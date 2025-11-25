import { ReactNode } from 'react';
import { Box } from '@mui/material';
import { ColumnConfig } from './FieldGrid';

export interface FieldGridItemProps {
  children: ReactNode;
  span?: ColumnConfig;
}

/**
 * 字段网格项组件
 * 支持跨列配置，用于控制表单字段在网格中的宽度
 */
export const FieldGridItem = ({
  children,
  span = {}
}: FieldGridItemProps) => {
  const resolved = {
    xs: span.xs ?? 1,
    sm: span.sm ?? span.xs ?? 1,
    md: span.md ?? span.sm ?? span.xs ?? 1,
    lg: span.lg ?? span.md ?? span.sm ?? span.xs ?? 1
  };

  return (
    <Box
      sx={{
        width: '100%',
        gridColumn: {
          xs: `span ${resolved.xs}`,
          sm: `span ${resolved.sm}`,
          md: `span ${resolved.md}`,
          lg: `span ${resolved.lg}`
        }
      }}
    >
      {children}
    </Box>
  );
};

export default FieldGridItem;
