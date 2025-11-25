import { useCallback, useMemo } from 'react';
import { Box, Chip, Typography, IconButton, Tooltip } from '@mui/material';
import { Clear as ClearIcon } from '@mui/icons-material';
import { useTranslate, useListContext } from 'react-admin';

interface ActiveFiltersProps {
  /**
   * Map of field names to their display labels.
   * If not provided, the field name will be used as-is.
   */
  fieldLabels?: Record<string, string>;
  /**
   * Map of field values to their display labels.
   * Useful for enum fields like status.
   */
  valueLabels?: Record<string, Record<string, string>>;
  /**
   * Fields to exclude from display
   */
  excludeFields?: string[];
}

/**
 * ActiveFilters displays the currently active filter conditions.
 * Shows a chip for each filter that can be individually removed.
 * 
 * @param fieldLabels - Optional map of field names to display labels
 * @param valueLabels - Optional map of field values to display labels
 * @param excludeFields - Optional list of fields to exclude from display
 */
export const ActiveFilters = ({
  fieldLabels = {},
  valueLabels = {},
  excludeFields = ['q'], // Exclude live search by default
}: ActiveFiltersProps) => {
  const translate = useTranslate();
  const { filterValues, setFilters, displayedFilters } = useListContext();

  const activeFilters = useMemo(() => {
    if (!filterValues) return [];

    return Object.entries(filterValues)
      .filter(([key, value]) => {
        // Exclude specified fields
        if (excludeFields.includes(key)) return false;
        // Exclude empty values
        if (value === undefined || value === null || value === '') return false;
        return true;
      })
      .map(([key, value]) => {
        const label = fieldLabels[key] || key;
        let displayValue = String(value);

        // Check for value labels
        if (valueLabels[key] && valueLabels[key][value as string]) {
          displayValue = valueLabels[key][value as string];
        }

        return {
          field: key,
          label,
          value,
          displayValue,
        };
      });
  }, [filterValues, fieldLabels, valueLabels, excludeFields]);

  const handleRemoveFilter = useCallback(
    (field: string) => {
      const newFilters = { ...filterValues };
      delete newFilters[field];
      setFilters(newFilters, displayedFilters);
    },
    [filterValues, setFilters, displayedFilters]
  );

  const handleClearAll = useCallback(() => {
    // Keep only excluded fields (like q for live search)
    const preservedFilters = excludeFields.reduce(
      (acc, field) => {
        if (filterValues && filterValues[field] !== undefined) {
          acc[field] = filterValues[field];
        }
        return acc;
      },
      {} as Record<string, unknown>
    );
    setFilters(preservedFilters, displayedFilters);
  }, [filterValues, setFilters, displayedFilters, excludeFields]);

  if (activeFilters.length === 0) {
    return null;
  }

  return (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        gap: 1,
        flexWrap: 'wrap',
        mb: 2,
        p: 1.5,
        backgroundColor: theme =>
          theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.02)',
        borderRadius: 1,
        border: '1px solid',
        borderColor: 'divider',
      }}
    >
      <Typography
        variant="body2"
        color="text.secondary"
        sx={{ fontWeight: 500, mr: 1 }}
      >
        {translate('filter.active_filters', { _: '当前筛选' })}:
      </Typography>

      {activeFilters.map(({ field, label, displayValue }) => (
        <Chip
          key={field}
          label={`${label}: ${displayValue}`}
          size="small"
          onDelete={() => handleRemoveFilter(field)}
          color="primary"
          variant="outlined"
          sx={{
            borderRadius: 1,
            '& .MuiChip-deleteIcon': {
              fontSize: 16,
            },
          }}
        />
      ))}

      {activeFilters.length > 1 && (
        <Tooltip title={translate('filter.clear_all', { _: '清除全部' })}>
          <IconButton
            size="small"
            onClick={handleClearAll}
            sx={{ ml: 1 }}
          >
            <ClearIcon fontSize="small" />
          </IconButton>
        </Tooltip>
      )}
    </Box>
  );
};

export default ActiveFilters;
