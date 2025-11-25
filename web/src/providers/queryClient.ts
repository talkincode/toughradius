import { QueryClient } from '@tanstack/react-query';
import { ApiError } from '../utils/apiClient';

/**
 * Determines whether a failed request should be retried.
 * 
 * Retry strategy:
 * - Do not retry authentication errors (401, 403)
 * - Retry network errors and server errors up to 3 times
 * - Use exponential backoff between retries
 */
const shouldRetryRequest = (failureCount: number, error: unknown) => {
  // Don't retry authentication errors
  if (error instanceof ApiError) {
    if (error.status === 401 || error.status === 403 || error.status === 404) {
      return false;
    }
  }
  
  // Retry network errors and 5xx errors up to 3 times
  return failureCount < 3;
};

/**
 * Calculate retry delay with exponential backoff.
 * Base delay is 1 second, doubles with each retry.
 * Max delay capped at 30 seconds.
 */
const getRetryDelay = (attemptIndex: number) => {
  const baseDelay = 1000; // 1 second
  const maxDelay = 30000; // 30 seconds
  const delay = Math.min(baseDelay * Math.pow(2, attemptIndex), maxDelay);
  // Add some jitter (±25%)
  const jitter = delay * 0.25 * (Math.random() - 0.5);
  return delay + jitter;
};

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      refetchOnReconnect: true,
      retry: shouldRetryRequest,
      retryDelay: getRetryDelay,
      staleTime: 30 * 1000, // 30 seconds
      gcTime: 5 * 60 * 1000, // 5 minutes
      // Network error handling
      networkMode: 'online',
    },
    mutations: {
      retry: shouldRetryRequest,
      retryDelay: getRetryDelay,
      networkMode: 'online',
    },
  },
});
