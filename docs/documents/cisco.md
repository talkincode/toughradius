# The Cisco BRAS device is connected to the ToughRADIUS server

Cisco User Manual for guidance on how to dock a Cisco Broadband Remote Access Server (BRAS) device to a ToughRADIUS server,
There are a series of steps involved. Here's a high-level process that includes the necessary command-line operations:

## 1. Configure RADIUS server information

First, you need to configure the information of the RADIUS server on the Cisco BRAS device. This usually includes the server's IP address and shared key.

```
radius-server host [RADIUS server IP address] key [shared key]
```

## 2. Configure authentication and accounting

Next, configure the device to use RADIUS for Authentication and Accounting.

```
aaa new-model
aaa authentication ppp default group radius
aaa accounting network default start-stop group radius
```

These commands enable AAA (Authentication, Authorization, and Accounting) and set the default PPP authentication and network accounting to use RADIUS.

## 3. Configure the user interface

Configure the user interface based on your network architecture. This may include setting up virtual templates, interface pools, and so on.

```
interface Virtual-Template1
 ip unnumbered [an interface]
 peer default ip address pool
 ppp authentication chap
```

## 4. Create an address pool

If your users will get IP addresses from BRAS devices, you need to create an address pool.

```
ip local pool [address pool name] [start IP address] [end IP address]
```

### 5. Test the configuration

Once the configuration is complete, test to ensure that the BRAS device can successfully communicate with the RADIUS server. This can be done by trying to connect from the client device.

## 6. Monitoring and troubleshooting

Monitor the logs of the BRAS and RADIUS to make sure everything is working properly. If you encounter problems, use the following command to troubleshoot:

```
debug radius authentication
debug radius accounting
```

Please note that this process is a basic guide, and the exact configuration may vary depending on your network environment and needs. Before any configuration is made,
Make sure you have read Cisco's official documentation in detail and understand your network architecture. At the same time, it is recommended to experiment with the configuration in a test environment other than the production environment.

After you configure BRAS, you need to create a corresponding VPE device in ToughRADIUS.
Then create a corresponding PPPoE user in ToughRADIUS, and finally create a PPPoE connection on the client device for dial-up testing with the PPPoE username and password.