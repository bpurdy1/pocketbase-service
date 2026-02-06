# PocketBase Server

A custom Go server wrapping PocketBase with extensible service architecture.

## Project Structure

```
.
├── main.go              # Entry point
├── server/
│   └── server.go        # Custom server struct wrapping PocketBase
└── services/
    └── health.go        # Example health check service
```

## Running the Server

```bash
go run . serve --http=0.0.0.0:8080
```

## Endpoints

- PocketBase Admin UI: `http://localhost:8080/_/`
- PocketBase API: `http://localhost:8080/api/*`
- Custom Health Check: `http://localhost:8080/api/custom/health`

## Adding New Services

Implement the `Service` interface:

```go
type Service interface {
    Name() string
    RegisterRoutes(mux *http.ServeMux)
}
```

Register in `main.go`:

```go
srv.RegisterService(services.NewMyService())
```

---

# Cloudflare Tunnel Setup for GL.iNet Flint 2 (GL-MT6000)

## Prerequisites

- SSH access to your Flint 2 router
- Cloudflare account (optional for quick tunnels)
- Domain connected to Cloudflare (for custom domains)

## Method 1: Using OpenWrt Packages (Recommended)

SSH into your router and run:

```bash
opkg update
opkg install cloudflared luci-app-cloudflared
```

Then configure via LuCI web interface at **Services → Cloudflare Tunnel**.

## Method 2: One-Command Script

### Quick Tunnel (No Cloudflare Account Needed)

```bash
wget -qO- https://raw.githubusercontent.com/adshrc/openwrt-cloudflared/main/script.sh | ash -s -- --url=http://192.168.8.1:80
```

This gives you a `*.trycloudflare.com` URL instantly.

### With Your Own Domain

```bash
wget -qO- https://raw.githubusercontent.com/adshrc/openwrt-cloudflared/main/script.sh | ash -s -- -l
```

Follow the prompts to log into your Cloudflare account.

## Method 3: Manual Setup

```bash
# Download cloudflared for aarch64
cd /tmp
wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64
chmod +x cloudflared-linux-arm64
mv cloudflared-linux-arm64 /usr/bin/cloudflared

# Authenticate with Cloudflare
cloudflared tunnel login

# Create a new tunnel
cloudflared tunnel create flint2

# Run the tunnel
cloudflared tunnel run flint2
```

## Exposing PocketBase Through Cloudflare Tunnel

Once your tunnel is running, configure it to expose your PocketBase server:

```bash
cloudflared tunnel route dns flint2 pocketbase.yourdomain.com
```

Create a config file at `/etc/cloudflared/config.yml`:

```yaml
tunnel: flint2
credentials-file: /etc/cloudflared/<tunnel-id>.json

ingress:
  - hostname: pocketbase.yourdomain.com
    service: http://localhost:8080
  - service: http_status:404
```

## Known Issue: Auto-Start on MT6000

Cloudflared may not auto-start on boot. Workaround:

Edit `/etc/rc.local` and add before `exit 0`:

```bash
sleep 30 && /etc/init.d/cloudflared restart
```

Or create a cron job:

```bash
crontab -e
```

Add:

```
@reboot sleep 60 && /etc/init.d/cloudflared restart
```

## Troubleshooting

### Check Tunnel Status

```bash
cloudflared tunnel list
cloudflared tunnel info flint2
```

### View Logs

```bash
logread | grep cloudflared
```

### Restart Service

```bash
/etc/init.d/cloudflared restart
```

## Resources

- [GL.iNet Forum - Cloudflared on GL-iNet routers](https://forum.gl-inet.com/t/cloudflared-remotely-managed-tunnel-on-gl-inet-routers/47733)
- [GitHub - openwrt-cloudflared](https://github.com/adshrc/openwrt-cloudflared)
- [Cloudflare Community - OpenWrt support](https://community.cloudflare.com/t/openwrt-support/610306)
- [Cloudflare Tunnel Documentation](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)
