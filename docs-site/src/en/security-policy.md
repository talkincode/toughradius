# Security Policy

> 中文版本：[安全策略](../zh/security-policy.md)

This chapter is the canonical home for ToughRADIUS security advisories and the
guidance that goes with them. The repository's
[`SECURITY.md`](https://github.com/talkincode/toughradius/blob/main/SECURITY.md)
keeps a short pointer back to this chapter so there is a single source of truth.

## Security advisories

### Default super-admin credentials (GHSA-2gwm-6gf5-8699)

ToughRADIUS no longer creates or accepts the historical first-start password.
A fresh install generates a one-time bootstrap password (or uses
`TOUGHRADIUS_ADMIN_PASSWORD`) and writes it to
`{workdir}/private/admin-bootstrap-password`. Upgrades that still store the
historical default rotate it on startup. Login and operator-password APIs reject
that credential.

| Item               | Details                                      |
| ------------------ | -------------------------------------------- |
| Vulnerability type | Use of Default Credentials (CWE-1392)        |
| Severity           | Critical                                     |
| Advisory           | GHSA-2gwm-6gf5-8699                          |
| Affected component | Super-admin bootstrap and `/api/v1/auth/login` |

#### Recommended actions

Upgrade, then confirm every operator uses a unique password. If you cannot read
`{workdir}/private/admin-bootstrap-password`, reset it with `cmd/reset-password`.

### XSS vulnerability fix (v8.0.8)

Version **v8.0.8** addresses a critical cross-site scripting (XSS) vulnerability.
The issue was found in the `errmsg` parameter handling in the login endpoint.

| Item               | Details                             |
| ------------------ | ----------------------------------- |
| Vulnerability type | Cross-Site Scripting (XSS)          |
| Severity           | Critical                            |
| Affected versions  | v8.0.1 – v8.0.7                     |
| Fixed version      | v8.0.8                              |
| Affected component | Login endpoint (`errmsg` parameter) |

#### Recommended actions

We strongly recommend that all users update to the latest version immediately.
See the [Documentation Map](./documentation-map.md) for the README and build
instructions you can follow to upgrade your deployment.
