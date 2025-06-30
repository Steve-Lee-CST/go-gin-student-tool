# Go Gin Student Tool

> Some simple tools

## Features

### Web Framework Tools (`gin_tool/`)
- **Request ID Middleware**: Automatic request ID generation and tracking
- **HTTP Logger**: Comprehensive request/response logging
- **Rate Limiting**: Configurable rate limiting middleware
- **HTTP Helper**: Utility functions for HTTP operations
- **Common Utilities**: Shared utility functions

### Microservice Tools (`micro_service_tool/`)
- **Coherency Cache**: Multi-level cache consistency middleware
  - Support for nested cache layers
  - Configurable timeout and context support
  - Thread-safe operations with lock mechanisms
  - Memory storage implementation
- **Block Call with Timeout**: Timeout management utilities

## Project Structure

```
go-gin-student-tool/
├── gin_tool/                 # Web framework utilities
│   ├── common.go            # Common utility functions
│   ├── http_helper.go       # HTTP operation helpers
│   ├── http_logger.go       # HTTP logging middleware
│   ├── rate_limit.go        # Rate limiting middleware
│   ├── reuqest_id.go        # Request ID middleware
│   └── util.go              # General utilities
├── micro_service_tool/      # Microservice components
│   └── coherency_cache/     # Cache consistency system
│       ├── coherency_cache.go      # Main cache implementation
│       ├── memory_storage.go       # Memory storage backend
│       └── coherency_storage_test.go # Comprehensive tests
│   └── block_call_with_timeout
│       ├── block_call_with_timeout.go  # Block Call Func With Timeout
├── handler/                 # HTTP handlers
├── middleware/              # Custom middleware
├── router/                  # Route definitions
├── config/   
├── script/  
└── main.go 
```
