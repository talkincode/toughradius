# Docker-compose 部署

```yml
version: "3"
services:
  loki:
    image: grafana/loki:2.7.1
    container_name: "loki"
    logging:
      driver: "json-file"
      options:
        max-size: "5m"
    restart: always
    environment:
      - TZ=Asia/Shanghai
      - LANG=zh_CN.UTF-8
    command: -config.file=/etc/loki/config.yml
    volumes:
      - /etc/loki/config.yml:/etc/loki/config.yml
      - loki-volume:/loki
    ports:
      - "3100:3100"
    networks:
      loki_network:

networks:
  loki_network:

volumes:
  loki-volume:
```

## loki 配置文件 /etc/loki/config.yml

如果日志文件要保存很久， 可以使用 azure blob 存储， 但是需要注意的是， azure blob 存储的文件名不能包含特殊字符， 例如： `:`

```yml
auth_enabled: false

server:
  http_listen_port: 3100

ingester:
  lifecycler:
    address: 127.0.0.1
    ring:
      kvstore:
        store: inmemory
      replication_factor: 1
    final_sleep: 0s
  chunk_idle_period: 5m
  chunk_retain_period: 30s
  max_transfer_retries: 0
  max_chunk_age: 20m
  wal:
    dir: /loki/wal

schema_config:
  configs:
  - from: "2020-12-11"
    index:
      period: 24h
      prefix: index_
    object_store: azure
    schema: v11
    store: boltdb-shipper

storage_config:
  azure:
    # Your Azure storage account name
    account_name:
    # For the account-key, see docs: https://docs.microsoft.com/en-us/azure/storage/common/storage-account-keys-manage?tabs=azure-portal
    account_key:
    # See https://docs.microsoft.com/en-us/azure/storage/blobs/storage-blobs-introduction#containers
    container_name:
    use_managed_identity:
    # Providing a user assigned ID will override use_managed_identity
    # user_assigned_id: <user-assigned-identity-id>
    request_timeout: 0
    # Configure this if you are using private azure cloud like azure stack hub and will use this endpoint suffix to compose container & blob storage URL. Ex: https://account_name.endpoint_suffix/container_name/blob_name
    endpoint_suffix: core.windows.net

  boltdb_shipper:
    active_index_directory: /loki/boltdb-shipper-active
    cache_location: /loki/boltdb-shipper-cache
    cache_ttl: 24h
    shared_store: azure

  filesystem:
    directory: /loki/chunks

compactor:
  working_directory: /loki/compactor
  shared_store: azure

limits_config:
  enforce_metric_name: false
  reject_old_samples: true
  reject_old_samples_max_age: 168h
  ingestion_rate_mb: 30
  ingestion_burst_size_mb: 15

chunk_store_config:
  max_look_back_period: 168h #回看日志行的最大时间，只适用于即时日志

table_manager:
  retention_deletes_enabled: true #日志保留周期开关，默认为false
  retention_period: 2160h #日志保留周期
```