# ToughRADIUS Data Model

## Data Structure Definition for CPEs

```sql
create table net_cpe
(
    id               bigserial
        primary key,
    node_id          bigint,
    system_name      text,
    cpe_type         text,
    sn               text,
    name             text,
    arch_name        text,
    software_version text,
    hardware_version text,
    model            text,
    vendor_code      text,
    oui              text,
    manufacturer     text,
    product_class    text,
    status           text,
    tags             text,
    task_tags        text,
    remark           text,
    uptime           bigint,
    memory_total     bigint,
    memory_free      bigint,
    cpu_usage        bigint,
    cwmp_status      text,
    cwmp_url         text,
    factoryreset_id  text,
    cwmp_last_inform timestamp with time zone,
    created_at       timestamp with time zone,
    updated_at       timestamp with time zone
); 

alter table public.net_cpe
    owner to postgres; 

create index idx_net_cpe_task_tags
    on public.net_cpe (task_tags); 

create index idx_net_cpe_status
    on public.net_cpe (status); 

create unique index idx_net_cpe_sn
    on public.net_cpe (sn); 

create index idx_net_cpe_cwmp_status
    on public.net_cpe (cwmp_status); 

```


## Node Data Structure Definition

```sql
create table net_node (
  id bigint primary key not null default nextval('net_node_id_seq'::regclass),
  name text,
  remark text,
  tags text,
  created_at timestamp with time zone,
  updated_at timestamp with time zone
); 
```


## Data Structure Definition of VPE(Bras)

```sql
create table public.net_vpe
(
    id          bigserial
        primary key,
    node_id     bigint,
    name        text,
    identifier  text,
    hostname    text,
    ipaddr      text,
    secret      text,
    coa_port    bigint,
    model       text,
    vendor_code text,
    status      text,
    tags        text,
    remark      text,
    created_at  timestamp with time zone,
    updated_at  timestamp with time zone
); 

alter table public.net_vpe
    owner to postgres; 

```


## Data Structure Definition of the Radius Profile

```sql
create table public.radius_profile (
  id bigint primary key not null default nextval('radius_profile_id_seq'::regclass),
  node_id bigint,
  name text,
  status text,
  addr_pool text,
  active_num bigint,
  up_rate bigint,
  down_rate bigint,
  remark text,
  created_at timestamp with time zone,
  updated_at timestamp with time zone
); 
create index idx_radius_profile_status on radius_profile using btree (status); 
```



## User's Data Structure Definition

```sql
create table public.radius_user
(
    id          bigserial
        primary key,
    node_id     bigint,
    profile_id  bigint,
    realname    text,
    mobile      text,
    username    text,
    password    text,
    addr_pool   text,
    active_num  bigint,
    up_rate     bigint,
    down_rate   bigint,
    vlanid1     bigint,
    vlanid2     bigint,
    ip_addr     text,
    mac_addr    text,
    bind_vlan   bigint,
    bind_mac    bigint,
    expire_time timestamp with time zone,
    status      text,
    remark      text,
    last_online timestamp with time zone,
    created_at  timestamp with time zone,
    updated_at  timestamp with time zone
); 

alter table public.radius_user
    owner to postgres; 

create unique index idx_radius_user_username
    on public.radius_user (username); 

create index idx_radius_user_profile_id
    on public.radius_user (profile_id); 

create index idx_radius_user_created_at
    on public.radius_user (created_at); 

create index idx_radius_user_status
    on public.radius_user (status); 

create index idx_radius_user_expire_time
    on public.radius_user (expire_time); 

create index idx_radius_user_active_num
    on public.radius_user (active_num); 


```


## radius account log

```sql
create table public.radius_accounting
(
    id                  bigserial
        primary key,
    username            text,
    acct_session_id     text,
    nas_id              text,
    nas_addr            text,
    nas_paddr           text,
    session_timeout     bigint,
    framed_ipaddr       text,
    framed_netmask      text,
    mac_addr            text,
    nas_port            bigint,
    nas_class           text,
    nas_port_id         text,
    nas_port_type       bigint,
    service_type        bigint,
    acct_session_time   bigint,
    acct_input_total    bigint,
    acct_output_total   bigint,
    acct_input_packets  bigint,
    acct_output_packets bigint,
    last_update         timestamp with time zone,
    acct_start_time     timestamp with time zone,
    acct_stop_time      timestamp with time zone
); 

alter table public.radius_accounting
    owner to postgres; 

create index idx_radius_accounting_acct_stop_time
    on public.radius_accounting (acct_stop_time); 

create index idx_radius_accounting_acct_start_time
    on public.radius_accounting (acct_start_time); 

create index idx_radius_accounting_acct_session_id
    on public.radius_accounting (acct_session_id); 

create index idx_radius_accounting_username
    on public.radius_accounting (username); 


```