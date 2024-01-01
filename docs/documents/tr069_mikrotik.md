## Mikrotik TR069 Client Setup for ToughRADIUS

这段脚本可以用来在Mikrotik设备上自动配置TR069客户端，以便于ToughRADIUS进行设备管理。
这段脚本由 ToughRADIUS 生成，并自动替换变量， 在系统设置-TR069功能菜单里可以找到

```bash
# mikrotik tr069 client setup script
# Install certificate
:global acsCaCertTxt "{{.CacrtContent}}";

/file print file=tr069_ca_cert.txt;
delay 2;
/file set tr069_ca_cert.txt contents=$acsCaCertTxt;
/certificate import file-name=tr069_ca_cert.txt passphrase="";
/file remove tr069_ca_cert.txt;

# Wait while ehter ifaces show up
:local count 0;
:while ([/interface ethernet find] = "") do={
    :if ($count = 30) do={
        /quit;
    }
    :delay 1s; :set count ($count +1);
};

#Get serial-number
:local sn;
:if ([/system resource get board-name] = "CHR") do={
  :set sn [/system license get system-id];
} else={
  :set sn [/system routerboard get serial-number];
}

:local existingClient [/ip dhcp-client find interface=ether1];
:if ( $existingClient = "" ) do={
  /ip dhcp-client add interface=ether1 disabled=no comment="defconf";
} else={
  :put "DHCP Client already exists on interface ether1.";
}

/ip dns set servers=8.8.8.8

/tr069-client set acs-url="{{.TR069AccessAddress}}" enabled=yes \
username="$sn" password="{{.TR069AccessPassword}}" periodic-inform-interval=60s

:global ToughradiusApiServer "{{.TR069AccessAddress}}";
:global ToughradiusApiToken "{{.ToughradiusApiToken}}";

:local setupdate [/system clock get date];
:local setuptime [/system clock get time];

:local note ("# Device Info \r\
    \n1. Serial Number: $sn \r\
    \n2. TR069 Setup Time: $setupdate $setuptime \r\
    \n");

/system note set note=$note;
```