
## 华为设备配置对接 ToughRADIUS

华为的BRAS (Broadband Remote Access Server) 如ME60系列，对接 ToughRADIUS (Remote Authentication Dial-In User Service) 服务器的详细流程和命令行操作主要涉及以下步骤：

### 1. 基本配置
- **配置终端参数**：如设置终端的波特率、数据位等。
  ```
  system-view
  user-interface console 0
  idle-timeout minutes [seconds]
  screen-length 0 temporary
  ```
  
### 2. 创建RADIUS服务器模板
- **配置RADIUS服务器地址和密钥**：指定RADIUS服务器的IP地址和共享密钥。
  ```
  radius-server template [模板名]
  radius-server shared-key cipher [共享密钥]
  radius-server authentication [服务器IP地址] [端口号] weight [权重]
  radius-server accounting [服务器IP地址] [端口号] weight [权重]
  radius-server retransmit [重传次数]
  radius-server timeout [超时时间]
  ```

### 3. 配置AAA（Authentication, Authorization, and Accounting）
- **配置AAA视图**：启用AAA功能，指定认证方式为RADIUS。
  ```
  aaa
  authentication-scheme [方案名]
  authentication-mode radius
  accounting-scheme [方案名]
  accounting-mode radius
  domain default
  authentication-scheme [方案名]
  accounting-scheme [方案名]
  ```

### 4. 配置用户接口
- **配置虚拟模板**：用于PPPoE或IPoE接入。
  ```
  interface Virtual-Template [模板号]
  ppp authentication-mode [认证方案名]
  ip address pool [地址池名]
  ```

### 5. 配置地址池
- **配置地址分配**：为拨号用户配置IP地址池。
  ```
  ip pool [地址池名] bas local
  gateway [网关IP]
  network [网络地址] mask [子网掩码]
  ```

### 6. 配置VLAN和接口
- **配置VLAN接口**：配置用于上网的VLAN。
  ```
  interface GigabitEthernet[接口号]
  port link-type access
  port default vlan [VLAN ID]
  ```

### 7. 调试和测试
- **测试认证和记账功能**：通过拨号尝试，检查RADIUS认证和记账是否正常。
  ```
  display radius-server statistics [模板名]
  display aaa online-fail-record
  ```

### 注意事项
- 确保网络连通性：确保BRAS能够与RADIUS服务器正常通信。
- 密钥一致性：确保在RADIUS服务器和BRAS上配置的共享密钥完全一致。
- 版本兼容性：检查华为BRAS与RADIUS服务器的软件版本是否兼容。

这个流程是一个基本的指南，具体的配置可能会根据网络的实际需求和设备的具体型号有所不同。在配置过程中，请参考华为的官方文档或咨询技术支持以确保正确配置。

当您在 BRAS 配置完成后，您需要在 ToughRADIUS 中创建一个对应的 VPE 设备，
然后在 ToughRADIUS 中创建一个对应的 PPPoE 用户，最后在客户端设备上创建一个 PPPoE 连接，使用 PPPoE 用户名和密码进行拨号测试。