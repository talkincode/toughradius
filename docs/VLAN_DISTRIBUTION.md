# VLAN Distribution for ToughRadius

This document explains how to configure VLAN distribution for network switches using ToughRadius.

## Overview

ToughRadius now supports VLAN assignment through RADIUS authentication responses using RFC 2868 tunnel attributes. When a user successfully authenticates, the system can automatically assign them to a specific VLAN on the network switch.

## Configuration

### 1. User VLAN Configuration

In the ToughRadius web interface, navigate to **RADIUS Users** management:

1. **Create/Edit User**: When creating or editing a user, you'll see two new fields:
   - **VLAN ID 1**: Primary VLAN ID for assignment
   - **VLAN ID 2**: Secondary VLAN ID (fallback)

2. **VLAN Assignment Logic**:
   - If VLAN ID 1 is configured (non-zero), it will be used for assignment
   - If VLAN ID 1 is empty and VLAN ID 2 is configured, VLAN ID 2 will be used
   - If both are empty, no VLAN assignment occurs

### 2. Switch Configuration

For H3C switches and other RFC 2868 compliant switches, ensure they are configured to:

1. Accept RADIUS tunnel attributes
2. Process VLAN assignments from RADIUS responses
3. Support dynamic VLAN assignment

### 3. RADIUS Attributes Sent

When a user with configured VLAN ID authenticates successfully, ToughRadius sends these tunnel attributes:

```
Tunnel-Type = VLAN (13)
Tunnel-Medium-Type = IEEE-802 (6) 
Tunnel-Private-Group-ID = <VLAN_ID>
```

## Example Configuration

**User Configuration:**
- Username: `user123`
- VLAN ID 1: `100`
- VLAN ID 2: `200`

**Result:** User will be assigned to VLAN 100 upon authentication.

## Troubleshooting

1. **VLAN assignment not working:**
   - Verify the switch supports RFC 2868 tunnel attributes
   - Check that VLAN exists on the switch
   - Ensure RADIUS authentication is successful first

2. **Check RADIUS logs:**
   - Monitor ToughRadius logs for successful authentication
   - Verify tunnel attributes are being sent in Access-Accept packets

3. **Switch debugging:**
   - Enable RADIUS debugging on the switch
   - Check if tunnel attributes are being received and processed

## Compatibility

This VLAN distribution feature has been tested with:
- H3C switches
- Other RFC 2868 compliant network devices

The implementation uses standard RFC 2868 tunnel attributes, making it compatible with most modern network switches that support RADIUS-based VLAN assignment.