# cDNS

A powerful DNS query tool with support for multiple nameservers, background processing, and API endpoints.


## Build & Run

To build the project:

```
make build
```

To run the CLI:

```
make run
```

Or directly:

```
go run cmd/cdns/main.go
```

## Usage

- Query DNS records:
  ```
  cdns query example.com 8.8.8.8 1.1.1.1
  ```
- Start the API server:
  ```
  cdns api
  ```
- Show version:
  ```
  cdns version
  ```
- List popular DNS servers:
  ```
  cdns dns-list
  ```

## API Endpoints

- `GET /api/v1/health` - Health check
- `GET /api/v1/dns-servers` - List DNS servers
- `POST /api/v1/query` - Query DNS records
- `POST /api/v1/query/background` - Start background DNS query
- `GET /api/v1/task/:id` - Get background task status
- `GET /api/v1/tasks` - List background tasks

---
