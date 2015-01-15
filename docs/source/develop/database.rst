ToughRADIUS数据字典
====================================

区域信息表  slc_node
-------------------------------------

.. start_table slc_node;node_id

=====================  ==============      ==========   ==============================================
属性                    类型（长度）         可否为空      描述
=====================  ==============      ==========   ==============================================
id                     int(11)             no           区域编号
node_name              varchar(32)         no           区域名称
node_desc              varchar(64)         no           区域描述
=====================  ==============      ==========   ==============================================

.. end_table

用户信息表  slc_member
-------------------------------------

.. start_table slc_member;member_id

=====================  ==============      ==========   ==============================================
属性                    类型（长度）         可否为空      描述
=====================  ==============      ==========   ==============================================
member_id               int(11)             no          用户id
node_id                 int(11)             no          用户区域id
member_name             varchar(64)         no          用户自助服务登陆名
password                varchar(128)        no          用户自助服务密码
realname                varchar(64)         no          用户姓名
idcard                  varchar(32)         no          用户证件
sex                     int(1)              no          用户性别
age                     int(3)              no          用户年龄
email                   varchar(128)        no          用户邮箱
mobile                  varchar(16)         no          用户手机
address                 varchar(128)        no          用户地址
create_time             varchar(19)         no          创建时间 格式：yyyy-mm-dd hh:mm:ss
update_time             varchar(19)         no          修改时间 格式：yyyy-mm-dd hh:mm:ss
=====================  ==============      ==========   ==============================================

.. end_table


用户订购表  slc_member_order
-------------------------------------

::

    # 交易支付状态：pay_status 0-未支付，1-已支付，2-已取消
    # 订单受理渠道 console: 管理系统  customer: 客户自助服务

.. start_table slc_member_order;order_id

=====================  ==============      ==========   ==============================================
属性                    类型（长度）         可否为空      描述
=====================  ==============      ==========   ==============================================
order_id                varchar(32)         no          订购时间
member_id               int(11)             no          订单所属用户id
product_id              int(11)             no          订单资费套餐id
account_number          varchar(32)         no          订单上网账号
order_fee               int(11)             no          订单应付费用
actual_fee              int(11)             no          订单实际缴纳费用
pay_status              int(1)              no          订单支付状态
accept_id               int(11)             no          订单受理日志id
order_source            varchar(16)         no          订单受理渠道
order_desc              varchar(255)        yes         订单描述
create_time             varchar(19)         no          创建时间 格式：yyyy-mm-dd hh:mm:ss
=====================  ==============      ==========   ==============================================

.. end_table



上网账号表  slc_rad_account
-------------------------------------

::

    # 用户状态 0:"预授权",1:"正常", 2:"停机",3:"销户", 4:"到期", 5:"未激活"


.. start_table slc_rad_account;account_number

=====================  ==============      ==========   ==============================================
属性                    类型（长度）         可否为空      描述
=====================  ==============      ==========   ==============================================
account_number          varchar(32)         no          上网账号
member_id               int(11)             no          账号所属用户id
product_id              int(11)             no          账号绑定资费
group_id                int(11)             yes         账号绑定组
password                varchar(64)         no          账号上网密码
status                  int(1)              no          账号状态
install_address         varchar(128)        yes         账号安装地址
balance                 int(11)             no          账号余额，分
time_length             int(11)             yes         账号上网时长
expire_date             varchar(10)         no          账号过期时间
user_concur_number      int(3)              no          账号认证并发数
bind_mac                int(1)              no          账号是否绑定mac
bind_vlan               int(1)              no          账号是否绑定vlan 
mac_addr                varchar(19)         yes         账号绑定MAC
vlan_id                 int(11)             yes         账号内层vlan
vlan_id2                int(11)             yes         账号外层vlan
ip_address              varchar(19)         yes         账号绑定ip地址
last_pause              varchar(19)         no          最后停机时间 格式：yyyy-mm-dd hh:mm:ss
create_time             varchar(19)         no          创建时间 格式：yyyy-mm-dd hh:mm:ss
update_time             varchar(19)         no          修改时间 格式：yyyy-mm-dd hh:mm:ss
=====================  ==============      ==========   ==============================================

.. end_table