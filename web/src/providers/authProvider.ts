import { AuthProvider } from 'react-admin';

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
      if (response.status < 200 || response.status >= 300) {
        throw new Error(response.statusText);
      }
      const result = await response.json();
      const auth = result.data || result; // 兼容包装和非包装格式
      localStorage.setItem('token', auth.token);
      localStorage.setItem('username', username);
      localStorage.setItem('permissions', JSON.stringify(auth.permissions || []));
      // 存储完整的用户信息
      if (auth.user) {
        localStorage.setItem('user', JSON.stringify(auth.user));
      }
      return Promise.resolve();
    } catch (error) {
      return Promise.reject(error);
    }
  },

  // 登出
  logout: async () => {
    localStorage.removeItem('token');
    localStorage.removeItem('username');
    localStorage.removeItem('permissions');
    localStorage.removeItem('user');
    return Promise.resolve();
  },

  // 检查错误（如 401、403）
  checkError: async (error) => {
    const status = error.status;
    if (status === 401 || status === 403) {
      localStorage.removeItem('token');
      localStorage.removeItem('username');
      localStorage.removeItem('permissions');
      localStorage.removeItem('user');
      return Promise.reject();
    }
    return Promise.resolve();
  },

  // 检查认证状态
  checkAuth: async () => {
    return localStorage.getItem('token') ? Promise.resolve() : Promise.reject();
  },

  // 获取权限
  getPermissions: async () => {
    const permissions = localStorage.getItem('permissions');
    return permissions ? Promise.resolve(JSON.parse(permissions)) : Promise.resolve([]);
  },

  // 获取用户身份信息
  getIdentity: async () => {
    const userStr = localStorage.getItem('user');
    if (userStr) {
      const user = JSON.parse(userStr);
      return Promise.resolve({
        id: user.id,
        fullName: user.realname || user.username,
        username: user.username,
        level: user.level,
        ...user,
      });
    }
    const username = localStorage.getItem('username');
    return Promise.resolve({
      id: username || 'anonymous',
      fullName: username || 'Anonymous',
      username: username || 'anonymous',
    });
  },
};
