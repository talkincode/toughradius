#coding:utf-8
radius_attrs = {

  'RouterOS':[
     {
        'attr_name':'Mikrotik-Recv-Limit',
        'attr_desc':u'从客户端收到报文总字节数限制,可以用于基于流量记费.'
     },
     {
        'attr_name':'Mikrotik-Xmit-Limit',
        'attr_desc':u'发给客户端收到报文总字节数限制,可以用于基于流量记费'
     },
    {
        'attr_name':'Mikrotik-Rate-Limit',
        'attr_desc':u'客户的速率限制。字符串表示，格式为 rx-rate[/tx-rate] 如 256k/256k 或者 1M/2M'
     },
     {
        'attr_name':'Mikrotik-Advertise-URL',
        'attr_desc':u'通知页面地址（URL）。一般用于HOTSPOT'
     }
  ],
  'huawei 1.1':[
     {
        'attr_name':'Huawei-Input-Average-Rate',
        'attr_desc':u'(integer)上行平均速率 单位kbps'
     },
     {
        'attr_name':'Huawei-Input-Peak-Rate',
        'attr_desc':u'(integer)上行最大速率 单位kbps'
     },
    {
        'attr_name':'Huawei-Output-Average-Rate',
        'attr_desc':u'(integer)下行平均速率 单位kbps'
     },
     {
        'attr_name':'Huawei-Output-Peak-Rate',
        'attr_desc':u'(integer)下行最大速率 单位kbps'
     }
  ],
  'Cisco':[
     {
        'attr_name':'Cisco-AVPair',
        'attr_desc':u'思科专用属性'
     }
  ]


}