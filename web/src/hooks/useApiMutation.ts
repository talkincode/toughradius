import { useMutation, UseMutationOptions } from '@tanstack/react-query';
import { ApiError, apiRequest } from '../utils/apiClient';

type MutationVariables = {
  path: string;
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  body?: unknown;
  headers?: HeadersInit;
};

type MutationResult<TData> = UseMutationOptions<TData, ApiError, MutationVariables>;

export const useApiMutation = <TData = unknown>(options?: MutationResult<TData>) =>
  useMutation<TData, ApiError, MutationVariables>({
    mutationFn: ({ path, method = 'POST', body, headers }) => {
      const init: RequestInit = {
        method,
        headers,
      };
      if (body !== undefined) {
        init.body = typeof body === 'string' ? body : JSON.stringify(body);
      }
      return apiRequest<TData>(path, init);
    },
    ...options,
  });
