import { useQuery, UseQueryOptions, QueryKey } from '@tanstack/react-query';
import { ApiError, apiRequest } from '../utils/apiClient';

type ApiQueryOptions<TData> = Omit<UseQueryOptions<TData, ApiError, TData, QueryKey>, 'queryKey' | 'queryFn'> & {
  path: string;
  queryKey: QueryKey;
  requestInit?: RequestInit;
};

export const useApiQuery = <TData = unknown>({
  path,
  queryKey,
  requestInit,
  ...options
}: ApiQueryOptions<TData>) =>
  useQuery<TData, ApiError, TData, QueryKey>({
    queryKey,
    queryFn: ({ signal }) => apiRequest<TData>(path, { ...requestInit, signal }),
    ...options,
  });
