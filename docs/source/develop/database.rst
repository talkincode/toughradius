ToughRADIUS数据字典
====================================


.. _slc_node_label:

slc_node
------------------------------------ 

区域表

.. start_table slc_node;id 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
id                     INTEGER           False             区域编号              
node_name              VARCHAR(32)       False             区域名                 
node_desc              VARCHAR(64)       False             区域描述              
=====================  ================  ================  ====================================

.. end_table


.. _slc_operator_label:

slc_operator
------------------------------------ 

操作员表 操作员类型 0 系统管理员 1 普通操作员

.. start_table slc_operator;id 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
id                     INTEGER           False             操作员id               
operator_type          INTEGER           False             操作员类型           
operator_name          VARCHAR(32)       False             操作员名称           
operator_pass          VARCHAR(128)      False             操作员密码           
operator_status        INTEGER           False             操作员状态,0/1       
operator_desc          VARCHAR(255)      False             操作员描述           
=====================  ================  ================  ====================================

.. end_table


.. _slc_operator_rule_label:

slc_operator_rule
------------------------------------ 

操作员权限表

.. start_table slc_operator_rule;id 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
id                     INTEGER           False             权限id                  
operator_name          VARCHAR(32)       False             操作员名称           
rule_path              VARCHAR(128)      False             权限URL                 
rule_name              VARCHAR(128)      False             权限名称              
rule_category          VARCHAR(128)      False             权限分类              
=====================  ================  ================  ====================================

.. end_table


.. _slc_param_label:

slc_param
------------------------------------ 

系统参数表  <radiusd default table>

.. start_table slc_param;param_name 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
param_name             VARCHAR(64)       False             参数名                 
param_value            VARCHAR(255)      False             参数值                 
param_desc             VARCHAR(255)      True              参数描述              
=====================  ================  ================  ====================================

.. end_table


.. _slc_rad_bas_label:

slc_rad_bas
------------------------------------ 

BAS设备表 <radiusd default table>

.. start_table slc_rad_bas;id 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
id                     INTEGER           False             设备id                  
vendor_id              VARCHAR(32)       False             厂商标识              
ip_addr                VARCHAR(15)       False             IP地址                  
bas_name               VARCHAR(64)       False             bas名称                 
bas_secret             VARCHAR(64)       False             共享密钥              
coa_port               INTEGER           False             CoA端口                 
time_type              SMALLINT          False             时区类型              
=====================  ================  ================  ====================================

.. end_table


.. _slc_rad_roster_label:

slc_rad_roster
------------------------------------ 

黑白名单 0 白名单 1 黑名单 <radiusd default table>

.. start_table slc_rad_roster;id 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
id                     INTEGER           False             黑白名单id            
mac_addr               VARCHAR(17)       False             mac地址                 
begin_time             VARCHAR(19)       False             生效开始时间        
end_time               VARCHAR(19)       False             生效结束时间        
roster_type            SMALLINT          False             黑白名单类型        
=====================  ================  ================  ====================================

.. end_table


.. _slc_member_label:

slc_member
------------------------------------ 

用户信息表

.. start_table slc_member;member_id 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
member_id              INTEGER           False             用户id                  
node_id                INTEGER           False             区域id                  
member_name            VARCHAR(64)       False             用户登录名           
password               VARCHAR(128)      False             用户登录密码        
realname               VARCHAR(64)       False                                       
idcard                 VARCHAR(32)       True              用户证件号码        
sex                    SMALLINT          True              用户性别0/1           
age                    INTEGER           True              用户年龄              
email                  VARCHAR(255)      True              用户邮箱              
mobile                 VARCHAR(16)       True              用户手机              
address                VARCHAR(255)      True              用户地址              
create_time            VARCHAR(19)       False             创建时间              
update_time            VARCHAR(19)       False             更新时间              
=====================  ================  ================  ====================================

.. end_table


.. _slc_member_order_label:

slc_member_order
------------------------------------ 


    订购信息表(交易记录)
    pay_status交易支付状态：0-未支付，1-已支付，2-已取消
    

.. start_table slc_member_order;order_id 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
order_id               VARCHAR(32)       False             订单id                  
member_id              INTEGER           False             用户id                  
product_id             INTEGER           False             资费id                  
account_number         VARCHAR(32)       False             上网账号              
order_fee              INTEGER           False             订单费用              
actual_fee             INTEGER           False             实缴费用              
pay_status             INTEGER           False             支付状态              
accept_id              INTEGER           False             受理id                  
order_source           VARCHAR(64)       False             订单来源              
order_desc             VARCHAR(255)      True              订单描述              
create_time            VARCHAR(19)       False             交易时间              
=====================  ================  ================  ====================================

.. end_table


.. _slc_rad_account_label:

slc_rad_account
------------------------------------ 


    上网账号表，每个会员可以同时拥有多个上网账号
    account_number 为每个套餐对应的上网账号，每个上网账号全局唯一
    用户状态 0:"预定",1:"正常", 2:"停机" , 3:"销户", 4:"到期"
    <radiusd default table>
    

.. start_table slc_rad_account;account_number 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
account_number         VARCHAR(32)       False             上网账号              
member_id              INTEGER           False             用户id                  
product_id             INTEGER           False             资费id                  
group_id               INTEGER           True              用户组id               
password               VARCHAR(128)      False             上网密码              
status                 INTEGER           False             用户状态              
install_address        VARCHAR(128)      False             装机地址              
balance                INTEGER           False             用户余额-分          
time_length            INTEGER           False             用户时长-秒          
expire_date            VARCHAR(10)       False             过期时间- ####-##-##  
user_concur_number     INTEGER           False             用户并发数           
bind_mac               SMALLINT          False             是否绑定mac           
bind_vlan              SMALLINT          False             是否绑定vlan          
mac_addr               VARCHAR(17)       True              mac地址                 
vlan_id                INTEGER           True              内层vlan                
vlan_id2               INTEGER           True              外层vlan                
ip_address             VARCHAR(15)       True              静态IP地址            
last_pause             VARCHAR(19)       True              最后停机时间        
create_time            VARCHAR(19)       False             创建时间              
update_time            VARCHAR(19)       False             更新时间              
=====================  ================  ================  ====================================

.. end_table


.. _slc_rad_account_attr_label:

slc_rad_account_attr
------------------------------------ 

上网账号扩展策略属性表

.. start_table slc_rad_account_attr;id 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
id                     INTEGER           False             属性id                  
account_number         VARCHAR(32)       False             上网账号              
attr_name              VARCHAR(255)      False             属性名                 
attr_value             VARCHAR(255)      False             属性值                 
attr_desc              VARCHAR(255)      True              属性描述              
=====================  ================  ================  ====================================

.. end_table


.. _slc_rad_product_label:

slc_rad_product
------------------------------------ 


    资费信息表 <radiusd default table>
    销售状态 product_status 0 正常 1 停用 资费停用后不允许再订购
    

.. start_table slc_rad_product;id 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
id                     INTEGER           False             资费id                  
product_name           VARCHAR(64)       False             资费名称              
product_policy         INTEGER           False             资费策略              
product_status         SMALLINT          False             资费状态              
bind_mac               SMALLINT          False             是否绑定mac           
bind_vlan              SMALLINT          False             是否绑定vlan          
concur_number          INTEGER           False             并发数                 
fee_period             VARCHAR(11)       True              开放认证时段        
fee_months             INTEGER           True              买断月数              
fee_price              INTEGER           False             资费价格              
input_max_limit        INTEGER           False             上行速率              
output_max_limit       INTEGER           False             下行速率              
create_time            VARCHAR(19)       False             创建时间              
update_time            VARCHAR(19)       False             更新时间              
=====================  ================  ================  ====================================

.. end_table


.. _slc_rad_product_attr_label:

slc_rad_product_attr
------------------------------------ 

资费扩展属性表 <radiusd default table>

.. start_table slc_rad_product_attr;id 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
id                     INTEGER           False             属性id                  
product_id             INTEGER           False             资费id                  
attr_name              VARCHAR(255)      False             属性名                 
attr_value             VARCHAR(255)      False             属性值                 
attr_desc              VARCHAR(255)      True              属性描述              
=====================  ================  ================  ====================================

.. end_table


.. _slc_rad_billing_label:

slc_rad_billing
------------------------------------ 

计费信息表 is_deduct 0 未扣费 1 已扣费 <radiusd default table>

.. start_table slc_rad_billing;id 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
id                     INTEGER           False             计费id                  
account_number         VARCHAR(253)      False             上网账号              
nas_addr               VARCHAR(15)       False             bas地址                 
acct_session_id        VARCHAR(253)      False             会话id                  
acct_start_time        VARCHAR(19)       False             计费开始时间        
acct_session_time      INTEGER           False             会话时长              
acct_length            INTEGER           False             扣费时长              
acct_fee               INTEGER           False             应扣费用              
actual_fee             INTEGER           False             实扣费用              
balance                INTEGER           False             当前余额              
is_deduct              INTEGER           False             是否扣费              
create_time            VARCHAR(19)       False             计费时间              
=====================  ================  ================  ====================================

.. end_table


.. _slc_rad_ticket_label:

slc_rad_ticket
------------------------------------ 

上网日志表 <radiusd default table>

.. start_table slc_rad_ticket;id 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
id                     INTEGER           False             日志id                  
account_number         VARCHAR(253)      False             上网账号              
acct_input_gigawords   INTEGER           True              会话的上行的字（4字节）的吉倍数
acct_output_gigawords  INTEGER           True              会话的下行的字（4字节）的吉倍数
acct_input_octets      INTEGER           True              会话的上行流量（字节数）
acct_output_octets     INTEGER           True              会话的下行流量（字节数）
acct_input_packets     INTEGER           True              会话的上行包数量  
acct_output_packets    INTEGER           True              会话的下行包数量  
acct_session_id        VARCHAR(253)      False             会话id                  
acct_session_time      INTEGER           False             会话时长              
acct_start_time        VARCHAR(19)       False             会话开始时间        
acct_stop_time         VARCHAR(19)       False             会话结束时间        
acct_terminate_cause   INTEGER           True              会话中止原因        
mac_addr               VARCHAR(128)      True              mac地址                 
calling_station_id     VARCHAR(128)      True              用户接入物理信息  
frame_id_netmask       VARCHAR(15)       True              地址掩码              
framed_ipaddr          VARCHAR(15)       True              IP地址                  
nas_class              VARCHAR(253)      True              bas class                 
nas_addr               VARCHAR(15)       False             bas地址                 
nas_port               VARCHAR(32)       True              接入端口              
nas_port_id            VARCHAR(255)      True              接入端口物理信息  
nas_port_type          INTEGER           True              接入端口类型        
service_type           INTEGER           True              接入服务类型        
session_timeout        INTEGER           True              会话超时时间        
start_source           INTEGER           False             会话开始来源        
stop_source            INTEGER           False             会话中止来源        
=====================  ================  ================  ====================================

.. end_table


.. _slc_rad_online_label:

slc_rad_online
------------------------------------ 

用户在线信息表 <radiusd default table>

.. start_table slc_rad_online;id 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
id                     INTEGER           False             在线id                  
account_number         VARCHAR(32)       False             上网账号              
nas_addr               VARCHAR(32)       False             bas地址                 
acct_session_id        VARCHAR(64)       False             会话id                  
acct_start_time        VARCHAR(19)       False             会话开始时间        
framed_ipaddr          VARCHAR(32)       False             IP地址                  
mac_addr               VARCHAR(32)       False             mac地址                 
nas_port_id            VARCHAR(255)      False             接入端口物理信息  
billing_times          INTEGER           False             已记账时间           
start_source           SMALLINT          False             会话开始来源        
=====================  ================  ================  ====================================

.. end_table


.. _slc_rad_online_stat_label:

slc_rad_online_stat
------------------------------------ 

用户在线统计表 <radiusd default table>

.. start_table slc_rad_online_stat;node_id,day_code,time_num 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
node_id                INTEGER           False             区域id                  
day_code               VARCHAR(10)       False             统计日期              
time_num               INTEGER           False             统计小时              
total                  INTEGER           True              在线数                 
=====================  ================  ================  ====================================

.. end_table


.. _slc_rad_accept_log_label:

slc_rad_accept_log
------------------------------------ 


    业务受理日志表
    open:开户 pause:停机 resume:复机 cancel:销户 next:续费 charge:充值
    

.. start_table slc_rad_accept_log;id 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
id                     INTEGER           False             日志id                  
accept_type            VARCHAR(16)       False             受理类型              
accept_desc            VARCHAR(512)      True              受理描述              
account_number         VARCHAR(32)       False             上网账号              
operator_name          VARCHAR(32)       True              操作员名              
accept_source          VARCHAR(128)      True              受理渠道来源        
accept_time            VARCHAR(19)       False             受理时间              
=====================  ================  ================  ====================================

.. end_table


.. _slc_rad_operate_log_label:

slc_rad_operate_log
------------------------------------ 

操作日志表

.. start_table slc_rad_operate_log;id 

=====================  ================  ================  ====================================
属性                    类型（长度）       可否为空           描述                              
=====================  ================  ================  ====================================
id                     INTEGER           False             日志id                  
operator_name          VARCHAR(32)       False             操作员名称           
operate_ip             VARCHAR(128)      True              操作员ip               
operate_time           VARCHAR(19)       False             操作时间              
operate_desc           VARCHAR(1024)     True              操作描述              
=====================  ================  ================  ====================================

.. end_table


