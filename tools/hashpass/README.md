# Password Hash Utility

This utility generates bcrypt password hashes for the Server Dashboard.

## Installation

```bash
cd tools/hashpass
go build -o hashpass
```

## Usage

### 1. Generate Hash (Display Only)

```bash
./hashpass
```

This will prompt for a password and display the bcrypt hash.

### 2. Update config.yaml

```bash
./hashpass --update
```

This will:
- Prompt for a password
- Generate a bcrypt hash
- Update the `config/config.yaml` file with the hash

### 3. Generate Environment Variable Format

```bash
./hashpass --env
```

This will output the hash in a format ready to export as an environment variable.

### 4. Specify Custom Config Path

```bash
./hashpass --config /path/to/config.yaml --update
```

## Examples

**Interactive hash generation:**
```bash
$ ./hashpass
Enter password to hash: ****
Confirm password: ****

✓ Password hashed successfully!

Bcrypt hash:
$2a$12$xVqYp5Z8nN7Kj3Qw9...

You can:
  1. Use --update to save to config.yaml
  2. Use --env for environment variable format
  3. Manually set: export AUTH_PASSWORD='$2a$12$xVq...'
```

**Update config file:**
```bash
$ ./hashpass --update
Enter password to hash: ****
Confirm password: ****

✓ Config file updated: config/config.yaml

Password hash has been saved to config.yaml
The application will now use bcrypt authentication.
```

**Generate for environment variable:**
```bash
$ ./hashpass --env
Enter password to hash: ****
Confirm password: ****

✓ Password hashed successfully!

Add this to your environment:
export AUTH_PASSWORD='$2a$12$xVqYp5Z8nN7Kj3Qw9...'

Or add to your shell profile (~/.bashrc, ~/.zshrc, etc.):
export AUTH_PASSWORD='$2a$12$xVqYp5Z8nN7Kj3Qw9...'
```

## Security Notes

- Bcrypt is a one-way hash - the original password cannot be recovered
- The hash can be safely stored in config files (though environment variables are still preferred)
- Cost factor is set to 12 (good balance of security and performance)
- Always use strong, unique passwords
- Never commit production passwords to version control
