# Security Policy

## Security Advisories

### XSS Vulnerability Fix (v8.0.8)

We have released a new version (v8.0.8) that addresses a critical security vulnerability related to cross-site scripting (XSS). The issue was found in the `errmsg` parameter handling in the login endpoint.

| Item                   | Details                             |
| ---------------------- | ----------------------------------- |
| **Vulnerability Type** | Cross-Site Scripting (XSS)          |
| **Severity**           | Critical                            |
| **Affected Versions**  | v8.0.1 ~ v8.0.7                     |
| **Fixed Version**      | v8.0.8                              |
| **Affected Component** | Login endpoint (`errmsg` parameter) |

#### Recommended Actions

We strongly recommend all users to update to the latest version immediately. You can update your project by following the instructions in our documentation.
