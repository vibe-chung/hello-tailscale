# hello-tailscale
A proof of concept webserver with embedded tailscale

## Overview

This is a simple "Hello World" webserver with an embedded [Tailscale](https://tailscale.com/) client. This allows you to deploy the webserver anywhere and access it securely via your Tailscale network (tailnet).

## Features

- Simple, clean "Hello World" web page
- Health check endpoint at `/health`
- Embedded Tailscale client for secure access
- Configurable hostname and port
- Graceful shutdown handling

## Prerequisites

- Go 1.25 or later
- A Tailscale account and tailnet

## Building

```bash
go build -o hello-tailscale
```

## Usage

```bash
./hello-tailscale [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-hostname` | `hello-tailscale` | Tailscale hostname for this service |
| `-port` | `80` | Port to listen on |
| `-state-dir` | OS default | Directory to store Tailscale state |

### Examples

Run with default settings:
```bash
./hello-tailscale
```

Run with custom hostname:
```bash
./hello-tailscale -hostname my-webserver
```

Run on a different port:
```bash
./hello-tailscale -port 8080
```

### First Run

On the first run, the application will print a URL to authenticate with Tailscale. Open this URL in your browser to authorize the device on your tailnet.

Once authenticated, you can access the webserver at:
- `http://<hostname>.<tailnet-name>.ts.net/` (if you have MagicDNS enabled)
- `http://<tailscale-ip>/`

## Endpoints

| Endpoint | Description |
|----------|-------------|
| `/` | Hello World page |
| `/health` | Health check endpoint (returns JSON) |

## Running Tests

```bash
go test -v
```

## License

MIT
