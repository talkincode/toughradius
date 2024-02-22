# Huawei device configuration and interconnection ToughRADIUS

Huawei's BRAS (Broadband Remote Access Server) such as the ME60 series, the detailed process and command line operations for connecting to the ToughRADIUS (Remote Authentication Dial-In User Service) server mainly involve the following steps:

## 1. Basic configuration

**Configure terminal parameters**: such as setting the terminal's baud rate, data bits, etc.

  ```
  system-view
  user-interface console 0
  idle-timeout minutes [seconds]
  screen-length 0 temporary
  ```
  
## 2. Create RADIUS server template

**Configure RADIUS server address and key**: Specify the IP address and shared key of the RADIUS server.

  ```
  radius-server template [template name]
  radius-server shared-key cipher [shared key]
  radius-server authentication [server IP address] [port number] weight [weight]
  radius-server accounting [server IP address] [port number] weight [weight]
  radius-server retransmit [number of retransmits]
  radius-server timeout [timeout]
  ```

## 3. Configure AAA (Authentication, Authorization, and Accounting)

**Configure AAA view**: Enable the AAA function and specify the authentication method as RADIUS.

  ```
  aaa
  authentication-scheme [scheme name]
  authentication-mode radius
  accounting-scheme [scheme name]
  accounting-mode radius
  domain default
  authentication-scheme [scheme name]
  accounting-scheme [scheme name]
  ```

## 4. Configure user interface

**Configure virtual template**: used for PPPoE or IPoE access.

  ```
  interface Virtual-Template [template number]
  ppp authentication-mode [authentication scheme name]
  ip address pool [address pool name]
  ```

## 5. Configure address pool

**Configure Address Allocation**: Configure an IP address pool for dial-up users.

  ```
  ip pool [address pool name] bas local
  gateway [gateway IP]
  network [network address] mask [subnet mask]
  ```

## 6. Configure VLAN and interface

**Configure VLAN interface**: Configure the VLAN used for Internet access.

  ```
  interface GigabitEthernet[interface number]
  port link-type access
  port default vlan [VLAN ID]
  ```

## 7. Debugging and Testing

**Test authentication and accounting functions**: Try dialing to check whether RADIUS authentication and accounting are normal.

  ```
  display radius-server statistics [template name]
  display aaa online-fail-record
  ```

## Precautions

- Ensure network connectivity: Ensure that BRAS can communicate normally with the RADIUS server.
- Key consistency: Ensure that the shared keys configured on the RADIUS server and BRAS are completely consistent.
- Version compatibility: Check whether the software versions of Huawei BRAS and RADIUS servers are compatible.

This process is a basic guide, and the specific configuration may vary based on the actual needs of the network and the specific model of the device. During the configuration process, please refer to Huawei's official documentation or consult technical support to ensure correct configuration.

After you complete the BRAS configuration, you need to create a corresponding VPE device in ToughRADIUS.
Then create a corresponding PPPoE user in ToughRADIUS, and finally create a PPPoE connection on the client device and use the PPPoE username and password to perform a dial-up test.