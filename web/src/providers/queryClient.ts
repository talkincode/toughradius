import { QueryClient } from '@tanstack/react-query';
import { ApiError } from '../utils/apiClient';

const shouldRetryRequest = (failureCount: number, error: unknown) => {
  if (error instanceof ApiError) {
    if (error.status === 401 || error.status === 403) {
      return false;
    }
  }
  return failureCount < 2;
};

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      refetchOnReconnect: true,
      retry: shouldRetryRequest,
      staleTime: 30 * 1000,
      gcTime: 5 * 60 * 1000,
    },
    mutations: {
      retry: shouldRetryRequest,
    },
  },
});
