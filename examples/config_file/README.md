# Config File Example

This example demonstrates how to use the ConfigFileTyped interface for accessing configuration data with type-safe getters and object functionality.

## Overview

The library provides two levels of configuration access:
1. **ConfigFileSource** - Lightweight interface that only provides basic value access (`GetValue`, `SetValue`, etc.)
2. **ConfigFileTyped** - Wrapper that adds type-safe getters and object access methods

## Key Innovation: Objects are ConfigFileTyped too!

When you use `GetObject()` or `GetObjectSlice()`, the returned objects are also `ConfigFileTyped` instances. This means you have full access to:
- All typed getters (`GetString`, `GetInt`, `GetBool`, etc.)
- All typed setters (`SetString`, `SetInt`, `SetBool`, etc.)
- Nested object access (objects can contain other objects)
- Slice getters and setters

### How it works under the hood:
- `mapConfigSource` provides a minimal `ConfigFileSource` implementation for in-memory maps
- `ConfigFileTypedWrapper` wraps this source to provide all the typed getters and setters
- This eliminates code duplication while providing full functionality for nested objects

## Features Demonstrated

1. **Simple Typed Values**: Using `GetInt()`, `GetBool()`, `GetUint64()`, etc. to retrieve configuration values with proper type safety.

2. **Array of Strings**: Using `GetStringSlice()` to retrieve string arrays from configuration.

3. **Array of Objects**: Using `GetObjectSlice()` to retrieve arrays of objects where each object is a full `ConfigFileTyped` instance:
   ```go
   endpoints := config.GetObjectSlice("endpoints")
   for _, endpoint := range endpoints {
       name := endpoint.GetString("name")
       timeout := endpoint.GetInt("timeout")
       authRequired := endpoint.GetBool("auth_required")
       rateLimit := endpoint.GetInt64("rate_limit")

       // Objects support all ConfigFileTyped methods
       protocols := endpoint.GetStringSlice("protocols")

       // Objects can be nested!
       if nested := endpoint.GetObject("nested_config"); nested != nil {
           // Access nested properties
       }
   }
   ```

4. **Nested Objects**: Using `GetObject()` to retrieve nested configuration objects. The returned object is also a `ConfigFileTyped` instance:
   ```go
   service := config.GetObject("service")
   if service != nil {
       // Use typed getters directly on the nested object
       name := service.GetString("name")
       version := service.GetString("version")

       // Access deeper nesting
       dbConfig := service.GetObject("database")
       if dbConfig != nil {
           host := dbConfig.GetString("host")
           port := dbConfig.GetInt("port")
       }
   }
   ```

5. **Error Handling**: Demonstrates graceful handling of missing or invalid configuration values.

6. **Type Flexibility**: Objects automatically handle type conversions from TOML/JSON, including `int64` to `int`, `float64` to `float32`, etc.

7. **Dynamic Object Creation**: Create, modify, and save configuration objects at runtime:
   ```go
   // Create a new object
   newEndpoint := cli.NewTypedConfigObject()
   newEndpoint.SetString("name", "new-service")
   newEndpoint.SetInt("port", 8080)
   newEndpoint.SetBool("enabled", true)

   // Set nested objects
   dbConfig := cli.NewTypedConfigObject()
   dbConfig.SetString("host", "localhost")
   dbConfig.SetInt("port", 5432)
   newEndpoint.SetObject("database", dbConfig)

   // Add to existing array
   endpoints := config.GetObjectSlice("endpoints")
   endpoints = append(endpoints, newEndpoint)
   config.SetObjectSlice("endpoints", endpoints)

   // Or set a single object
   config.SetObject("new_service", newEndpoint)
   ```

## Running the Example

```bash
cd examples/config_file
go build -o config-example main.go
./config-example
```

## Configuration File Structure

The `test.toml` file shows various configuration patterns:

- Simple key-value pairs (`timeout`, `debug`)
- Arrays of strings (`api_keys`)
- Arrays of objects (`endpoints`, `servers`, `features`)
- Nested objects with their own properties (`service.database`, `service.cache`)

## Type Handling

The ConfigFileTyped getters automatically handle type conversions from TOML's native types:
- TOML integers are parsed as `int64` but can be accessed via `GetInt()` or `GetInt64()`
- Numeric types can be converted between int, int64, uint, etc.
- `map[string]interface{}` from TOML is automatically converted for the object getters

## Integration

To use these features in your own code:

```go
import (
    "github.com/paularlott/cli"
    "github.com/paularlott/cli/toml"
)

// Create the basic config
baseConfig := toml.NewConfigFile(&configFileName, searchPathFunc)

// Wrap it with the typed wrapper
config := cli.NewTypedConfigFile(baseConfig)

// Load it
if err := config.LoadData(); err != nil {
    log.Fatal(err)
}

// Use typed getters
timeout := config.GetInt("timeout")
debug := config.GetBool("debug")

// Use object slices for complex data
endpoints := config.GetObjectSlice("endpoints")
for _, e := range endpoints {
    // Each endpoint is a full ConfigFileTyped!
    if e.GetBool("auth_required") {
        fmt.Println("Auth required for:", e.GetString("name"))

        // Access arrays within objects
        protocols := e.GetStringSlice("protocols")

        // Objects can even be nested within objects
        if nested := e.GetObject("nested_config"); nested != nil {
            // Full typed access on nested objects
        }
    }
}

// Access nested objects - they're also ConfigFileTyped!
service := config.GetObject("service")
if service != nil {
    dbName := service.GetString("database.name")  // Access nested using dot notation
    port := service.GetInt("database.port")

    // Or get the nested object directly
    dbConfig := service.GetObject("database")
    if dbConfig != nil {
        host := dbConfig.GetString("host")  // Relative path within the object
        poolSize := dbConfig.GetInt("pool_size")
    }
}

// Create and set new objects dynamically
newEndpoint := cli.NewTypedConfigObject()
newEndpoint.SetString("name", "new-service")
newEndpoint.SetInt("port", 8080)
newEndpoint.SetBool("enabled", true)

// Add to existing array
endpoints := config.GetObjectSlice("endpoints")
endpoints = append(endpoints, newEndpoint)
config.SetObjectSlice("endpoints", endpoints)

// Or set a single object
config.SetObject("new_service", newEndpoint)

// Objects can also be nested when creating
dbConfig := cli.NewTypedConfigObject()
dbConfig.SetString("host", "localhost")
dbConfig.SetInt("port", 5432)
newEndpoint.SetObject("database", dbConfig)
```

## Key Points

- Always use `NewTypedConfigFile()` to wrap your base config source if you need typed access
- **Objects returned by `GetObject()` and `GetObjectSlice()` are full `ConfigFileTyped` instances**
- This gives you complete access to all getters, setters, and even nested object access
- Objects from files are read-only wrappers (setters will return errors)
- **Dynamically created objects** (via `NewTypedConfigObject()`) support both getters and setters
- Use `SetObject()` and `SetObjectSlice()` to modify configuration at runtime
- Type conversion is handled automatically - TOML's `int64` values can be read as `int` if preferred
- The approach supports unlimited nesting depth - objects can contain other objects which contain more objects