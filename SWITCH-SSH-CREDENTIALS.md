# Switch-Specific SSH Credentials

## Overview
Edgecore switches and other network devices often use different SSH credentials than servers and VMs. The dashboard now supports per-switch SSH authentication configuration.

## Configuration

### Switch-Specific Credentials
Add SSH credential fields to any switch in `config/config.yaml`:

```yaml
switches:
  - id: "sw001"
    name: "Core Switch 1"
    ip_address: "192.168.1.100"
    hostname: "coreswitch1.local"
    port: 22
    enabled: true
    controller_ip: "192.168.1.250"
    openflow_version: "1.3"
    # Edgecore switches often have different credentials
    ssh_username: "admin"                # Override global SSH username
    ssh_password: "edgecore"             # Switch-specific password
    # ssh_key_path: "~/.ssh/edgecore_rsa"  # Or use key-based auth
    
  - id: "sw002"
    name: "Access Switch 2"
    ip_address: "192.168.1.101"
    hostname: "accessswitch2.local"
    port: 22
    enabled: true
    controller_ip: "192.168.1.250"
    openflow_version: "1.3"
    ssh_username: "root"
    ssh_key_path: "~/.ssh/switch_key"
    # Omit ssh_password to use key-only authentication
```

### Credential Precedence
For each switch, the dashboard uses credentials in this order:

1. **Switch-specific credentials** (if defined):
   - `ssh_username` - Switch-specific username
   - `ssh_password` - Switch-specific password
   - `ssh_key_path` - Switch-specific private key path

2. **Global SSH credentials** (fallback):
   - From the main `ssh` section in config.yaml
   - Used if switch doesn't specify custom credentials

3. **Mixed credentials**:
   - You can specify only some fields (e.g., different username but same key)
   - Missing fields automatically fall back to global config

### Examples

#### Same credentials for all devices
```yaml
ssh:
  enabled: true
  username: "admin"
  password: "secretpass"
  
switches:
  - id: "sw001"
    # No ssh_* fields = uses global credentials
```

#### Different password per switch
```yaml
ssh:
  enabled: true
  username: "admin"
  private_key_path: "~/.ssh/id_rsa"
  
switches:
  - id: "sw001"
    ssh_password: "edgecore"  # Uses "admin" username from global, password override
  - id: "sw002"
    ssh_password: "netgear"   # Different password, same username
```

#### Completely separate credentials
```yaml
ssh:
  enabled: true
  username: "sysadmin"
  private_key_path: "~/.ssh/server_key"
  
switches:
  - id: "sw001"
    ssh_username: "admin"
    ssh_password: "edgecore"
    # Completely different from servers
```

## Implementation Details

### Code Changes

1. **internal/config/config.go**
   - Added `SSHUsername`, `SSHPassword`, `SSHKeyPath` fields to `SwitchConfig` struct

2. **internal/services/network.go**
   - Added `getSwitchSSHClient()` function
   - Creates per-switch SSH clients with custom credentials
   - Falls back to global client if no custom credentials specified
   - Updated `MonitorSwitch()` to use `getSwitchSSHClient()`

3. **config/config.yaml**
   - Added example switch configurations with SSH credentials

### How It Works

1. When monitoring a switch, `MonitorSwitch()` calls `getSwitchSSHClient()`
2. `getSwitchSSHClient()` looks up the switch config by ID
3. If custom SSH fields are found, creates a new SSH client with those credentials
4. If no custom fields, returns the global SSH client
5. Missing fields (e.g., username but no password) fall back to global values

### SSH Client Creation
The `NewSSHClient()` function supports:
- **Private key authentication** (preferred)
- **Password authentication** (fallback)
- **Both methods** (tries key first, then password)

### Security Considerations

- Store passwords in environment variables or secrets management
- Use SSH keys instead of passwords when possible
- Restrict file permissions on private keys (chmod 600)
- Consider using SSH agent for key management

## Testing

### Mock Mode
In development mode (`use_mock_data: true`), SSH credentials are not used. The dashboard generates realistic mock data for all switches.

### Production Mode
Set `use_mock_data: false` to enable real SSH monitoring:

```yaml
monitoring:
  use_mock_data: false
  
ssh:
  enabled: true
  username: "sysadmin"
  private_key_path: "~/.ssh/id_rsa"
  timeout_seconds: 10
  
switches:
  - id: "sw001"
    name: "Edgecore Switch"
    ip_address: "192.168.1.100"
    ssh_username: "admin"
    ssh_password: "edgecore"
```

### Verification
Check logs for SSH connection messages:
- Successful: "SSH monitoring enabled for production metrics"
- Switch-specific: Creates client per switch (no explicit log unless error)
- Fallback: "Failed to create SSH client for switch X, falling back to global client"

## Common Edgecore Defaults

| Model | Default Username | Default Password |
|-------|-----------------|------------------|
| AS4610 | admin | admin |
| AS5712 | admin | admin |
| AS7712 | admin | admin |

**Note**: Always change default passwords in production environments.

## Troubleshooting

### Switch shows "offline" status
- Verify IP address is correct and switch is reachable
- Check firewall rules allow SSH (port 22)
- Confirm SSH credentials are correct
- Test SSH manually: `ssh admin@192.168.1.100`

### "SSH monitoring failed" in logs
- Check username/password combination
- Verify SSH key path is correct and file exists
- Ensure switch has SSH enabled
- Check timeout settings (increase if network is slow)

### Mixing authentication methods
If both password and key are specified, the SSH client tries:
1. Private key authentication first
2. Password authentication as fallback

This is useful for environments with mixed authentication requirements.
