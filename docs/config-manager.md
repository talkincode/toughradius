# ConfigManager - 灵活的配置管理系统

## 设计目标

解决旧版 `SysConfig` 的以下问题:

1. ✅ **无需改表扩展** - 配置定义在代码中,不需要修改数据库结构
2. ✅ **类型安全** - 支持强类型读取(int/bool/duration/JSON 等)
3. ✅ **自动验证** - 枚举、范围、自定义验证器
4. ✅ **性能优化** - 内置缓存机制,减少数据库查询
5. ✅ **向后兼容** - 不破坏现有代码

## 核心概念

### 1. ConfigSchema - 配置定义

每个配置项通过 `ConfigSchema` 定义元数据:

```go
type ConfigSchema struct {
    Key         string              // 配置键 "category.name"
    Type        ConfigType          // 值类型
    Default     string              // 默认值
    Enum        []string            // 枚举值限制
    Min/Max     *int64              // 数值范围
    Required    bool                // 是否必需
    Description string              // 描述
    Validator   func(string) error  // 自定义验证
}
```

### 2. 数据库结构保持不变

继续使用简单的 KV 结构 (`sys_config` 表):

```
type + name -> value
```

### 3. 配置注册表

所有配置在 `registerSchemas()` 中注册,无需改表:

```go
cm.Register(&ConfigSchema{
    Key:         "radius.MaxSessions",
    Type:        TypeInt,
    Default:     "1",
    Min:         ptrInt64(1),
    Max:         ptrInt64(100),
    Description: "Maximum concurrent sessions per user",
})
```

## 使用指南

### 读取配置

```go
// 字符串配置
eapMethod := app.GApp().ConfigMgr().GetString("radius", "EapMethod")

// 整数配置
maxSessions := app.GApp().ConfigMgr().GetInt("radius", "MaxSessions")

// 布尔配置 (enabled/disabled, true/false, 1/0)
ignorePwd := app.GApp().ConfigMgr().GetBool("radius", "RadiusIgnorePwd")

// 时间间隔配置
interval := app.GApp().ConfigMgr().GetDuration("radius", "CacheTimeout")

// JSON 配置
type RateLimitConfig struct {
    Enabled     bool   `json:"enabled"`
    MaxRequests int    `json:"max_requests"`
    Window      string `json:"window"`
}
var cfg RateLimitConfig
err := app.GApp().ConfigMgr().GetJSON("feature", "RateLimit", &cfg)
```

### 写入配置

```go
// 自动验证(枚举、范围等)
err := app.GApp().ConfigMgr().Set("radius", "EapMethod", "eap-mschapv2")
if err != nil {
    // 验证失败
}
```

### 扩展新配置

只需在 `config_manager.go` 的 `registerSchemas()` 中添加:

```go
// 简单配置
cm.Register(&ConfigSchema{
    Key:         "radius.SessionTimeout",
    Type:        TypeInt,
    Default:     "3600",
    Description: "Default session timeout in seconds",
})

// 枚举配置
cm.Register(&ConfigSchema{
    Key:         "radius.AuthMethod",
    Type:        TypeString,
    Default:     "pap",
    Enum:        []string{"pap", "chap", "eap"},
    Description: "Authentication method",
})

// 范围限制配置
minVal := int64(100)
maxVal := int64(10000)
cm.Register(&ConfigSchema{
    Key:         "radius.MaxRetries",
    Type:        TypeInt,
    Default:     "3",
    Min:         &minVal,
    Max:         &maxVal,
    Description: "Maximum retry attempts",
})

// 自定义验证器
cm.Register(&ConfigSchema{
    Key:     "system.AdminEmail",
    Type:    TypeString,
    Default: "",
    Validator: func(value string) error {
        if !strings.Contains(value, "@") {
            return fmt.Errorf("invalid email format")
        }
        return nil
    },
})

// JSON 配置
cm.Register(&ConfigSchema{
    Key:         "feature.RateLimit",
    Type:        TypeJSON,
    Default:     `{"enabled":true,"max_requests":100,"window":"1m"}`,
    Description: "Rate limiting configuration",
})
```

### 缓存管理

```go
// 清除所有缓存
app.GApp().ConfigMgr().InvalidateCache()

// 清除单个配置缓存
app.GApp().ConfigMgr().InvalidateCacheKey("radius", "EapMethod")

// 配置缓存 TTL (默认 5 分钟)
cm.cacheTTL = 10 * time.Minute
```

## 配置类型

### 支持的类型

| 类型              | Go 类型         | 示例值                       |
| ----------------- | --------------- | ---------------------------- |
| `TypeString`      | `string`        | `"eap-md5"`                  |
| `TypeInt`         | `int`           | `"100"`                      |
| `TypeBool`        | `bool`          | `"enabled"`, `"true"`, `"1"` |
| `TypeDuration`    | `time.Duration` | `"5m"`, `"300s"`             |
| `TypeStringSlice` | `[]string`      | `["a","b","c"]` (JSON)       |
| `TypeJSON`        | `interface{}`   | `{"key":"value"}`            |

### 类型转换规则

#### Bool 类型

- `"enabled"`, `"true"`, `"1"` → `true`
- 其他 → `false`

#### Duration 类型

- 使用 Go 标准格式: `"300s"`, `"5m"`, `"1h30m"`

#### JSON 类型

- 必须是有效的 JSON 字符串
- 使用 `GetJSON()` 解析到目标结构

## 迁移指南

### 从旧版 API 迁移

```go
// 旧方式
value := app.GApp().GetSettingsStringValue("radius", "EapMethod")
intVal := app.GApp().GetSettingsInt64Value("radius", "MaxSessions")

// 新方式 (推荐)
value := app.GApp().ConfigMgr().GetString("radius", "EapMethod")
intVal := app.GApp().ConfigMgr().GetInt("radius", "MaxSessions")
```

### 兼容性

旧版 API 仍然保留,不会破坏现有代码:

- `GetSettingsStringValue()`
- `GetSettingsInt64Value()`
- `GetRadiusSettingsStringValue()`
- `GetSystemSettingsStringValue()`

但建议逐步迁移到新的 `ConfigMgr` API。

## 管理界面集成

### 获取所有配置定义

```go
// 用于动态生成配置表单
schemas := app.GApp().ConfigMgr().GetAllSchemas()

for key, schema := range schemas {
    fmt.Printf("配置: %s\n", key)
    fmt.Printf("  类型: %v\n", schema.Type)
    fmt.Printf("  默认值: %s\n", schema.Default)
    fmt.Printf("  描述: %s\n", schema.Description)

    if len(schema.Enum) > 0 {
        fmt.Printf("  可选值: %v\n", schema.Enum)
    }

    if schema.Min != nil {
        fmt.Printf("  最小值: %d\n", *schema.Min)
    }
    if schema.Max != nil {
        fmt.Printf("  最大值: %d\n", *schema.Max)
    }
}
```

### 前端表单示例

可以根据 `ConfigSchema` 自动生成表单:

```javascript
// React Admin 示例
const ConfigInput = ({ schema }) => {
  switch (schema.Type) {
    case 0: // TypeString
      if (schema.Enum && schema.Enum.length > 0) {
        return (
          <SelectInput choices={schema.Enum.map((e) => ({ id: e, name: e }))} />
        );
      }
      return <TextInput />;

    case 1: // TypeInt
      return <NumberInput min={schema.Min} max={schema.Max} />;

    case 2: // TypeBool
      return <BooleanInput />;

    case 5: // TypeJSON
      return <JsonInput />;

    default:
      return <TextInput />;
  }
};
```

## 性能优化

### 缓存策略

1. **读取缓存**: 配置值缓存 5 分钟(可配置)
2. **写入清除**: 更新配置时自动清除对应缓存
3. **手动清除**: 支持单个或全部缓存清除

### 批量读取优化

如果需要读取多个配置,建议:

```go
// 不推荐: 多次单独读取
eap := cm.GetString("radius", "EapMethod")
pwd := cm.GetBool("radius", "IgnorePwd")
timeout := cm.GetInt("radius", "Timeout")

// 推荐: 定义配置结构体一次性读取
type RadiusConfig struct {
    EapMethod   string
    IgnorePwd   bool
    Timeout     int
}

func GetRadiusConfig() *RadiusConfig {
    cm := app.GApp().ConfigMgr()
    return &RadiusConfig{
        EapMethod: cm.GetString("radius", "EapMethod"),
        IgnorePwd: cm.GetBool("radius", "IgnorePwd"),
        Timeout:   cm.GetInt("radius", "Timeout"),
    }
}
```

## 最佳实践

### 1. 配置命名规范

```
category.ConfigName
```

- `category`: 小写,表示模块(如 `radius`, `system`, `feature`)
- `ConfigName`: 大驼峰,清晰描述功能

示例:

- ✅ `radius.EapMethod`
- ✅ `radius.MaxSessions`
- ✅ `feature.EnableRateLimit`
- ❌ `Radius_eap_method`
- ❌ `max-sessions`

### 2. 默认值设置

- 总是提供合理的默认值
- 默认值应该是"安全"的配置
- 避免使用空字符串作为默认值(除非确实需要)

### 3. 验证策略

- 使用 `Enum` 限制可选值
- 使用 `Min/Max` 限制数值范围
- 使用 `Validator` 实现复杂验证逻辑
- 验证应该在写入时进行,而非读取时

### 4. 废弃配置

对于要删除的配置:

```go
cm.Register(&ConfigSchema{
    Key:         "system.SystemTitle",
    Type:        TypeString,
    Default:     "",
    Description: "[DEPRECATED] Use frontend theme config instead",
})
```

- 标记为 `[DEPRECATED]`
- 保留一段时间后再删除
- 提供迁移建议

## 示例:完整的配置模块

```go
// 1. 注册配置
cm.Register(&ConfigSchema{
    Key:         "radius.VendorSupport",
    Type:        TypeStringSlice,
    Default:     `["14988","2011","9"]`,
    Description: "Supported RADIUS vendor codes",
})

// 2. 读取配置
vendors := app.GApp().ConfigMgr().GetStringSlice("radius", "VendorSupport")

// 3. 使用配置
func isVendorSupported(vendorCode string) bool {
    vendors := app.GApp().ConfigMgr().GetStringSlice("radius", "VendorSupport")
    for _, v := range vendors {
        if v == vendorCode {
            return true
        }
    }
    return false
}

// 4. 更新配置
newVendors := []string{"14988", "2011", "9", "3902"}
jsonData, _ := json.Marshal(newVendors)
err := app.GApp().ConfigMgr().Set("radius", "VendorSupport", string(jsonData))
```

## 总结

**新配置系统的优势:**

1. **灵活扩展** - 无需修改数据库表结构
2. **类型安全** - 编译期和运行时双重保护
3. **自动验证** - 保证配置有效性
4. **高性能** - 内置缓存减少数据库压力
5. **易于管理** - 配置定义集中,便于文档化和管理界面生成

**适用场景:**

- ✅ 运行时可变的配置(用户可通过界面修改)
- ✅ 需要持久化的配置
- ✅ 需要验证的配置
- ❌ 启动时固定的配置(应该放在 `config.yml`,修改需重启服务)
- ❌ 性能极度敏感的配置(考虑在启动时加载到内存)

## 配置更新策略

### config.yml 配置的更新

`config.yml` 中的配置在应用启动时加载,修改后需要重启服务才能生效:

```bash
# 1. 编辑配置文件
vim /etc/toughradius.yml

# 2. 重启服务
systemctl restart toughradius
# 或
./toughradius -c toughradius.yml
```

**适合放在 config.yml 的配置:**

- 数据库连接信息 (host, port, user, password)
- 服务监听端口 (web.port, radiusd.auth_port)
- 工作目录 (system.workdir)
- 日志配置 (logger.\*)
- TLS 证书路径

**优点:**

- 环境隔离(开发/测试/生产)
- 版本控制(可通过 git 管理)
- 容器化部署友好(ConfigMap/Secret)

### sys_config 配置的更新

使用 ConfigManager 的配置可以通过 Web 界面或 API 实时更新:

```go
// 通过 API 更新
PUT /api/system/settings/:id
{
  "value": "new-value"
}

// 自动验证 + 更新
err := app.GApp().ConfigMgr().Set("radius", "EapMethod", "eap-mschapv2")
```

**实时生效机制:**

1. **立即生效** (缓存失效后下次读取):

```go
func updateEapMethod(newValue string) {
    // 更新配置
    app.GApp().ConfigMgr().Set("radius", "EapMethod", newValue)
    // 缓存自动失效,下次 Get 时读到新值
}
```

2. **主动推送** (配置变更监听):

```go
// 在 RadiusService 中监听配置变更
type RadiusService struct {
    eapMethod string
}

func (s *RadiusService) OnConfigChanged(category, name, value string) {
    if category == "radius" && name == "EapMethod" {
        s.eapMethod = value
        zap.L().Info("EAP method updated", zap.String("value", value))
    }
}

// 更新配置时触发通知
func updateConfigWithNotify(category, name, value string) error {
    if err := app.GApp().ConfigMgr().Set(category, name, value); err != nil {
        return err
    }

    // 通知所有服务配置已更新
    notifyConfigChanged(category, name, value)
    return nil
}
```

3. **重载机制** (针对高性能场景):

```go
// RadiusService 启动时加载配置到内存
type RadiusService struct {
    config *RadiusConfig  // 内存缓存
}

type RadiusConfig struct {
    EapMethod   string
    MaxSessions int
    // ... 其他高频访问的配置
}

func (s *RadiusService) LoadConfig() {
    cm := app.GApp().ConfigMgr()
    s.config = &RadiusConfig{
        EapMethod:   cm.GetString("radius", "EapMethod"),
        MaxSessions: cm.GetInt("radius", "MaxSessions"),
    }
}

// 提供 reload 接口
func (s *RadiusService) ReloadConfig() {
    s.LoadConfig()
    zap.L().Info("radius config reloaded")
}

// API 触发重载
POST /api/system/reload
```

### 混合策略(推荐)

对于某些配置,可以同时支持 `config.yml` 和 `sys_config`:

```go
func GetEapMethod() string {
    // 优先从数据库读取
    dbValue := app.GApp().ConfigMgr().GetString("radius", "EapMethod")
    if dbValue != "" {
        return dbValue
    }

    // 降级到配置文件默认值
    return "eap-md5"
}
```

**配置优先级:**

1. 数据库配置 (sys_config) - 最高优先级,可在线修改
2. 环境变量 - 部署时覆盖
3. 配置文件 (config.yml) - 默认值

### 不重启更新 config.yml 的方案

如果确实需要不重启更新 config.yml 中的配置,可以实现配置热重载:

```go
// config/reload.go
type ConfigReloader struct {
    configFile string
    app        *app.Application
    watcher    *fsnotify.Watcher
}

func (r *ConfigReloader) WatchConfig() {
    watcher, _ := fsnotify.NewWatcher()
    watcher.Add(r.configFile)

    for {
        select {
        case event := <-watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                r.reloadConfig()
            }
        }
    }
}

func (r *ConfigReloader) reloadConfig() {
    newCfg := config.LoadConfig(r.configFile)

    // 只更新可热更新的配置
    if r.canHotReload("logger.level") {
        r.updateLogLevel(newCfg.Logger.Mode)
    }

    zap.L().Info("config reloaded from file")
}
```

但这种方式复杂度高,一般只用于特定场景。

### 决策树

```
需要配置更新?
├─ 是否需要重启服务才能生效?
│  ├─ 是 → config.yml (数据库连接、服务端口等)
│  └─ 否 → sys_config (ConfigManager)
│
├─ 访问频率如何?
│  ├─ 极高(每秒1000+) → 内存缓存 + reload 接口
│  ├─ 中频(每秒10-100) → ConfigManager (5分钟缓存)
│  └─ 低频(定时任务等) → ConfigManager
│
└─ 是否需要不同环境不同值?
   ├─ 是 → config.yml + 环境变量
   └─ 否 → sys_config (用户可自行调整)
```
