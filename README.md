# Let's-a-go - Vehicle State Monitor

A lightweight Go tool that monitors the vehicle state in Redis and updates the dashboard when the vehicle is ready to drive.

## Features

- Connects to Redis at 192.168.7.1:6379
- Monitors the `state` field in the `vehicle` hash for state transitions
- When state changes from `stand-by` to `parked`, it:
  - Sets `ready` to `true` in the `dashboard` hash
  - Publishes a `ready` message to the `dashboard` channel
- Lightweight and efficient for embedded systems
- Optimized for ARMv7l architecture

## Building

### Prerequisites

- Go 1.15+
- Make

### Build for Local Testing

```bash
make
```

### Build for ARMv7l Target

```bash
make dist-arm
```
This creates an optimized and stripped binary called `letsago-arm`.

## Installation

### Automatic Installation (on target device)

1. Build the ARM binary:
   ```bash
   make dist-arm
   ```

2. Copy the entire directory to the target device

3. Run the install script with root privileges:
   ```bash
   sudo ./install.sh
   ```

4. Enable and start the service:
   ```bash
   sudo systemctl enable letsago
   sudo systemctl start letsago
   ```

### Manual Installation

1. Build the ARM binary:
   ```bash
   make dist-arm
   ```

2. Copy the binary to the target:
   ```bash
   scp letsago-arm user@target:/usr/bin/letsago
   ```

3. Set executable permissions:
   ```bash
   ssh user@target "chmod +x /usr/bin/letsago"
   ```

4. Copy the service file:
   ```bash
   scp letsago.service user@target:/etc/systemd/system/
   ```

5. Enable and start the service:
   ```bash
   ssh user@target "systemctl daemon-reload && systemctl enable --now letsago"
   ```

## Usage

Once installed and running, the service will automatically:
- Connect to Redis at 192.168.7.1:6379
- Monitor the vehicle state transitions
- Update dashboard when state transitions from `stand-by` to `parked`

### Checking Status

```bash
systemctl status letsago
```

### Viewing Logs

```bash
journalctl -u letsago
```

## Development

- Main logic is in `main.go`
- Modify Redis connection details or behavior in the constants section
- Use `make` to build for local testing
