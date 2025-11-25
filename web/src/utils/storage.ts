/**
 * 存储工具函数
 * 用于管理 localStorage 中的数据，特别是区分需要保留和需要清除的数据
 */

// 需要在退出登录时保留的 localStorage 键
const PERSISTENT_KEYS = ['locale', 'theme'];

/**
 * 清除认证相关的 localStorage 数据
 * 保留用户偏好设置（如语言、主题等）
 */
export const clearAuthStorage = () => {
  const keysToRemove = ['token', 'username', 'permissions', 'user'];
  keysToRemove.forEach(key => localStorage.removeItem(key));
};

/**
 * 清除所有 localStorage 数据，除了需要保留的用户偏好
 */
export const clearAllStorageExceptPreferences = () => {
  // 保存需要保留的数据
  const persistentData: Record<string, string | null> = {};
  PERSISTENT_KEYS.forEach(key => {
    persistentData[key] = localStorage.getItem(key);
  });
  
  // 清空所有数据
  localStorage.clear();
  
  // 恢复需要保留的数据
  PERSISTENT_KEYS.forEach(key => {
    if (persistentData[key] !== null) {
      localStorage.setItem(key, persistentData[key]!);
    }
  });
};
