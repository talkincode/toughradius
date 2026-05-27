import { fetchUtils } from 'react-admin';

export const API_BASE = '/api/v1';

const normalizePath = (path: string): string => {
  if (path.startsWith('http://') || path.startsWith('https://')) {
    return path;
  }

  if (path.startsWith(API_BASE)) {
    return path;
  }

  if (path.startsWith('/')) {
    return `${API_BASE}${path}`;
  }

  return `${API_BASE}/${path}`;
};

const buildHeaders = (rawHeaders?: HeadersInit) => {
  const headers = new Headers(rawHeaders ?? { Accept: 'application/json' });
  if (!headers.has('Accept')) {
    headers.set('Accept', 'application/json');
  }
  const token = localStorage.getItem('token');
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }
  return headers;
};

const withAuth = <T extends RequestInit>(options: T = {} as T): T => {
  const headers = buildHeaders(options.headers);
  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  return {
    ...options,
    headers,
  } as T;
};
const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null;

export const extractData = <T = unknown>(payload: unknown): T => {
  if (isRecord(payload) && 'data' in payload) {
    return (payload as { data: T }).data;
  }
  return payload as T;
};

export const extractTotal = (payload: unknown): number => {
  if (isRecord(payload) && typeof payload.meta === 'object' && payload.meta !== null && 'total' in (payload.meta as Record<string, unknown>)) {
    return Number((payload.meta as Record<string, unknown>).total) || 0;
  }
  if (isRecord(payload) && 'total' in payload) {
    return Number(payload.total) || 0;
  }
  if (isRecord(payload) && Array.isArray(payload.data)) {
    return payload.data.length;
  }
  if (Array.isArray(payload)) {
    return payload.length;
  }
  return 0;
};

export const httpClient = (url: string, options: fetchUtils.Options = {}) =>
  fetchUtils.fetchJson(normalizePath(url), withAuth(options));

export class ApiError extends Error {
  status: number;
  body: unknown;

  constructor(status: number, body: unknown, message?: string) {
    const derivedMessage = (() => {
      if (message) {
        return message;
      }
      if (isRecord(body) && typeof body.message === 'string') {
        return body.message;
      }
      return `Request failed with status ${status}`;
    })();

    super(derivedMessage);
    this.status = status;
    this.body = body;
  }
}

export const apiRequest = async <T = unknown>(path: string, init: RequestInit = {}) => {
  const response = await fetch(normalizePath(path), withAuth(init));
  const contentType = response.headers.get('content-type') ?? '';

  let payload: unknown = null;
  if (response.status !== 204) {
    if (contentType.includes('application/json')) {
      payload = await response.json().catch(() => null);
    } else {
      payload = await response.text();
    }
  }

  if (!response.ok) {
    throw new ApiError(response.status, payload);
  }

  if (payload === null || payload === undefined || payload === '') {
    return payload as T;
  }

  return extractData<T>(payload);
};
