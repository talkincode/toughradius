# LDAP / AD Authentication Backend

> 中文版本：[LDAP / AD 认证后端](../zh/auth-ldap.md)

ToughRADIUS can verify a user's password against an external LDAP directory or
Microsoft Active Directory by performing an LDAP **bind**, instead of (or in
addition to) the password stored in its own database. This lets you reuse an
existing corporate directory without copying or migrating passwords.

> **PAP-family only.** The LDAP backend authenticates by binding with the
> cleartext password, so it serves only **bare PAP** and **EAP-TTLS inner PAP**
> (RFC 5281). Challenge/response methods — CHAP, MS-CHAP, MS-CHAPv2, EAP-MD5 and
> PEAP-MSCHAPv2 — **cannot** be served, because the server never holds the
> cleartext password it would need to recompute their responses. When LDAP is
> enabled those methods are deliberately rejected with a diagnostic reason; see
> [How authentication flows](#how-authentication-flows) below.

## When to use it (and when not to)

Use the LDAP backend when you must serve users whose passwords already live in
a directory — an OpenLDAP server, FreeIPA, or Active Directory — and you do not
want to provision a certificate per client. It is the pragmatic bridge for
mixed and legacy estates: protect the tunnel with a server certificate, then
send the username and password inside it.

Be honest about the trade-off. **EAP-TTLS/PAP transmits the cleartext password
inside the TLS tunnel** — it is protected *only* by that tunnel. The security of
the deployment therefore rests on a strong, properly verified server
certificate and TLS 1.2+. If every client can present a certificate, prefer
EAP-TLS. If you need directory-backed passwords for legacy clients, LDAP +
EAP-TTLS/PAP is the right tool — just size your expectations accordingly.

## Deployment model

The LDAP backend **replaces only the password check**. It does **not** create
accounts or carry authorization data. For every user you still need a local
`RadiusUser` row, because authorization (profile/plan, rate limits, expiry,
concurrent-session limit, address pool, VLAN, MAC binding) is loaded from the
local database *before* authentication runs.

In other words:

- **Authentication** (is the password correct?) → the LDAP directory, by bind.
- **Authorization** (what is this user allowed to do?) → the local `RadiusUser`
  row, exactly as without LDAP.

A global `ldap.Enabled` switch turns the backend on or off for the whole server;
there is no per-user "authentication source" field. MAC-address authentication
always bypasses LDAP.

## Configuration

Configure the backend on the **System Configuration** page of the admin UI,
under the **LDAP** group. All items are also editable through the settings API.
The backend re-reads its configuration on every authentication attempt, so
changes take effect immediately without a restart.

| Key | Type | Default | Applies to | Description |
| --- | --- | --- | --- | --- |
| `ldap.Enabled` | bool | `false` | both | Turn the LDAP backend on. Off by default. |
| `ldap.ServerURL` | string | _(empty)_ | both | Directory URL, e.g. `ldap://dc.example.com:389` or `ldaps://dc.example.com:636`. |
| `ldap.BindMode` | enum | `template` | both | `template` or `search` (see below). |
| `ldap.BindDNTemplate` | string | _(empty)_ | template | DN template with a single `%s` for the username, e.g. `uid=%s,ou=people,dc=example,dc=com` or the AD UPN form `%s@example.com`. |
| `ldap.BaseDN` | string | _(empty)_ | search | Subtree base for the user lookup, e.g. `dc=example,dc=com`. |
| `ldap.UserFilter` | string | `(uid=%s)` | search | Filter with a single `%s` for the username (escaped before substitution), e.g. `(uid=%s)` or AD `(sAMAccountName=%s)`. |
| `ldap.SearchBindDN` | string | _(empty)_ | search | DN of the read-only service account used to find users, e.g. `cn=svc-radius,ou=svc,dc=example,dc=com`. |
| `ldap.SearchBindPassword` | string | _(empty)_ | search | Password for the service-account DN. |
| `ldap.StartTLS` | bool | `false` | both | Upgrade an `ldap://` connection to TLS with StartTLS (RFC 4513 §3) before binding. Leave off for `ldaps://`. |
| `ldap.TLSSkipVerify` | bool | `false` | both | Skip TLS certificate verification. **Insecure — lab / self-signed only.** |
| `ldap.Timeout` | int (s) | `5` | both | Dial and per-operation timeout in seconds (1–60). |

### Bind modes

**Template mode** (`ldap.BindMode = template`) is the simplest: the username is
substituted into `BindDNTemplate` to form a DN, and the server binds directly as
that DN with the supplied password. Use it when every user's DN follows a fixed
pattern.

```text
BindMode       = template
BindDNTemplate = uid=%s,ou=people,dc=example,dc=com
# Active Directory alternative:
# BindDNTemplate = %s@example.com
```

**Search mode** (`ldap.BindMode = search`) handles directories where the DN is
not predictable. The server first binds as a read-only **service account**
(`SearchBindDN` / `SearchBindPassword`), searches under `BaseDN` with
`UserFilter` to locate the user's DN, and then **re-binds as the user** with the
supplied password to verify it.

```text
BindMode           = search
ServerURL          = ldaps://dc.example.com:636
BaseDN             = dc=example,dc=com
UserFilter         = (sAMAccountName=%s)      # Active Directory
SearchBindDN       = cn=svc-radius,ou=svc,dc=example,dc=com
SearchBindPassword = ********
```

### Transport security

- Use an `ldaps://` URL for implicit TLS (port 636), **or** an `ldap://` URL
  together with `StartTLS = true` to upgrade the plaintext connection (RFC 4513
  §3). Do not enable StartTLS on an `ldaps://` URL.
- Leave `TLSSkipVerify = false` in production. Enable it only against a lab or
  self-signed directory — it disables certificate verification and exposes the
  bind to interception.

## How authentication flows

| Method | Behaviour when `ldap.Enabled = true` |
| --- | --- |
| Bare PAP | Authenticated by LDAP bind. |
| EAP-TTLS / inner PAP | Authenticated by LDAP bind; MS-MPPE keys are still derived for the tunnel. |
| CHAP / MS-CHAP / MS-CHAPv2 | **Rejected** with a diagnostic reason (no cleartext password to bind with). |
| EAP-MD5 / PEAP-MSCHAPv2 | **Rejected** for the same reason. |
| MAC authentication | Bypasses LDAP entirely. |

The rejection of challenge/response methods is deliberate and centralized at the
password-retrieval boundary, not forked at the protocol ingress. When LDAP is
active the local `RadiusUser.Password` is usually empty, and a challenge/response
method that derived its expected value from an empty secret could otherwise
**falsely accept any user** — so those methods are failed closed instead.

## Security model

The backend is conservative by design:

- **Empty username or password is rejected before any network operation.** A
  bind with a DN and an empty password is treated as an *anonymous* bind by many
  servers and would otherwise succeed (RFC 4513 §5.1.2), so it is refused up
  front.
- **Injection is neutralized.** In search mode the username is escaped with LDAP
  filter escaping (RFC 4515 §3); in template mode it is DN-escaped before
  substitution.
- **An ambiguous search is refused.** If `UserFilter` matches more than one
  entry, the attempt is rejected rather than guessing.
- **A directory outage is never reported as a wrong password.** An invalid-
  credentials result (LDAP code 49) maps to a password rejection; every other
  failure — unreachable directory, misconfiguration, service-account bind
  failure — maps to a *backend-unavailable* outcome. This distinction also drives
  the metrics below.

## Observability

Rejections are counted by reason (see the [Operations Guide](./ops-guide.md) for
the metrics endpoint):

- `radus_reject_passwd_error` — the password was wrong (bind returned
  invalid-credentials).
- `radus_reject_ldap_error` — the backend could not give an answer (directory
  unreachable, StartTLS failure, service-account bind failure, misconfiguration).

Keeping these separate means a directory outage shows up as
`radus_reject_ldap_error`, not as a spike of "wrong password" — so an alert on
`radus_reject_ldap_error` cleanly signals a directory problem rather than user
error.

> **Note.** A `radus_accept` success counter is defined but is **not yet
> incremented** in the current release (its wiring is tracked as a follow-up).
> Until then, base alerting on the two reject counters above rather than on a
> computed success rate.

## Troubleshooting

| Symptom | Likely cause |
| --- | --- |
| Every login fails, `radus_reject_ldap_error` climbing | `ServerURL` wrong/unreachable, TLS handshake failing, or (search mode) the service account cannot bind. |
| Windows/AD clients rejected, PAP users fine | The clients are using PEAP-MSCHAPv2 or MS-CHAPv2, which LDAP cannot serve — switch them to EAP-TTLS/PAP or use the local password backend. |
| Search mode finds no user | Check `BaseDN` and `UserFilter`; confirm the service account can read user entries. |
| Bind works in the lab but fails in production | `TLSSkipVerify` was masking an invalid certificate — install a trusted certificate and turn it off. |
| Passwords accepted but the session has no plan/limits | The local `RadiusUser` row is missing — LDAP only checks the password; create the local account for authorization. |

## See also

- [Operations Guide](./ops-guide.md) — EAP/TLS certificates, metrics, process model.
- [Concepts & Terminology](./concepts.md) — PAP, EAP, EAP-TTLS.
- [Protocol & RFC Reference](./rfc-index.md) — RFC 5281 (EAP-TTLS), RFC 4511/4513 (LDAP).
