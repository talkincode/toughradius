import { AuthProvider, UserIdentity } from 'react-admin';
import { clearAuthStorage } from '../utils/storage';

type OperatorUser = {
  id?: string | number;
  username?: string;
  realname?: string;
  level?: string;
  status?: string;
  mobile?: string;
  email?: string;
  remark?: string;
  [key: string]: unknown;
};

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null;

const readCachedUser = (): OperatorUser | null => {
  const userStr = localStorage.getItem('user');
  if (!userStr) {
    return null;
  }
  try {
    const parsed = JSON.parse(userStr) as unknown;
    if (!isRecord(parsed)) {
      return null;
    }
    // Support both raw operator objects and nested { user: operator } payloads.
    if (isRecord(parsed.user)) {
      return parsed.user as OperatorUser;
    }
    return parsed as OperatorUser;
  } catch {
    return null;
  }
};

const parseJwtPayload = (token: string): Record<string, unknown> | null => {
  try {
    const parts = token.split('.');
    if (parts.length < 2) {
      return null;
    }
    const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
    const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '=');
    const json = decodeURIComponent(
      atob(padded)
        .split('')
        .map(char => `%${`00${char.charCodeAt(0).toString(16)}`.slice(-2)}`)
        .join(''),
    );
    const payload = JSON.parse(json) as unknown;
    return isRecord(payload) ? payload : null;
  } catch {
    return null;
  }
};

const buildIdentity = (user: OperatorUser | null, fallbackUsername?: string | null): UserIdentity => {
  const jwtToken = localStorage.getItem('token');
  const jwtPayload = jwtToken ? parseJwtPayload(jwtToken) : null;
  const jwtUsername = typeof jwtPayload?.username === 'string' ? jwtPayload.username : undefined;
  const jwtLevel = typeof jwtPayload?.role === 'string' ? jwtPayload.role : undefined;
  const jwtId = jwtPayload?.sub != null ? String(jwtPayload.sub) : undefined;

  const username =
    user?.username ||
    fallbackUsername ||
    localStorage.getItem('username') ||
    jwtUsername ||
    undefined;

  const realname = user?.realname || undefined;
  const level = user?.level || jwtLevel || undefined;
  const id = user?.id ?? jwtId ?? username ?? 'unknown';
  const fullName = realname || username || 'User';

  if (username) {
    localStorage.setItem('username', username);
  }

  return {
    ...(user ?? {}),
    id,
    fullName,
    username,
    level,
  };
};

const fetchIdentityFromApi = async (token: string): Promise<OperatorUser | null> => {
  try {
    const response = await fetch('/api/v1/auth/me', {
      method: 'GET',
      headers: {
        Accept: 'application/json',
        Authorization: `Bearer ${token}`,
      },
    });
    if (!response.ok) {
      return null;
    }
    const result = (await response.json()) as unknown;
    const data = isRecord(result) && 'data' in result ? result.data : result;
    if (!isRecord(data)) {
      return null;
    }
    const user = isRecord(data.user) ? (data.user as OperatorUser) : (data as OperatorUser);
    if (!user.username && !user.realname) {
      return null;
    }
    localStorage.setItem('user', JSON.stringify(user));
    if (user.username) {
      localStorage.setItem('username', user.username);
    }
    if (Array.isArray(data.permissions)) {
      localStorage.setItem('permissions', JSON.stringify(data.permissions));
    }
    return user;
  } catch {
    return null;
  }
};

export const authProvider: AuthProvider = {
  // 登录
  login: async ({ username, password }) => {
    const request = new Request('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
      headers: new Headers({ 'Content-Type': 'application/json' }),
    });

    try {
      const response = await fetch(request);
      const result = await response.json();

      if (response.status < 200 || response.status >= 300) {
        const errorMessage = result?.message || result?.error || response.statusText || '登录失败';
        throw new Error(errorMessage);
      }

      const auth = result.data || result;

      if (!auth.token) {
        throw new Error('登录响应中缺少 token');
      }

      localStorage.setItem('token', auth.token);
      localStorage.setItem('username', username);
      localStorage.setItem('permissions', JSON.stringify(auth.permissions || []));

      if (auth.user) {
        localStorage.setItem('user', JSON.stringify(auth.user));
        if (auth.user.username) {
          localStorage.setItem('username', auth.user.username);
        }
      }

      // Ensure localStorage writes are visible before navigation continues.
      await new Promise(resolve => setTimeout(resolve, 0));

      return Promise.resolve();
    } catch (error) {
      console.error('登录错误:', error);
      return Promise.reject(error);
    }
  },

  // 登出
  logout: async () => {
    clearAuthStorage();
    return Promise.resolve();
  },

  // 检查错误（如 401、403）
  checkError: async error => {
    const status = error.status;

    if (status === 401) {
      clearAuthStorage();
      return Promise.reject({ message: '认证已过期，请重新登录' });
    }

    // 403 表示权限不足，但不需要登出
    if (status === 403) {
      return Promise.resolve();
    }

    return Promise.resolve();
  },

  // 检查认证状态 - 快速同步检查以避免闪烁
  checkAuth: () => {
    const token = localStorage.getItem('token');

    if (!token) {
      return Promise.reject({ message: 'No token found', logoutUser: true });
    }

    if (token.length < 10) {
      clearAuthStorage();
      return Promise.reject({ message: 'Invalid token format', logoutUser: true });
    }

    return Promise.resolve();
  },

  // 获取权限
  getPermissions: async () => {
    const permissions = localStorage.getItem('permissions');
    return permissions ? Promise.resolve(JSON.parse(permissions)) : Promise.resolve([]);
  },

  // 获取用户身份信息（供 AppBar UserMenu 显示 fullName）
  getIdentity: async () => {
    const cachedUser = readCachedUser();
    const cachedUsername = localStorage.getItem('username');
    if (cachedUser?.username || cachedUser?.realname || cachedUsername) {
      return buildIdentity(cachedUser, cachedUsername);
    }

    const token = localStorage.getItem('token');
    if (token) {
      const remoteUser = await fetchIdentityFromApi(token);
      if (remoteUser) {
        return buildIdentity(remoteUser, remoteUser.username);
      }
      // Last resort: JWT claims still carry username/role for display.
      return buildIdentity(null, cachedUsername);
    }

    return buildIdentity(null, cachedUsername);
  },
};
