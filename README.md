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

## 🔧 Usage Examples

### Basic Gin Application
```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/Steve-Lee-CST/go-gin-student-tool/router"
)

func main() {
    engine := gin.Default()
    router.RegisterRouter(engine)
    engine.Run(":80")
}
```

### Using Coherency Cache
```go
// Create multi-level cache
L3Storage := NewMemoryStorage[string, string]()
L2Cache := NewCoherencyStorage(
    NewMemoryStorage[string, string](),
    L3Storage,
)
L1Cache := NewCoherencyStorage(
    NewMemoryStorage[string, string](),
    L2Cache,
)

// Set and get values
key := "test_key"
value := "test_value"

err := L1Cache.Set(context.Background(), &key, &value, time.Millisecond*100)
if err != nil {
    log.Fatal(err)
}

getValue, err := L1Cache.Get(context.Background(), &key, time.Second*10)
if err != nil {
    log.Fatal(err)
}
```

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./...

# Run specific test with verbose output
go test -v ./micro_service_tool/coherency_cache/

# Run benchmarks
go test -bench=. ./micro_service_tool/coherency_cache/
```

## 📋 API Endpoints

- `GET /ping` - Health check endpoint
- `GET /request-id` - Get current request ID
- Additional endpoints can be added through the router system

## 🔒 Dependencies

- **Gin**: Web framework
- **Go Redis**: Redis client (optional)
- **UUID**: Unique identifier generation
- **Rate Limiter**: Request rate limiting

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🎯 Use Cases

- **Student Projects**: Quick setup for web applications
- **Microservices**: Cache consistency and middleware components
- **API Development**: Rate limiting and logging utilities
- **Learning**: Reference implementation for Go/Gin patterns

## 🔍 Key Features

- **Thread-safe**: All operations are designed for concurrent access
- **Configurable**: Flexible timeout and configuration options
- **Tested**: Comprehensive test coverage
- **Extensible**: Easy to extend with custom implementations
- **Production-ready**: Designed for real-world applications