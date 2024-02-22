# FreeRadius rest module Configuration

[FreeRadius configuration case](https://github.com/talkincode/toughradius/tree/main/assets/freeradius)

This is the core configuration of the ToughRadius REST API, 
you need to pay attention to the parameters sent here, 
these parameters are all parameters supported by ToughRadius

If you are interconnecting with a private network, you do not need to configure TLS, 
but if you are connecting to a public network, it is recommended to use TLS encryption

> Be careful not to omit this configuration section when displaying it, every part is important!

```ini

rest {
	tls {
#		ca_file	= ${certdir}/cacert.pem
#		ca_path	= ${certdir}
#		certificate_file        = /path/to/radius.crt
#		private_key_file	= /path/to/radius.key
#		private_key_password	= "supersecret"
#		random_file		= /dev/urandom
		check_cert = no
		check_cert_cn = no
	}
	connect_uri = $ENV{FREERADIUS_API_URL}
	connect_timeout = 6.0
    authorize {
                #uri = "${..connect_uri}/user/%{User-Name}/mac/%{Called-Station-ID}?action=authorize"
                uri = "${..connect_uri}/freeradius/authorize"
                method = 'post'
                body = 'post'
                data = "username=%{urlquote:%{User-Name}}&nasip=%{urlquote:%{NAS-IP-Address}}&nasid=%{urlquote:%{NAS-Identifier}}"
                #tls = ${..tls}
    }

    authenticate {
                #uri = "${..connect_uri}/user/%{User-Name}/mac/%{Called-Station-ID}?action=authenticate"
                uri = "${..connect_uri}/freeradius/authenticate"
                method = 'post'
                body = 'post'
                data = "username=%{urlquote:%{User-Name}}&nasip=%{urlquote:%{NAS-IP-Address}}&nasid=%{urlquote:%{NAS-Identifier}}"
                #force_to = 'plain'
                #tls = ${..tls}
    }


    accounting {
                #uri = "${..connect_uri}/user/%{User-Name}/sessions/%{Acct-Unique-Session-ID}"
                uri = "${..connect_uri}/freeradius/accounting"
                method = 'post'
                body = 'post'
                data = "username=%{urlquote:%{User-Name}}&nasip=%{urlquote:%{NAS-IP-Address}}&nasid=%{urlquote:%{NAS-Identifier}}\
&acctSessionId=%{urlquote:%{Acct-Session-Id}}&macAddr=%{urlquote:%{Calling-Station-Id}}&acctSessionTime=%{urlquote:%{Acct-Session-Time}}\
&acctInputOctets=%{urlquote:%{Acct-Input-Octets}}&acctOutputOctets=%{urlquote:%{Acct-Output-Octets}}\
&acctInputGigawords=%{urlquote:%{Acct-Input-Gigawords}}&acctOutputGigawords=%{urlquote:%{Acct-Output-Gigawords}}\
&acctInputPackets=%{urlquote:%{Acct-Input-Packets}}&acctOutputPackets=%{urlquote:%{Acct-Output-Packets}}\
&nasPortId=%{urlquote:%{NAS-Port-Id}}&framedIPAddress=%{urlquote:%{Framed-IP-Address}}\
&sessionTimeout=%{urlquote:%{Session-Timeout}}&framedIPNetmask=%{urlquote:%{Framed-IP-Netmask}}\
&acctStatusType=%{urlquote:%{Acct-Status-Type}}"
    }

    post-auth {
                #uri = "${..connect_uri}/user/%{User-Name}/mac/%{Called-Station-ID}?action=post-auth"
                uri = "${..connect_uri}/freeradius/postauth"
                method = 'post'
                body = 'post'
                #tls = ${..tls}
    }

	pool {
		start = ${thread[pool].start_servers}
		min = ${thread[pool].min_spare_servers}
		max = ${thread[pool].max_servers}
		spare = ${thread[pool].max_spare_servers}
		uses = 0
		retry_delay = 30
		lifetime = 0
		idle_timeout = 60
	}
}
```