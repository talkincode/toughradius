# freeRADIUS integration


![freeradius-toughradius](https://github.com/talkincode/toughradius/assets/377938/f735d45d-3325-49e5-8b73-21c6205248e3)


TOUGHRADIUS integrates the FreeRADIUS API interface, extending its already comprehensive authentication capabilities and providing even more robust solutions to its clients. 
Our integration of the FreeRADIUS API allows for seamless integration with existing network infrastructures and enables us to offer a wider range of authentication options to meet the unique needs of our clients. 
Whether you require support for 802.1X, Wi-Fi, VPN, or other network access protocols, TOUGHRADIUS has you covered. With our advanced authentication capabilities and integration with FreeRADIUS, our clients can enjoy a secure, reliable, and efficient network management experience.


- FreeRadius enables the REST Module to interface with ToughRadius and uses HTTP parameters to pass user information
- Freeradius acts as the core of the RADIUS protocol processing
- Toughradius acts as a function for user management, billing, device management, device configuration management, and more
- Toughradius starts an HTTP server that listens for Freeradius requests and handles user authentication, billing, device management, and more

> Please refer to [Freeradius Official Documentation](https://networkradius.com/doc/3.0.10/raddb/mods-available/rest.html)

> [FreeRadius configuration case](https://github.com/talkincode/toughradius/tree/main/assets/freeradius)

To implement the integration of FreeRadius, you must have sufficient knowledge of FreeRadius

## FreeRadius Default AuthType Configuration

[FreeRadius configuration case](https://github.com/talkincode/toughradius/tree/main/assets/freeradius)

Modify the freeradius configuration file, probably /etc/freeradius/3.0/users

```ini
#
# Default for PPP: dynamic IP address, PPP mode, VJ-compression.
# NOTE: we do not use Hint = "PPP", since PPP might also be auto-detected
#	by the terminal server in which case there may not be a "P" suffix.
#	The terminal server sends "Framed-Protocol = PPP" for auto PPP.
#
#DEFAULT        Auth-Type := python
DEFAULT Auth-Type := rest

DEFAULT	Framed-Protocol == PPP
	Framed-Protocol = PPP,
	Framed-Compression = Van-Jacobson-TCP-IP

#
```


## FreeRadius client Configuration

[FreeRadius configuration case](https://github.com/talkincode/toughradius/tree/main/assets/freeradius)

注意: 请根据实际情况修改配置文件中的参数， 在设备中配置的 radius 密钥实际上是这里的 secret 参数

```conf
client localhost {
	ipaddr = 127.0.0.1
	proto = *
	secret = testing123
	require_message_authenticator = no
	nas_type	 = other
	limit {
		max_connections = 16
		lifetime = 0
		idle_timeout = 30
	}
}

client any {
        ipaddr          = 0.0.0.0/0
        secret          = mysecret
}
```


## 配置 freeradius sites-enabled/default

[FreeRadius configuration case](https://github.com/talkincode/toughradius/tree/main/assets/freeradius)

Enable the REST module in FreeRadius's configuration file, 
note that this configuration is just a simplified case, 
telling you where you should enable REST, 

For more freeradius configurations, please configure them according to the actual situation

```ini
server default {
    listen {
        type = auth
        ipaddr = *
        port = 1812
        limit {
              max_connections = 6
              lifetime = 0
              idle_timeout = 30
        }
    }
    
    listen {
        ipaddr = *
        port = 1813
        type = acct
        limit {
    
        }
    }
    
    listen {
        type = auth
        ipv6addr = ::	# any.  ::1 == localhost
        port = 1822
        limit {
              max_connections = 0
              lifetime = 0
              idle_timeout = 30
        }
    }
    
    listen {
        ipv6addr = ::
        port = 1823
        type = acct
        
        limit {
        }
    }
    
    authorize {
        auth_log
        rest
        chap
        eap {
            ok = return
        }
        pap
    }
    
    authenticate {
        Auth-Type rest {
            rest
        }
    
        Auth-Type PAP {
            pap
        }
        Auth-Type CHAP {
            chap
        }
        Auth-Type MS-CHAP {
            mschap
        }
        mschap
        digest
        eap
    }
    
    preacct {
        preprocess
        acct_unique
        suffix
        files
    }
    
    accounting {
        detail
        unix
        -sql
        exec
        attr_filter.accounting_response
    }
    
    session {
    
    }
    
    post-auth {
        update {
            &reply: += &session-state:
        }
        reply_log
        -sql
        exec
        remove_reply_message_if_eap
        Post-Auth-Type REJECT {
            # log failed authentications in SQL, too.
            -sql
            attr_filter.access_reject
            eap
            remove_reply_message_if_eap
        }
    
    
    }
    pre-proxy {
    }
        
    post-proxy {
    }

}

```
