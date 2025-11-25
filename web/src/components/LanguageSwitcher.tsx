import { MenuItem, ListItemIcon, ListItemText } from '@mui/material';
import { useSetLocale, useLocaleState } from 'react-admin';
import LanguageIcon from '@mui/icons-material/Language';

export const LanguageSwitcher = () => {
  const setLocale = useSetLocale();
  const [locale] = useLocaleState();

  const handleLanguageChange = () => {
    const newLocale = locale === 'zh-CN' ? 'en-US' : 'zh-CN';
    setLocale(newLocale);
  };

  return (
    <MenuItem onClick={handleLanguageChange}>
      <ListItemIcon>
        <LanguageIcon fontSize="small" />
      </ListItemIcon>
      <ListItemText>{locale === 'zh-CN' ? 'English' : '简体中文'}</ListItemText>
    </MenuItem>
  );
};
