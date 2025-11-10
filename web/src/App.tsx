import { Admin, Resource, CustomRoutes } from 'react-admin';
import { Route } from 'react-router-dom';
import { dataProvider } from './providers/dataProvider';
import { authProvider } from './providers/authProvider';
import Dashboard from './pages/Dashboard';
import { AccountSettings } from './pages/AccountSettings';
import { SystemConfigPage } from './pages/SystemConfigPage';
import { LoginPage } from './pages/LoginPage';
import { CustomLayout } from './components';
import { theme, darkTheme } from './theme';

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
import {
  NASList,
  NASEdit,
  NASCreate,
  NASShow,
} from './resources/nas';
import {
  NodeList,
  NodeEdit,
  NodeCreate,
  NodeShow,
} from './resources/nodes';
import {
  OperatorList,
  OperatorEdit,
  OperatorCreate,
  OperatorShow,
} from './resources/operators';

const App = () => (
  <Admin
    dataProvider={dataProvider}
    authProvider={authProvider}
    dashboard={Dashboard}
    loginPage={LoginPage}
    title="ToughRADIUS v9"
    theme={theme}
    darkTheme={darkTheme}
    defaultTheme="light"
    layout={CustomLayout}
    requireAuth
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

    {/* NAS 设备管理 */}
    <Resource
      name="network/nas"
      list={NASList}
      edit={NASEdit}
      create={NASCreate}
      show={NASShow}
      options={{ label: 'NAS设备' }}
    />

    {/* 网络节点 */}
    <Resource
      name="network/nodes"
      list={NodeList}
      edit={NodeEdit}
      create={NodeCreate}
      show={NodeShow}
      options={{ label: '网络节点' }}
    />

    {/* 操作员管理 */}
    <Resource
      name="system/operators"
      list={OperatorList}
      edit={OperatorEdit}
      create={OperatorCreate}
      show={OperatorShow}
      options={{ label: '操作员管理' }}
    />

    {/* 自定义路由 */}
    <CustomRoutes>
      <Route path="/account/settings" element={<AccountSettings />} />
      <Route path="/system/config" element={<SystemConfigPage />} />
    </CustomRoutes>
  </Admin>
);

export default App;
