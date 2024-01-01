## Mikrotik TR069 Client Setup for ToughRADIUS

Here is the guide to configure ROUTEROS to integrate with TOUGHRADIUS:

Firstly, we need to execute the following commands on ROUTEROS:

```
# Create address pool
/ip pool
add name=dhcp ranges=192.168.1.100-192.168.1.200

# Create profile
/ppp profile
add name=radius local-address=192.168.1.1 remote-address=dhcp

# Configure RADIUS server
/radius
add service=ppp,hotspot address=Radius_Server_IP secret=Radius_Secret authentication-port=1812 accounting-port=1813

# Configure accounting interval
/radius incoming
set accept=yes port=3799

# Configure PPP
/ppp aaa
set accounting=yes interim-update=2m use-radius=yes

```

Next, we need to create a corresponding VPE device on toughradius:

```json
{
  "ID": 1,
  "NodeId": 1,
  "LdapId": 0,
  "Name": "RouterOS",
  "Identifier": "RouterOS",
  "Hostname": "RouterOS Host",
  "Ipaddr": "RouterOS IP",
  "Secret": "Radius Secret",
  "CoaPort": 3799,
  "Model": "RouterOS",
  "VendorCode": "14988",
  "Status": "enabled",
  "Tags": "",
  "Remark": "",
  "CreatedAt": "2022-06-01T08:00:00.000Z",
  "UpdatedAt": "2022-06-01T08:00:00.000Z"
}
```

Lastly, we carry out testing:

> you also need to create the user account on the TOUGHRADIUS system

Execute the following command on ROUTEROS to create a new PPP user:

```
/ppp secret
add name=testuser password=testpass profile=radius service=pppoe

```

Configure a PPPoE connection on your client device, with the username testuser and the password testpass.

Connect to the PPPoE, check the ROUTEROS and toughradius logs to confirm that the user's authentication and accounting information are being transmitted correctly.

If the connection is successful, you should see the online status and usage of the user on the toughradius admin interface.

You may also try to change the bandwidth limit for the user or disconnect the user on toughradius and check if the CoA (Change of Authorization) feature takes effect on ROUTEROS.

