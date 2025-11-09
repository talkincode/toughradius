import { Admin, Resource } from 'react-admin';
import { dataProvider } from './providers/dataProvider';
import { authProvider } from './providers/authProvider';
import Dashboard from './pages/Dashboard';
import { CustomLayout } from './components';
import { theme } from './theme';

// 导入资源组件
import {
  RadiusUserList,
  RadiusUserEdit,
  RadiusUserCreate,
  RadiusUserShow,
} from './resources/radiusUsers';
import { OnlineSessionList, OnlineSessionShow } from './resources/onlineSessions';
import { AccountingList, AccountingShow } from './resources/accounting';
import {
  RadiusProfileList,
  RadiusProfileEdit,
  RadiusProfileCreate,
  RadiusProfileShow,
} from './resources/radiusProfiles';

const App = () => (
  <Admin
    dataProvider={dataProvider}
    authProvider={authProvider}
    dashboard={Dashboard}
    title="ToughRADIUS v9"
    theme={theme}
    layout={CustomLayout}
  >
    {/* RADIUS 用户管理 */}
    <Resource
      name="radius/users"
      list={RadiusUserList}
      edit={RadiusUserEdit}
      create={RadiusUserCreate}
      show={RadiusUserShow}
      options={{ label: 'RADIUS用户' }}
    />

    {/* 在线会话 */}
    <Resource
      name="radius/online"
      list={OnlineSessionList}
      show={OnlineSessionShow}
      options={{ label: '在线会话' }}
    />

    {/* 计费记录 */}
    <Resource
      name="radius/accounting"
      list={AccountingList}
      show={AccountingShow}
      options={{ label: '计费记录' }}
    />

    {/* RADIUS 配置 */}
    <Resource
      name="radius/profiles"
      list={RadiusProfileList}
      edit={RadiusProfileEdit}
      create={RadiusProfileCreate}
      show={RadiusProfileShow}
      options={{ label: 'RADIUS配置' }}
    />
  </Admin>
);

export default App;
