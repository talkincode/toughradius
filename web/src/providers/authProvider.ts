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
      const result = await response.json();
      
      if (response.status < 200 || response.status >= 300) {
        // 返回后端的错误消息
        const errorMessage = result?.message || result?.error || response.statusText || '登录失败';
        throw new Error(errorMessage);
      }
      
      const auth = result.data || result; // 兼容包装格式
      
      if (!auth.token) {
        throw new Error('登录响应中缺少 token');
      }
      
      // 同步存储所有认证信息
      localStorage.setItem('token', auth.token);
      localStorage.setItem('username', username);
      localStorage.setItem('permissions', JSON.stringify(auth.permissions || []));
      
      // 存储完整的用户信息
      if (auth.user) {
        localStorage.setItem('user', JSON.stringify(auth.user));
      }
      
      // 确保数据已经写入 localStorage 后再继续
      await new Promise(resolve => setTimeout(resolve, 0));
      
      console.log('登录成功，token 已保存:', localStorage.getItem('token') ? '✓' : '✗');
      
      return Promise.resolve();
    } catch (error) {
      console.error('登录错误:', error);
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
    
    // 401 表示未认证，需要重新登录
    if (status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('username');
      localStorage.removeItem('permissions');
      localStorage.removeItem('user');
      return Promise.reject({ message: '认证已过期，请重新登录' });
    }
    
    // 403 表示权限不足，但不需要登出
    // 只是显示错误消息，保持登录状态
    if (status === 403) {
      return Promise.resolve(); // 不触发登出，只显示错误
    }
    
    return Promise.resolve();
  },

  // 检查认证状态 - 快速同步检查以避免闪烁
  checkAuth: () => {
    const token = localStorage.getItem('token');
    
    // 立即同步返回，避免异步延迟导致的闪烁
    if (!token) {
      return Promise.reject({ message: 'No token found', logoutUser: true });
    }
    
    // 简单验证 token 格式（避免明显无效的 token）
    if (token.length < 10) {
      // 清除无效的认证信息
      localStorage.removeItem('token');
      localStorage.removeItem('username');
      localStorage.removeItem('permissions');
      localStorage.removeItem('user');
      return Promise.reject({ message: 'Invalid token format', logoutUser: true });
    }
    
    return Promise.resolve();
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
