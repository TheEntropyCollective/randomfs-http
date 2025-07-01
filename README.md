# RandomFS HTTP Server

A high-performance HTTP server for the Owner Free File System, providing REST API access and web interface for storing and retrieving files using randomized blocks on IPFS.

## Overview

RandomFS HTTP Server is a production-ready HTTP server that exposes RandomFS functionality through a REST API and serves a modern web interface. It's built with Gorilla Mux for routing and provides comprehensive file management capabilities.

## Features

- **REST API**: Full HTTP API for programmatic access
- **Web Interface**: Modern, responsive web UI for file management
- **CORS Support**: Cross-origin resource sharing enabled
- **Health Endpoints**: System health and status monitoring
- **File Upload/Download**: Drag-and-drop file operations
- **rd:// URL Generation**: Automatic URL creation for file sharing
- **Statistics API**: Real-time system metrics
- **Configurable**: Customizable ports, directories, and cache settings

## Installation

### From Source
```bash
git clone https://github.com/TheEntropyCollective/randomfs-http
cd randomfs-http
go build -o randomfs-http
```

### Binary Download
Download the latest release for your platform from the [releases page](https://github.com/TheEntropyCollective/randomfs-http/releases).

## Quick Start

```bash
# Start server with default settings
./randomfs-http

# Start on custom port
./randomfs-http -port 9000

# Start with custom IPFS endpoint
./randomfs-http -ipfs http://192.168.1.100:5001

# Start with custom web interface directory
./randomfs-http -web ./custom-web
```

## Configuration

### Command Line Flags
- `-port`: HTTP server port (default: 8080)
- `-ipfs`: IPFS API endpoint (default: http://localhost:5001)
- `-data`: Data directory (default: ./data)
- `-cache`: Cache size in bytes (default: 500MB)
- `-web`: Web interface directory (default: ./web)

### Environment Variables
- `RANDOMFS_PORT`: HTTP server port
- `RANDOMFS_IPFS_API`: IPFS API endpoint
- `RANDOMFS_DATA_DIR`: Data directory
- `RANDOMFS_CACHE_SIZE`: Cache size in bytes
- `RANDOMFS_WEB_DIR`: Web interface directory

## API Reference

### Base URL
```
http://localhost:8080/api/v1
```

### Endpoints

#### POST /store
Store a file in RandomFS.

**Request:**
- Method: `POST`
- Content-Type: `multipart/form-data`
- Body: File upload

**Response:**
```json
{
  "success": true,
  "rd_url": "rd://QmX...abc/text/plain/example.txt",
  "filename": "example.txt",
  "content_type": "text/plain",
  "size": 1024
}
```

#### GET /retrieve/{hash}
Retrieve a file by its representation hash.

**Request:**
- Method: `GET`
- Path: `/retrieve/{hash}`

**Response:**
- File content with appropriate Content-Type header

#### GET /download/{hash}
Download a file by its representation hash.

**Request:**
- Method: `GET`
- Path: `/download/{hash}`

**Response:**
- File content with Content-Disposition header for download

#### GET /stats
Get system statistics.

**Request:**
- Method: `GET`

**Response:**
```json
{
  "files_stored": 42,
  "blocks_generated": 156,
  "total_size": 1048576,
  "cache_hits": 89,
  "cache_misses": 12
}
```

#### GET /health
Health check endpoint.

**Request:**
- Method: `GET`

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "1.0.0"
}
```

#### GET /parse/{rd_url}
Parse a rd:// URL and return its components.

**Request:**
- Method: `GET`
- Path: `/parse/{rd_url}`

**Response:**
```json
{
  "hash": "QmX...abc",
  "content_type": "text/plain",
  "filename": "example.txt",
  "valid": true
}
```

## Web Interface

The server includes a modern web interface accessible at the root URL:

```
http://localhost:8080
```

### Features
- **Drag & Drop Upload**: Simply drag files to upload
- **File Management**: View stored files and their rd:// URLs
- **Download Links**: Direct download links for files
- **URL Sharing**: Copy rd:// URLs for sharing
- **Real-time Stats**: Live system statistics
- **Responsive Design**: Works on desktop and mobile

## Examples

### Using curl

```bash
# Store a file
curl -X POST -F "file=@example.txt" http://localhost:8080/api/v1/store

# Retrieve a file
curl http://localhost:8080/api/v1/retrieve/QmX...abc

# Get statistics
curl http://localhost:8080/api/v1/stats

# Parse a rd:// URL
curl http://localhost:8080/api/v1/parse/rd://QmX...abc/text/plain/example.txt
```

### Using JavaScript

```javascript
// Store a file
const formData = new FormData();
formData.append('file', fileInput.files[0]);

fetch('/api/v1/store', {
  method: 'POST',
  body: formData
})
.then(response => response.json())
.then(data => {
  console.log('File stored:', data.rd_url);
});

// Get statistics
fetch('/api/v1/stats')
.then(response => response.json())
.then(data => {
  console.log('Files stored:', data.files_stored);
});
```

## Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o randomfs-http

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/randomfs-http .
EXPOSE 8080
CMD ["./randomfs-http"]
```

## Development

```bash
# Clone repository
git clone https://github.com/TheEntropyCollective/randomfs-http
cd randomfs-http

# Install dependencies
go mod tidy

# Build
go build -o randomfs-http

# Run tests
go test -v

# Run in development mode
./randomfs-http -port 8080 -web ./web
```

## Production Deployment

### Systemd Service
```ini
[Unit]
Description=RandomFS HTTP Server
After=network.target

[Service]
Type=simple
User=randomfs
WorkingDirectory=/opt/randomfs-http
ExecStart=/opt/randomfs-http/randomfs-http
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Nginx Reverse Proxy
```nginx
server {
    listen 80;
    server_name randomfs.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Dependencies

- Go 1.21+
- [randomfs-core](https://github.com/TheEntropyCollective/randomfs-core) library
- IPFS node (Kubo) with HTTP API enabled

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Related Projects

- [randomfs-core](https://github.com/TheEntropyCollective/randomfs-core) - Core library
- [randomfs-cli](https://github.com/TheEntropyCollective/randomfs-cli) - Command-line interface
- [randomfs-web](https://github.com/TheEntropyCollective/randomfs-web) - Web interface 