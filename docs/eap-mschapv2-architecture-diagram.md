# ToughRADIUS EAP-MSCHAPv2 架构流程图

## 认证流程时序图

### 第一阶段：Identity Exchange

```
    客户端                   NAS                    RADIUS Server (1812)
      │                       │                              │
      │  EAP-Response/       │                              │
      │  Identity            │                              │
      │  "username"          │                              │
      ├──────────────────────→                              │
      │                       │  Access-Request             │
      │                       │  + EAP-Message              │
      │                       ├──────────────────────────────→
      │                       │                              │
      │                       │                    ┌─────────────────────┐
      │                       │                    │ AuthService.        │
      │                       │                    │ ServeRADIUS()       │
      │                       │                    └─────────┬───────────┘
      │                       │                              │
      │                       │                    ┌─────────▼───────────┐
      │                       │                    │ ParseEAPMessage()   │
      │                       │                    │ Code=2, Type=1      │
      │                       │                    └─────────┬───────────┘
      │                       │                              │
      │                       │                    ┌─────────▼───────────┐
      │                       │                    │ eapHelper.Handle    │
      │                       │                    │ EAPAuthentication() │
      │                       │                    └─────────┬───────────┘
      │                       │                              │
      │                       │                    ┌─────────▼───────────┐
      │                       │                    │ Coordinator.Handle  │
      │                       │                    │ EAPRequest()        │
      │                       │                    └─────────┬───────────┘
      │                       │                              │
      │                       │                    ┌─────────▼──────────────────────┐
      │                       │                    │ handleIdentityResponse()       │
      │                       │                    │ ├─ GetEapMethod()              │
      │                       │                    │ │  → "eap-mschapv2"           │
      │                       │                    │ └─ handlerRegistry.GetHandler  │
      │                       │                    │    (TypeMSCHAPv2=26)           │
      │                       │                    └─────────┬──────────────────────┘
      │                       │                              │
      │                       │                    ┌─────────▼──────────────────────┐
      │                       │                    │ MSCHAPv2Handler.               │
      │                       │                    │ HandleIdentity()               │
      │                       │                    │                                │
      │                       │                    │ ┌────────────────────────────┐ │
      │                       │                    │ │1. GenerateRandomBytes(16)  │ │
      │                       │                    │ │   Challenge=[0x1a,0x2b...] │ │
      │                       │                    │ └────────────────────────────┘ │
      │                       │                    │ ┌────────────────────────────┐ │
      │                       │                    │ │2. buildChallengeRequest()  │ │
      │                       │                    │ │   [Request|ID|Len|Type|    │ │
      │                       │                    │ │    OpCode|Challenge|Name]  │ │
      │                       │                    │ └────────────────────────────┘ │
      │                       │                    │ ┌────────────────────────────┐ │
      │                       │                    │ │3. 创建 EAPState             │ │
      │                       │                    │ │   StateID = UUID()         │ │
      │                       │                    │ │   {Username, Challenge,    │ │
      │                       │                    │ │    Method, Success=false}  │ │
      │                       │                    │ └────────────────────────────┘ │
      │                       │                    │ ┌────────────────────────────┐ │
      │                       │                    │ │4. StateManager.SetState()  │ │
      │                       │                    │ │   states["f3e2..."] = state│ │
      │                       │                    │ └────────────────────────────┘ │
      │                       │                    └─────────┬──────────────────────┘
      │                       │                              │
      │                       │   Access-Challenge           │
      │                       │   + State: "f3e2d1c0..."     │
      │                       │   + EAP-Message:             │
      │                       │     [MSCHAPv2-Challenge]     │
      │                       ←──────────────────────────────┤
      │  EAP-Request/        │                              │
      │  MSCHAPv2-Challenge  │                              │
      │  Challenge=[16 bytes]│                              │
      ←──────────────────────┤                              │
      │                       │                              │
```

### 第二阶段：MSCHAPv2 Response & Verify

```
      │                       │                              │
      │ [客户端计算]          │                              │
      │ ┌──────────────────┐ │                              │
      │ │ PeerChallenge =  │ │                              │
      │ │ Random(16)       │ │                              │
      │ │                  │ │                              │
      │ │ NTResponse =     │ │                              │
      │ │ GenerateNTResp() │ │                              │
      │ │ (AuthChallenge,  │ │                              │
      │ │  PeerChallenge,  │ │                              │
      │ │  Username, Pwd)  │ │                              │
      │ └──────────────────┘ │                              │
      │                       │                              │
      │  EAP-Response/       │                              │
      │  MSCHAPv2-Response   │                              │
      │  + PeerChallenge     │                              │
      │  + NTResponse[24]    │                              │
      ├──────────────────────→                              │
      │                       │  Access-Request             │
      │                       │  + State: "f3e2d1c0..."     │
      │                       │  + EAP-Message              │
      │                       ├──────────────────────────────→
      │                       │                              │
      │                       │                    ┌─────────────────────┐
      │                       │                    │ ServeRADIUS()       │
      │                       │                    │ ParseEAPMessage()   │
      │                       │                    │ Code=2, Type=26     │
      │                       │                    └─────────┬───────────┘
      │                       │                              │
      │                       │                    ┌─────────▼───────────┐
      │                       │                    │ Coordinator.handle  │
      │                       │                    │ ChallengeResponse() │
      │                       │                    └─────────┬───────────┘
      │                       │                              │
      │                       │                    ┌─────────▼──────────────────────┐
      │                       │                    │ 1. Get State from Packet       │
      │                       │                    │    stateID = State_GetString() │
      │                       │                    │                                │
      │                       │                    │ 2. StateManager.GetState()     │
      │                       │                    │    → {Username, Challenge...}  │
      │                       │                    └─────────┬──────────────────────┘
      │                       │                              │
      │                       │                    ┌─────────▼──────────────────────┐
      │                       │                    │ MSCHAPv2Handler.               │
      │                       │                    │ HandleResponse()               │
      │                       │                    │                                │
      │                       │                    │ ┌────────────────────────────┐ │
      │                       │                    │ │1. parseResponse()          │ │
      │                       │                    │ │   → PeerChallenge[16]      │ │
      │                       │                    │ │   → NTResponse[24]         │ │
      │                       │                    │ │   → Flags, Name            │ │
      │                       │                    │ └────────────────────────────┘ │
      │                       │                    │ ┌────────────────────────────┐ │
      │                       │                    │ │2. pwdProvider.GetPassword()│ │
      │                       │                    │ │   → "password123"          │ │
      │                       │                    │ └────────────────────────────┘ │
      │                       │                    │ ┌────────────────────────────┐ │
      │                       │                    │ │3. verifyResponse()         │ │
      │                       │                    │ │   ┌──────────────────────┐ │ │
      │                       │                    │ │   │ RFC 2759:            │ │ │
      │                       │                    │ │   │ ExpectedNT =         │ │ │
      │                       │                    │ │   │ GenerateNTResponse() │ │ │
      │                       │                    │ │   │ (AuthChal, PeerChal, │ │ │
      │                       │                    │ │   │  User, Pwd)          │ │ │
      │                       │                    │ │   └──────────────────────┘ │ │
      │                       │                    │ │   ┌──────────────────────┐ │ │
      │                       │                    │ │   │ bytes.Equal(         │ │ │
      │                       │                    │ │   │   ExpectedNT,        │ │ │
      │                       │                    │ │   │   NTResponse)        │ │ │
      │                       │                    │ │   │ → ✓ MATCH!           │ │ │
      │                       │                    │ │   └──────────────────────┘ │ │
      │                       │                    │ │   ┌──────────────────────┐ │ │
      │                       │                    │ │   │ RFC 3079:            │ │ │
      │                       │                    │ │   │ RecvKey = MakeKey()  │ │ │
      │                       │                    │ │   │ SendKey = MakeKey()  │ │ │
      │                       │                    │ │   └──────────────────────┘ │ │
      │                       │                    │ │   ┌──────────────────────┐ │ │
      │                       │                    │ │   │ AuthResponse =       │ │ │
      │                       │                    │ │   │ GenerateAuthResp()   │ │ │
      │                       │                    │ │   │ → "S=..."[42 bytes]  │ │ │
      │                       │                    │ │   └──────────────────────┘ │ │
      │                       │                    │ │   ┌──────────────────────┐ │ │
      │                       │                    │ │   │ Add MS Attributes:   │ │ │
      │                       │                    │ │   │ - MSCHAP2-Success    │ │ │
      │                       │                    │ │   │ - MPPE-Recv-Key      │ │ │
      │                       │                    │ │   │ - MPPE-Send-Key      │ │ │
      │                       │                    │ │   │ - Encryption-Policy  │ │ │
      │                       │                    │ │   │ - Encryption-Types   │ │ │
      │                       │                    │ │   └──────────────────────┘ │ │
      │                       │                    │ └────────────────────────────┘ │
      │                       │                    │ ┌────────────────────────────┐ │
      │                       │                    │ │4. state.Success = true     │ │
      │                       │                    │ │   StateManager.SetState()  │ │
      │                       │                    │ └────────────────────────────┘ │
      │                       │                    └─────────┬──────────────────────┘
      │                       │                              │
      │                       │                    ┌─────────▼──────────────────────┐
      │                       │                    │ AuthenticateUserWithPlugins()  │
      │                       │                    │ ├─ SkipPasswordValidation()    │
      │                       │                    │ ├─ Check MaxSessions           │
      │                       │                    │ ├─ Check Expiration            │
      │                       │                    │ ├─ Check MAC Binding           │
      │                       │                    │ └─ Check VLAN Binding          │
      │                       │                    └─────────┬──────────────────────┘
      │                       │                              │
      │                       │                    ┌─────────▼──────────────────────┐
      │                       │                    │ SendEAPSuccess()               │
      │                       │                    │ ┌────────────────────────────┐ │
      │                       │                    │ │ EncodeEAPHeader(Success)   │ │
      │                       │                    │ │ [Code=3|ID|Len=4]          │ │
      │                       │                    │ └────────────────────────────┘ │
      │                       │                    │ ┌────────────────────────────┐ │
      │                       │                    │ │ SetEAPMessageAndAuth()     │ │
      │                       │                    │ │ - EAP-Message              │ │
      │                       │                    │ │ - Message-Authenticator    │ │
      │                       │                    │ └────────────────────────────┘ │
      │                       │                    └─────────┬──────────────────────┘
      │                       │                              │
      │                       │   Access-Accept              │
      │                       │   + EAP-Message: Success     │
      │                       │   + MS-MSCHAP2-Success       │
      │                       │   + MS-MPPE-Recv-Key         │
      │                       │   + MS-MPPE-Send-Key         │
      │                       │   + [其他厂商属性]           │
      │                       ←──────────────────────────────┤
      │  EAP-Success         │                              │
      ←──────────────────────┤                              │
      │                       │                              │
      │                       │                    ┌─────────▼───────────┐
      │                       │                    │ CleanupState()      │
      │                       │                    │ DeleteState(StateID)│
      │                       │                    └─────────────────────┘
      │                       │                              │
  ┌───▼────┐                 │                              │
  │已认证  │                 │                              │
  └────────┘                 │                              │
```

## 组件架构分层视图

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                           RADIUS Protocol Layer                                      │
│  ┌─────────────────────────────────────────────────────────────────────────────┐    │
│  │  AuthService.ServeRADIUS()  (internal/radiusd/radius_auth.go)              │    │
│  │  ├─ 接收 RADIUS Access-Request (UDP 1812)                                   │    │
│  │  ├─ 解析用户名、NAS、State 属性                                              │    │
│  │  ├─ 检测 EAP 消息 (ParseEAPMessage)                                         │    │
│  │  └─ 路由到 EAP 或传统认证流程                                               │    │
│  └─────────────────────────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────────────────────────┘
                                         │
                                         ↓
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                            EAP Framework Layer                                       │
│  ┌─────────────────────────────────────────────────────────────────────────────┐    │
│  │  EAPAuthHelper  (internal/radiusd/eap_helper.go)                            │    │
│  │  └─ HandleEAPAuthentication() - 主入口                                       │    │
│  └─────────────────────────────────────────────────────────────────────────────┘    │
│                                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────────────┐    │
│  │  Coordinator  (internal/radiusd/plugins/eap/coordinator.go)                 │    │
│  │  ├─ HandleEAPRequest() - 消息分发总控                                        │    │
│  │  ├─ handleIdentityResponse() - Identity 阶段                                │    │
│  │  ├─ handleChallengeResponse() - Challenge 阶段                              │    │
│  │  ├─ handleNak() - NAK 处理                                                  │    │
│  │  ├─ SendEAPSuccess() - 发送成功                                             │    │
│  │  └─ SendEAPFailure() - 发送失败                                             │    │
│  └─────────────────────────────────────────────────────────────────────────────┘    │
│                                                                                      │
│  ┌────────────────────┐  ┌─────────────────────┐  ┌──────────────────────────┐     │
│  │ EAPStateManager    │  │ PasswordProvider    │  │ HandlerRegistry          │     │
│  │ (interfaces.go)    │  │ (interfaces.go)     │  │ (interfaces.go)          │     │
│  └─────────┬──────────┘  └──────────┬──────────┘  └───────────┬──────────────┘     │
│            │                        │                          │                     │
│            ↓                        ↓                          ↓                     │
│  ┌────────────────────┐  ┌─────────────────────┐  ┌──────────────────────────┐     │
│  │MemoryStateManager  │  │DefaultPasswordProv  │  │ Registry                 │     │
│  │(memory_state_...)  │  │(password_...)       │  │(registry.go)             │     │
│  │- states map        │  │- GetPassword()      │  │- eapHandlers map[Type]   │     │
│  │- SetState()        │  │- MAC/Normal区分     │  │- RegisterEAPHandler()    │     │
│  │- GetState()        │  │                     │  │- GetHandler()            │     │
│  │- DeleteState()     │  │                     │  │                          │     │
│  └────────────────────┘  └─────────────────────┘  └──────────────────────────┘     │
└──────────────────────────────────────────────────────────────────────────────────────┘
                                         │
                                         ↓
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                           EAP Handler Layer (Pluggable)                              │
│                                                                                      │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────────────────┐      │
│  │ EAPHandler       │  │ EAPHandler       │  │ EAPHandler                   │      │
│  │ Interface        │  │ Interface        │  │ Interface                    │      │
│  └─────────┬────────┘  └─────────┬────────┘  └─────────┬────────────────────┘      │
│            │                     │                      │                            │
│            ↓                     ↓                      ↓                            │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────────────────┐      │
│  │ MD5Handler       │  │ OTPHandler       │  │ MSCHAPv2Handler              │      │
│  │ (Type 4)         │  │ (Type 5)         │  │ (Type 26)                    │      │
│  │ md5_handler.go   │  │ otp_handler.go   │  │ mschapv2_handler.go          │      │
│  │                  │  │                  │  │                              │      │
│  │- Name()          │  │- Name()          │  │- Name()                      │      │
│  │- EAPType()       │  │- EAPType()       │  │- EAPType()                   │      │
│  │- CanHandle()     │  │- CanHandle()     │  │- CanHandle()                 │      │
│  │- HandleIdentity()│  │- HandleIdentity()│  │- HandleIdentity()            │      │
│  │- HandleResponse()│  │- HandleResponse()│  │  ├─ GenerateRandomBytes(16)  │      │
│  │                  │  │                  │  │  ├─ buildChallengeRequest()  │      │
│  │                  │  │                  │  │  └─ CreateState & Save       │      │
│  │                  │  │                  │  │- HandleResponse()            │      │
│  │                  │  │                  │  │  ├─ parseResponse()          │      │
│  │                  │  │                  │  │  ├─ GetPassword()            │      │
│  │                  │  │                  │  │  └─ verifyResponse()         │      │
│  │                  │  │                  │  │     ├─ RFC 2759 Algorithms   │      │
│  │                  │  │                  │  │     ├─ RFC 3079 MPPE Keys    │      │
│  │                  │  │                  │  │     └─ Add MS Attributes     │      │
│  └──────────────────┘  └──────────────────┘  └──────────────────────────────┘      │
└──────────────────────────────────────────────────────────────────────────────────────┘
                                         │
                                         ↓
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                          Vendor & Crypto Layer                                       │
│  ┌─────────────────────────────────────────────────────────────────────────────┐    │
│  │  Microsoft Vendor Attributes  (vendors/microsoft/generated.go)              │    │
│  │  ├─ VendorID: 311                                                           │    │
│  │  ├─ MSCHAP2Success_Add()                                                    │    │
│  │  ├─ MSMPPERecvKey_Add()                                                     │    │
│  │  ├─ MSMPPESendKey_Add()                                                     │    │
│  │  ├─ MSMPPEEncryptionPolicy_Add()                                            │    │
│  │  └─ MSMPPEEncryptionTypes_Add()                                             │    │
│  └─────────────────────────────────────────────────────────────────────────────┘    │
│                                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────────────┐    │
│  │  layeh.com/radius Library                                                   │    │
│  │  ├─ rfc2759: GenerateNTResponse(), GenerateAuthenticatorResponse()         │    │
│  │  ├─ rfc3079: MakeKey() (MPPE Key Derivation)                               │    │
│  │  ├─ rfc2865: User-Name, State, etc.                                        │    │
│  │  └─ rfc2869: EAP-Message, Message-Authenticator                            │    │
│  └─────────────────────────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

## 状态管理流程

```
                           ┌─────────────────────────────┐
                           │  MemoryStateManager         │
                           │  states: map[string]*State  │
                           └──────────┬──────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    │                 │                 │
                    ↓                 ↓                 ↓
         ┌──────────────────┐ ┌──────────────┐ ┌──────────────┐
         │ StateID: "f3e2..." │ │ StateID: ... │ │ StateID: ... │
         └──────────────────┘ └──────────────┘ └──────────────┘
                    │
                    ↓
         ┌────────────────────────────────────────────┐
         │ EAPState {                                 │
         │   Username:  "testuser"                    │
         │   Challenge: [0x1a, 0x2b, ..., 0xff]  16B  │
         │   StateID:   "f3e2d1c0-b4a5-..."           │
         │   Method:    "eap-mschapv2"                │
         │   Success:   false → true                  │
         │   Data: {                                  │
         │     "ms_identifier": 1                     │
         │   }                                        │
         │ }                                          │
         └────────────────────────────────────────────┘

状态生命周期:
  1. HandleIdentity  → SetState(stateID, {Success: false})
  2. HandleResponse  → GetState(stateID) → 验证 → SetState(stateID, {Success: true})
  3. SendEAPSuccess  → CleanupState() → DeleteState(stateID)
```

## MSCHAPv2 密码验证算法流程 (RFC 2759)

```
输入:
  - AuthenticatorChallenge (服务器 Challenge)    [16 bytes]
  - PeerChallenge (客户端 Challenge)              [16 bytes]
  - Username                                      [variable]
  - Password                                      [variable]
  - NTResponse (客户端发送的响应)                 [24 bytes]

步骤:
  ┌──────────────────────────────────────────────────────────────┐
  │ 1. 服务器计算期望的 NT-Response                              │
  │    ExpectedNTResponse = GenerateNTResponse(                  │
  │        AuthenticatorChallenge,                               │
  │        PeerChallenge,                                        │
  │        Username,                                             │
  │        Password                                              │
  │    )                                                         │
  │    → [24 bytes]                                              │
  └──────────────────────────────────────────────────────────────┘
                         │
                         ↓
  ┌──────────────────────────────────────────────────────────────┐
  │ 2. 验证客户端响应                                            │
  │    if bytes.Equal(ExpectedNTResponse, NTResponse) {          │
  │        // 密码正确                                           │
  │    } else {                                                  │
  │        // 密码错误 → Reject                                  │
  │    }                                                         │
  └──────────────────────────────────────────────────────────────┘
                         │ (密码正确)
                         ↓
  ┌──────────────────────────────────────────────────────────────┐
  │ 3. 生成 Authenticator Response (RFC 2759)                    │
  │    AuthenticatorResponse = GenerateAuthenticatorResponse(    │
  │        AuthenticatorChallenge,                               │
  │        PeerChallenge,                                        │
  │        ExpectedNTResponse,                                   │
  │        Username,                                             │
  │        Password                                              │
  │    )                                                         │
  │    → "S=<42 hex characters>"                                 │
  └──────────────────────────────────────────────────────────────┘
                         │
                         ↓
  ┌──────────────────────────────────────────────────────────────┐
  │ 4. 生成 MPPE 密钥 (RFC 3079)                                 │
  │    RecvKey = MakeKey(ExpectedNTResponse, Password, false)    │
  │    SendKey = MakeKey(ExpectedNTResponse, Password, true)     │
  │    → 用于后续数据加密                                        │
  └──────────────────────────────────────────────────────────────┘
                         │
                         ↓
  ┌──────────────────────────────────────────────────────────────┐
  │ 5. 添加 Microsoft 厂商属性到 RADIUS 响应                     │
  │    - MS-MSCHAP2-Success: [ID(1B) + AuthResp(42B)]           │
  │    - MS-MPPE-Recv-Key: RecvKey                              │
  │    - MS-MPPE-Send-Key: SendKey                              │
  │    - MS-MPPE-Encryption-Policy: EncryptionAllowed           │
  │    - MS-MPPE-Encryption-Types: RC4-40or128BitAllowed        │
  └──────────────────────────────────────────────────────────────┘
```

## 配置与插件注册流程

```
┌────────────────────────────────────────────────────────────────────────────────┐
│ main.go                                                                        │
│ ├─ InitPlugins(sessionRepo, accountingRepo)                                   │
│ │  └─ internal/radiusd/plugins/init.go                                        │
│ │     ├─ registry.RegisterEAPHandler(eaphandlers.NewMD5Handler())             │
│ │     ├─ registry.RegisterEAPHandler(eaphandlers.NewOTPHandler())             │
│ │     └─ registry.RegisterEAPHandler(eaphandlers.NewMSCHAPv2Handler())  ✓     │
│ │                                                                              │
│ └─ radiusService := radiusd.NewRadiusService()                                │
│    └─ ListenRadiusAuthServer(NewAuthService(radiusService))                   │
└────────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ↓
┌────────────────────────────────────────────────────────────────────────────────┐
│ Global Registry (internal/radiusd/registry/registry.go)                       │
│                                                                                │
│ type Registry struct {                                                        │
│     eapHandlers map[uint8]eap.EAPHandler                                      │
│     mu          sync.RWMutex                                                  │
│ }                                                                              │
│                                                                                │
│ eapHandlers:                                                                  │
│   ├─ [4]  → &MD5Handler{}                                                     │
│   ├─ [5]  → &OTPHandler{}                                                     │
│   └─ [26] → &MSCHAPv2Handler{}  ✓                                             │
└────────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ↓
┌────────────────────────────────────────────────────────────────────────────────┐
│ 配置读取 (database sys_config table)                                          │
│                                                                                │
│ SELECT value FROM sys_config WHERE name = 'RadiusEapMethod';                  │
│ → "eap-mschapv2"                                                               │
│                                                                                │
│ GetEapMethod() 调用链:                                                         │
│ AuthService → app.GApp().GetSettingsStringValue("radius", "RadiusEapMethod")  │
└────────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ↓
┌────────────────────────────────────────────────────────────────────────────────┐
│ 运行时查找 Handler                                                             │
│                                                                                │
│ handleIdentityResponse() {                                                    │
│     eapMethod := GetEapMethod()  // "eap-mschapv2"                            │
│     switch eapMethod {                                                        │
│         case "eap-mschapv2":                                                  │
│             handler = handlerRegistry.GetHandler(TypeMSCHAPv2)  // Type 26    │
│             // → 返回 &MSCHAPv2Handler{}                                      │
│     }                                                                          │
│ }                                                                              │
└────────────────────────────────────────────────────────────────────────────────┘
```

## 总结

以上架构流程图展示了 ToughRADIUS 中 EAP-MSCHAPv2 的完整实现，包括：

- **三阶段认证流程**：Identity → Challenge → Response/Verify → Success
- **组件分层架构**：RADIUS Layer → EAP Framework → Handler Layer → Vendor/Crypto
- **状态管理机制**：基于内存的 StateID 映射，完整的生命周期管理
- **密码验证算法**：严格遵循 RFC 2759 和 RFC 3079 标准
- **配置与注册**：从数据库读取配置，通过全局注册表动态查找 Handler

整个实现采用**插件化架构**，通过接口抽象和注册表模式，实现了高度的扩展性和可维护性。
