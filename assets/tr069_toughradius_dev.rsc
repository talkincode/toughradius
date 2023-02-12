#Get serial-number
:local sn;
:set sn [/system routerboard get serial-number];
:if ([:len $sn]=0) do={
    :set sn [/system license get system-id];
}

/tr069-client set acs-url="http://10.189.189.56:1819" enabled=yes \
username="$sn" password="examplesecurepassword" periodic-inform-interval=30s
