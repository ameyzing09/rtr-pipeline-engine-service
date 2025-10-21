# Hot Reload Setup with Air

Air is a live-reload utility for Go applications, similar to nodemon for Node.js. It automatically rebuilds and restarts your server when files change.

## Installation

### Option 1: Using Go Install (Recommended)
```bash
go install github.com/air-verse/air@latest
```

Make sure `$GOPATH/bin` is in your PATH. You can check with:
```bash
go env GOPATH
```

### Option 2: Using Binary Download
Download the latest binary from: https://github.com/air-verse/air/releases

## Usage

### Start the server with hot reload:
```bash
air
```

That's it! Air will:
- Watch for file changes in `.go` files
- Automatically rebuild the project
- Restart the server
- Display build errors in the console

### Configuration

The configuration is in `.air.toml`. Key settings:

- **Build command**: `go build -o ./tmp/main.exe ./cmd/server`
- **Watch files**: `.go`, `.tpl`, `.tmpl`, `.html` files
- **Excluded dirs**: `tmp`, `vendor`, `bin`, `node_modules`
- **Excluded files**: `*_test.go` (test files)

### Tips

1. Air will create a `tmp/` directory for temporary builds (this is gitignored)
2. Build errors are logged to `tmp/build-errors.log`
3. Press `Ctrl+C` to stop Air and the server

### Comparison with nodemon

| Feature | nodemon | Air |
|---------|---------|-----|
| Watch files | ✅ | ✅ |
| Auto restart | ✅ | ✅ |
| Config file | `nodemon.json` | `.air.toml` |
| Command | `nodemon` | `air` |
