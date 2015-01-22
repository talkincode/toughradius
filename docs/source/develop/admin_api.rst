ToughRADIUS管理接口
========================

接口协议
-----------------------

ToughRADIUS的管理功能主要通过websocket实现，WebSocket协议本质上是一个基于 TCP 的协议。为了建立一个 WebSocket 连接，客户端浏览器首先要向服务器发起一个 HTTP 请求，这个请求和通常的 HTTP 请求不同，包含了一些附加头信息，其中附加头信息”Upgrade: WebSocket”表明这是一个申请协议升级的 HTTP 请求，服务器端解析这些附加的头信息然后产生应答信息返回给客户端，客户端和服务器端的 WebSocket 连接就建立起来了，双方就可以通过这个连接通道自由的传递信息，并且这个连接会持续存在直到客户端或者服务器端的某一方主动的关闭连接。

以上描述看起来并不容易理解，看看一个例子或许更有帮助:

.. code-block:: javascript

    var updater = {
        socket: null,
        start: function () {
            var url = "ws://192.168.8.1:1815";
            updater.socket = new WebSocket(url);
            updater.socket.onmessage = function (event) {
                updater.showMessage(JSON.parse(event.data));
            }
            updater.socket.onclose = function (event) {
                updater.start();
            }
            updater.socket.onopen = updater.onOpen;
        },

        showMessage: function (message) {
            alert(message.data)
            window.location.reload();
        },

        onOpen: function(event){},

        unlock_online: function(nas_addr,session_id){
            var message = {
                process : "admin_unlock_online",
                nas_addr : nas_addr,
                acct_session_id : session_id
            };
            data = JSON.stringify(message);
            updater.socket.send(data);
        },

        disconnect_online: function(nas_addr,session_id){
            var message = {
                process : "admin_coa_request",
                nas_addr : nas_addr,
                acct_session_id : session_id,
                message_type : 'disconnect'
            };
            data = JSON.stringify(message);
            updater.socket.send(data);
        }
    };

以上代码就是ToughRADIUS的web管理系统中用来解锁在线用户和强制下线的调用，如果你写过javascript应用，很容易就看懂了。

接口配置
-----------------------

注意adminport和admin_allows两个参数，adminport是提供服务的端口，admin_allows是表示允许访问的地址。

.. code-block:: javascript

    {
        "radiusd":
        {
            "authport": 1812,
            "acctport": 1813,
            "adminport": 1815,
            "admin_allows":"192.168.88.100,192.168.88.100",
            "dictfile": "./radiusd/dict/dictionary",
            "debug":1,
            "cache_timeout":600
        }
    }


接口定义
-----------------------

所有接口通讯采用json数据格式。

+ 请求消息::
    
    process是必须参数，表示调用哪一个接口

    {'process':'admin_update_cache',...}

+ 响应消息::

    code表示成功或失败，0是成功，1是失败，data是返回数据或错误消息

    {'code':0,'data':''}



更新缓存
~~~~~~~~~~~~~~~~~~~~~~~~

ToughRADIUS使用了缓存来提升系统性能，当关键数据修改更新后，必须更新缓存，以保持数据一致。

+ 请求消息::

    {
        'process' : 'admin_update_cache',
        'cache_class' : 'param',//已定义的类型有：
    }

    cache_class参数::

        param，在更新系统参数后调用，更新参数缓存。
        account，在更新上网账号后调用，更新上网账号缓存。
        bas，在更新bas信息后调用，更新bas缓存。
        group，在更新用户组信息后调用，更新用户组缓存。
        roster，在更新黑白名单信息后调用，更新黑白名单缓存。
        product，在更新资费信息后调用，更新资费缓存。

+ 响应消息::

    {'code':0,'data':'param cache update ok'}













