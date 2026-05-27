import { Pagination, PaginationProps } from 'react-admin';

export const DEFAULT_PAGE_SIZES = [25, 50, 100, 200];

export const ServerPagination = (props: PaginationProps) => (
  <Pagination rowsPerPageOptions={DEFAULT_PAGE_SIZES} {...props} />
);
