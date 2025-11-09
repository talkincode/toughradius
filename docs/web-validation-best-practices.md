# Web 参数解析和校验最佳实践

## 当前问题

当前的验证方式存在以下问题：

```go
// 当前方式：手动验证，代码冗长
func CreateProfile(c echo.Context) error {
    var req ProfileRequest
    if err := c.Bind(&req); err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析请求参数", err.Error())
    }

    // 手动验证每个字段
    if profile.Name == "" {
        return fail(c, http.StatusBadRequest, "MISSING_NAME", "Profile 名称不能为空", nil)
    }

    // 需要手动检查更多规则...
}
```

**缺点**：

- ❌ 验证逻辑分散，难以维护
- ❌ 代码重复，每个接口都要写类似代码
- ❌ 缺少统一的错误格式
- ❌ 难以实现复杂的验证规则（如正则、范围、关联验证等）
- ❌ 没有国际化支持

## 推荐方案对比

### 1. go-playground/validator ⭐⭐⭐⭐⭐ (推荐)

**GitHub**: https://github.com/go-playground/validator

**特点**：

- ✅ 最流行的验证库（18k+ stars）
- ✅ 声明式验证（通过 struct tag）
- ✅ 内置 100+ 验证规则
- ✅ 支持自定义验证器
- ✅ 与 Echo 完美集成
- ✅ 支持国际化错误消息
- ✅ 性能优秀

**适用场景**: 所有类型的项目，特别是 RESTful API

### 2. go-ozzo/ozzo-validation ⭐⭐⭐⭐

**GitHub**: https://github.com/go-ozzo/ozzo-validation

**特点**：

- ✅ 代码优先的验证方式
- ✅ 流畅的 API
- ✅ 不依赖 struct tag
- ✅ 易于测试

**适用场景**: 需要动态验证规则或不喜欢 tag 的项目

### 3. asaskevich/govalidator ⭐⭐⭐

**GitHub**: https://github.com/asaskevich/govalidator

**特点**：

- ✅ 简单易用
- ✅ 提供常用验证函数

**适用场景**: 简单的验证需求

## 集成 go-playground/validator 到项目

### 步骤 1: 安装依赖

```bash
go get github.com/go-playground/validator/v10
```

### 步骤 2: 创建验证器包装

创建文件 `internal/pkg/validator/validator.go`:

```go
package validator

import (
    "fmt"
    "strings"

    "github.com/go-playground/validator/v10"
    "github.com/labstack/echo/v4"
)

// CustomValidator Echo 自定义验证器
type CustomValidator struct {
    validator *validator.Validate
}

// NewValidator 创建验证器实例
func NewValidator() *CustomValidator {
    v := validator.New()

    // 注册自定义验证规则
    registerCustomValidations(v)

    return &CustomValidator{validator: v}
}

// Validate 实现 echo.Validator 接口
func (cv *CustomValidator) Validate(i interface{}) error {
    if err := cv.validator.Struct(i); err != nil {
        return echo.NewHTTPError(400, formatValidationError(err))
    }
    return nil
}

// formatValidationError 格式化验证错误为友好消息
func formatValidationError(err error) map[string]interface{} {
    errors := make(map[string]string)

    for _, err := range err.(validator.ValidationErrors) {
        field := strings.ToLower(err.Field())
        switch err.Tag() {
        case "required":
            errors[field] = fmt.Sprintf("%s 不能为空", err.Field())
        case "email":
            errors[field] = "请输入有效的邮箱地址"
        case "min":
            errors[field] = fmt.Sprintf("%s 最小长度为 %s", err.Field(), err.Param())
        case "max":
            errors[field] = fmt.Sprintf("%s 最大长度为 %s", err.Field(), err.Param())
        case "gte":
            errors[field] = fmt.Sprintf("%s 必须大于等于 %s", err.Field(), err.Param())
        case "lte":
            errors[field] = fmt.Sprintf("%s 必须小于等于 %s", err.Field(), err.Param())
        case "oneof":
            errors[field] = fmt.Sprintf("%s 必须是以下值之一: %s", err.Field(), err.Param())
        default:
            errors[field] = fmt.Sprintf("%s 验证失败: %s", err.Field(), err.Tag())
        }
    }

    return map[string]interface{}{
        "error":   "VALIDATION_ERROR",
        "message": "请求参数验证失败",
        "details": errors,
    }
}

// registerCustomValidations 注册自定义验证规则
func registerCustomValidations(v *validator.Validate) {
    // 示例：验证 IP 地址池格式
    v.RegisterValidation("addrpool", func(fl validator.FieldLevel) bool {
        value := fl.Field().String()
        if value == "" {
            return true // 允许空值，使用 required 标签控制必填
        }
        // 简单验证 CIDR 格式
        return strings.Contains(value, "/")
    })

    // 示例：验证状态值
    v.RegisterValidation("radiusstatus", func(fl validator.FieldLevel) bool {
        value := fl.Field().String()
        return value == "enabled" || value == "disabled" || value == ""
    })
}
```

### 步骤 3: 在 webserver 中注册验证器

修改 `internal/webserver/server.go`:

```go
import (
    customValidator "github.com/talkincode/toughradius/v9/internal/pkg/validator"
)

func InitWebServer() {
    e := echo.New()

    // 注册自定义验证器
    e.Validator = customValidator.NewValidator()

    // ... 其他配置
}
```

### 步骤 4: 改进 ProfileRequest 结构

修改 `internal/adminapi/profiles.go`:

```go
// ProfileRequest 用于处理前端发送的混合类型 JSON
type ProfileRequest struct {
    Name       string      `json:"name" validate:"required,min=1,max=100"`
    Status     interface{} `json:"status" validate:"omitempty,oneof=enabled disabled"`
    AddrPool   string      `json:"addr_pool" validate:"omitempty,addrpool"`
    ActiveNum  int         `json:"active_num" validate:"gte=0,lte=100"`
    UpRate     int         `json:"up_rate" validate:"gte=0,lte=1000000"`
    DownRate   int         `json:"down_rate" validate:"gte=0,lte=1000000"`
    Domain     string      `json:"domain" validate:"omitempty,max=50"`
    IPv6Prefix string      `json:"ipv6_prefix" validate:"omitempty,cidrv6"`
    BindMac    interface{} `json:"bind_mac"`
    BindVlan   interface{} `json:"bind_vlan"`
    Remark     string      `json:"remark" validate:"omitempty,max=500"`
    NodeId     interface{} `json:"node_id"`
}

// CreateProfile 创建 RADIUS Profile（改进版）
func CreateProfile(c echo.Context) error {
    var req ProfileRequest

    // Bind 和 Validate 一步完成
    if err := c.Bind(&req); err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析请求参数", err.Error())
    }

    // 自动验证（通过 Echo Validator）
    if err := c.Validate(&req); err != nil {
        // 验证错误已经被格式化
        return err
    }

    // 转换为 RadiusProfile
    profile := req.toRadiusProfile()

    // 业务逻辑验证（如名称唯一性）
    var count int64
    app.GDB().Model(&domain.RadiusProfile{}).Where("name = ?", profile.Name).Count(&count)
    if count > 0 {
        return fail(c, http.StatusConflict, "NAME_EXISTS", "Profile 名称已存在", nil)
    }

    // 设置默认值
    if profile.Status == "" {
        profile.Status = "enabled"
    }

    if err := app.GDB().Create(profile).Error; err != nil {
        return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "创建 Profile 失败", err.Error())
    }

    return ok(c, profile)
}
```

## 常用验证标签

### 基础验证

| 标签        | 说明     | 示例                         |
| ----------- | -------- | ---------------------------- |
| `required`  | 必填字段 | `validate:"required"`        |
| `omitempty` | 允许为空 | `validate:"omitempty,email"` |
| `-`         | 跳过验证 | `validate:"-"`               |

### 字符串验证

| 标签            | 说明           | 示例                                |
| --------------- | -------------- | ----------------------------------- |
| `min=n`         | 最小长度       | `validate:"min=3"`                  |
| `max=n`         | 最大长度       | `validate:"max=100"`                |
| `len=n`         | 固定长度       | `validate:"len=10"`                 |
| `email`         | 邮箱格式       | `validate:"email"`                  |
| `url`           | URL 格式       | `validate:"url"`                    |
| `alpha`         | 只包含字母     | `validate:"alpha"`                  |
| `alphanum`      | 只包含字母数字 | `validate:"alphanum"`               |
| `numeric`       | 只包含数字     | `validate:"numeric"`                |
| `contains=text` | 包含子串       | `validate:"contains=admin"`         |
| `oneof=a b c`   | 枚举值         | `validate:"oneof=enabled disabled"` |

### 数字验证

| 标签    | 说明     | 示例                 |
| ------- | -------- | -------------------- |
| `eq=n`  | 等于     | `validate:"eq=5"`    |
| `ne=n`  | 不等于   | `validate:"ne=0"`    |
| `gt=n`  | 大于     | `validate:"gt=0"`    |
| `gte=n` | 大于等于 | `validate:"gte=1"`   |
| `lt=n`  | 小于     | `validate:"lt=100"`  |
| `lte=n` | 小于等于 | `validate:"lte=100"` |

### 网络验证

| 标签     | 说明      | 示例                |
| -------- | --------- | ------------------- |
| `ip`     | IP 地址   | `validate:"ip"`     |
| `ipv4`   | IPv4 地址 | `validate:"ipv4"`   |
| `ipv6`   | IPv6 地址 | `validate:"ipv6"`   |
| `cidr`   | CIDR 格式 | `validate:"cidr"`   |
| `cidrv4` | IPv4 CIDR | `validate:"cidrv4"` |
| `cidrv6` | IPv6 CIDR | `validate:"cidrv6"` |
| `mac`    | MAC 地址  | `validate:"mac"`    |

### 组合验证

```go
type User struct {
    Username string `validate:"required,min=3,max=50,alphanum"`
    Email    string `validate:"required,email"`
    Age      int    `validate:"required,gte=18,lte=120"`
    Status   string `validate:"oneof=active inactive banned"`
    Website  string `validate:"omitempty,url"`
}
```

## 完整示例：改进后的 Profile API

### profiles.go (完整版)

```go
package adminapi

import (
    "net/http"
    "strconv"

    "github.com/labstack/echo/v4"
    "github.com/talkincode/toughradius/v9/internal/app"
    "github.com/talkincode/toughradius/v9/internal/domain"
)

// ProfileRequest 用于处理前端发送的混合类型 JSON
type ProfileRequest struct {
    Name       string      `json:"name" validate:"required,min=1,max=100"`
    Status     interface{} `json:"status"`
    AddrPool   string      `json:"addr_pool" validate:"omitempty,addrpool"`
    ActiveNum  int         `json:"active_num" validate:"gte=0,lte=100"`
    UpRate     int         `json:"up_rate" validate:"gte=0,lte=10000000"`
    DownRate   int         `json:"down_rate" validate:"gte=0,lte=10000000"`
    Domain     string      `json:"domain" validate:"omitempty,max=50"`
    IPv6Prefix string      `json:"ipv6_prefix" validate:"omitempty"`
    BindMac    interface{} `json:"bind_mac"`
    BindVlan   interface{} `json:"bind_vlan"`
    Remark     string      `json:"remark" validate:"omitempty,max=500"`
    NodeId     interface{} `json:"node_id"`
}

func CreateProfile(c echo.Context) error {
    var req ProfileRequest

    if err := c.Bind(&req); err != nil {
        return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析请求参数", err.Error())
    }

    // 自动验证
    if err := c.Validate(&req); err != nil {
        return err // 已格式化的错误
    }

    profile := req.toRadiusProfile()

    // 业务逻辑验证
    var count int64
    app.GDB().Model(&domain.RadiusProfile{}).Where("name = ?", profile.Name).Count(&count)
    if count > 0 {
        return fail(c, http.StatusConflict, "NAME_EXISTS", "Profile 名称已存在", nil)
    }

    if profile.Status == "" {
        profile.Status = "enabled"
    }

    if err := app.GDB().Create(profile).Error; err != nil {
        return fail(c, http.StatusInternalServerError, "CREATE_FAILED", "创建 Profile 失败", err.Error())
    }

    return ok(c, profile)
}
```

## 优势对比

### 改进前 vs 改进后

| 方面     | 改进前          | 改进后              |
| -------- | --------------- | ------------------- |
| 代码行数 | ~15 行验证代码  | ~0 行（标签声明）   |
| 可维护性 | ❌ 分散在各处   | ✅ 集中在结构体定义 |
| 复用性   | ❌ 每个接口重复 | ✅ 标签可复用       |
| 错误信息 | ❌ 需手动编写   | ✅ 自动生成         |
| 验证规则 | ❌ 有限         | ✅ 100+ 内置规则    |
| 国际化   | ❌ 不支持       | ✅ 支持             |
| 性能     | ⚠️ 一般         | ✅ 优秀             |

## 测试示例

```go
func TestProfileRequestValidation(t *testing.T) {
    validator := customValidator.NewValidator()

    tests := []struct {
        name    string
        request ProfileRequest
        wantErr bool
    }{
        {
            name: "有效的请求",
            request: ProfileRequest{
                Name:      "test-profile",
                ActiveNum: 1,
                UpRate:    1024,
                DownRate:  2048,
            },
            wantErr: false,
        },
        {
            name: "名称为空",
            request: ProfileRequest{
                Name: "",
            },
            wantErr: true,
        },
        {
            name: "速率超出范围",
            request: ProfileRequest{
                Name:   "test",
                UpRate: 99999999,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.Validate(&tt.request)
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## 迁移建议

### 分阶段迁移

1. **第一阶段**: 新接口使用 validator
2. **第二阶段**: 逐步改造现有接口
3. **第三阶段**: 统一错误处理格式

### 兼容性考虑

- ✅ 不影响现有代码
- ✅ 可以逐步迁移
- ✅ 两种方式可以并存

## 参考资源

- [go-playground/validator 官方文档](https://pkg.go.dev/github.com/go-playground/validator/v10)
- [Echo 框架集成示例](https://echo.labstack.com/docs/request#validate-data)
- [自定义验证器示例](https://github.com/go-playground/validator/blob/master/_examples/custom-validation/main.go)

## 总结

使用 `go-playground/validator` 可以：

1. ✅ **减少 80% 的验证代码**
2. ✅ **提高代码可读性和维护性**
3. ✅ **统一错误处理格式**
4. ✅ **提供更丰富的验证规则**
5. ✅ **支持自定义验证逻辑**
6. ✅ **更好的测试支持**

**推荐立即开始在新接口中使用，逐步替换旧接口的手动验证代码。**
