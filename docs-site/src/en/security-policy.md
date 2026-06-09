# Security Policy

> 中文版本：[安全策略](../zh/security-policy.md)

This chapter is the canonical home for ToughRADIUS security advisories and the
guidance that goes with them. The repository's
[`SECURITY.md`](https://github.com/talkincode/toughradius/blob/main/SECURITY.md)
keeps a short pointer back to this chapter so there is a single source of truth.

## Security advisories

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
