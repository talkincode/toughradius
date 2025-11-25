import { Skeleton, Box, Card, CardContent, Stack, Table, TableBody, TableRow, TableCell } from '@mui/material';

/**
 * ListSkeleton displays a loading skeleton for list views.
 * Shows a placeholder for search bar, filter buttons, and table rows.
 */
export const ListSkeleton = ({ rows = 5 }: { rows?: number }) => (
  <Box sx={{ width: '100%', p: 2 }}>
    {/* Header skeleton with search and buttons */}
    <Box
      sx={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        mb: 3,
      }}
    >
      <Skeleton variant="rectangular" width={200} height={40} sx={{ borderRadius: 1 }} />
      <Box sx={{ display: 'flex', gap: 1 }}>
        <Skeleton variant="rectangular" width={80} height={36} sx={{ borderRadius: 1 }} />
        <Skeleton variant="rectangular" width={80} height={36} sx={{ borderRadius: 1 }} />
        <Skeleton variant="rectangular" width={80} height={36} sx={{ borderRadius: 1 }} />
      </Box>
    </Box>

    {/* Table header skeleton */}
    <Box
      sx={{
        display: 'grid',
        gridTemplateColumns: 'repeat(6, 1fr)',
        gap: 2,
        mb: 2,
        p: 1,
        backgroundColor: 'grey.100',
        borderRadius: 1,
      }}
    >
      {[...Array(6)].map((_, i) => (
        <Skeleton key={i} variant="text" height={24} />
      ))}
    </Box>

    {/* Table rows skeleton */}
    {[...Array(rows)].map((_, rowIndex) => (
      <Box
        key={rowIndex}
        sx={{
          display: 'grid',
          gridTemplateColumns: 'repeat(6, 1fr)',
          gap: 2,
          py: 1.5,
          px: 1,
          borderBottom: '1px solid',
          borderColor: 'divider',
        }}
      >
        {[...Array(6)].map((_, colIndex) => (
          <Skeleton
            key={colIndex}
            variant="text"
            height={20}
            width={colIndex === 0 ? '80%' : '60%'}
          />
        ))}
      </Box>
    ))}

    {/* Pagination skeleton */}
    <Box
      sx={{
        display: 'flex',
        justifyContent: 'flex-end',
        alignItems: 'center',
        mt: 2,
        gap: 2,
      }}
    >
      <Skeleton variant="text" width={100} />
      <Box sx={{ display: 'flex', gap: 1 }}>
        <Skeleton variant="circular" width={32} height={32} />
        <Skeleton variant="circular" width={32} height={32} />
      </Box>
    </Box>
  </Box>
);

/**
 * FormSkeleton displays a loading skeleton for form views (create/edit).
 * Shows placeholders for form sections and input fields.
 */
export const FormSkeleton = ({ sections = 3 }: { sections?: number }) => (
  <Box sx={{ width: '100%', maxWidth: 800, mx: 'auto', p: 2 }}>
    {[...Array(sections)].map((_, sectionIndex) => (
      <Card
        key={sectionIndex}
        elevation={0}
        sx={{
          mb: 3,
          border: '1px solid',
          borderColor: 'divider',
          borderRadius: 2,
        }}
      >
        <CardContent>
          {/* Section title */}
          <Skeleton variant="text" width={150} height={28} sx={{ mb: 0.5 }} />
          <Skeleton variant="text" width={250} height={20} sx={{ mb: 2 }} />

          {/* Form fields */}
          <Stack spacing={2}>
            <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 2 }}>
              {[...Array(4)].map((_, fieldIndex) => (
                <Box key={fieldIndex}>
                  <Skeleton variant="text" width={80} height={16} sx={{ mb: 0.5 }} />
                  <Skeleton
                    variant="rectangular"
                    height={40}
                    sx={{ borderRadius: 1 }}
                  />
                </Box>
              ))}
            </Box>
          </Stack>
        </CardContent>
      </Card>
    ))}

    {/* Toolbar skeleton */}
    <Box sx={{ display: 'flex', gap: 2, justifyContent: 'flex-end', mt: 2 }}>
      <Skeleton variant="rectangular" width={80} height={36} sx={{ borderRadius: 1 }} />
      <Skeleton variant="rectangular" width={80} height={36} sx={{ borderRadius: 1 }} />
    </Box>
  </Box>
);

/**
 * ShowSkeleton displays a loading skeleton for show/detail views.
 * Shows placeholders for detail cards and tables.
 */
export const ShowSkeleton = () => (
  <Box sx={{ width: '100%', p: { xs: 2, sm: 3, md: 4 } }}>
    <Stack spacing={3}>
      {/* Action bar skeleton */}
      <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 1 }}>
        <Skeleton variant="rectangular" width={80} height={36} sx={{ borderRadius: 1 }} />
        <Skeleton variant="rectangular" width={80} height={36} sx={{ borderRadius: 1 }} />
        <Skeleton variant="rectangular" width={80} height={36} sx={{ borderRadius: 1 }} />
      </Box>

      {/* Main info card skeleton */}
      <Card elevation={2}>
        <CardContent>
          <Skeleton variant="text" width={150} height={32} sx={{ mb: 2 }} />
          <Box sx={{ borderTop: '1px solid', borderColor: 'divider', pt: 2 }}>
            <Table>
              <TableBody>
                {[...Array(4)].map((_, i) => (
                  <TableRow key={i}>
                    <TableCell sx={{ width: '30%', backgroundColor: 'grey.50' }}>
                      <Skeleton variant="text" width={100} />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="text" width="60%" />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Box>
        </CardContent>
      </Card>

      {/* Two column cards skeleton */}
      <Box sx={{ display: 'grid', gap: 3, gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' } }}>
        {[...Array(2)].map((_, cardIndex) => (
          <Card key={cardIndex} elevation={2}>
            <CardContent>
              <Skeleton variant="text" width={120} height={28} sx={{ mb: 2 }} />
              <Box sx={{ borderTop: '1px solid', borderColor: 'divider', pt: 2 }}>
                <Table>
                  <TableBody>
                    {[...Array(3)].map((_, i) => (
                      <TableRow key={i}>
                        <TableCell sx={{ width: '40%', backgroundColor: 'grey.50' }}>
                          <Skeleton variant="text" width={80} />
                        </TableCell>
                        <TableCell>
                          <Skeleton variant="text" width="50%" />
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </Box>
            </CardContent>
          </Card>
        ))}
      </Box>
    </Stack>
  </Box>
);

/**
 * DashboardSkeleton displays a loading skeleton for the dashboard.
 * Shows placeholders for stat cards and charts.
 */
export const DashboardSkeleton = () => (
  <Box sx={{ width: '100%', p: 2 }}>
    {/* Title skeleton */}
    <Skeleton variant="text" width={200} height={40} sx={{ mb: 1 }} />
    <Skeleton variant="text" width={400} height={24} sx={{ mb: 3 }} />

    {/* Stat cards skeleton */}
    <Box
      sx={{
        display: 'grid',
        gap: 2,
        gridTemplateColumns: { xs: '1fr', sm: 'repeat(2, 1fr)', md: 'repeat(4, 1fr)' },
        mb: 3,
      }}
    >
      {[...Array(4)].map((_, i) => (
        <Card key={i} elevation={1}>
          <CardContent>
            <Skeleton variant="circular" width={40} height={40} sx={{ mb: 1 }} />
            <Skeleton variant="text" width={80} height={20} />
            <Skeleton variant="text" width={60} height={32} />
          </CardContent>
        </Card>
      ))}
    </Box>

    {/* Chart cards skeleton */}
    <Box
      sx={{
        display: 'grid',
        gap: 2,
        gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)' },
      }}
    >
      {[...Array(2)].map((_, i) => (
        <Card key={i} elevation={1}>
          <CardContent>
            <Skeleton variant="text" width={150} height={28} sx={{ mb: 2 }} />
            <Skeleton variant="rectangular" height={200} sx={{ borderRadius: 1 }} />
          </CardContent>
        </Card>
      ))}
    </Box>
  </Box>
);

export default {
  ListSkeleton,
  FormSkeleton,
  ShowSkeleton,
  DashboardSkeleton,
};
