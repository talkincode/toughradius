import polyglotI18nProvider from 'ra-i18n-polyglot';
import zhCN from './zh-CN';
import enUS from './en-US';

const translations = {
  'zh-CN': zhCN,
  'en-US': enUS,
};

// 从 localStorage 获取保存的语言设置，如果没有则使用默认语言
const getDefaultLocale = () => {
  const savedLocale = localStorage.getItem('locale');
  return savedLocale && translations[savedLocale as keyof typeof translations] 
    ? savedLocale 
    : 'zh-CN';
};

const baseI18nProvider = polyglotI18nProvider(
  (locale) => translations[locale as keyof typeof translations] || translations['zh-CN'],
  getDefaultLocale(), // 使用保存的语言或默认语言
  [
    { locale: 'zh-CN', name: '简体中文' },
    { locale: 'en-US', name: 'English' },
  ],
  { allowMissing: true }
);

// 包装 i18nProvider 以在切换语言时保存到 localStorage
export const i18nProvider = {
  ...baseI18nProvider,
  changeLocale: (locale: string) => {
    // 保存语言设置到 localStorage
    localStorage.setItem('locale', locale);
    return baseI18nProvider.changeLocale(locale);
  },
};
