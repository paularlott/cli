package main

import (
	"fmt"
	"log"

	"github.com/paularlott/cli"
	cli_toml "github.com/paularlott/cli/toml"
)

func main() {
	// Create a TOML config file source pointing to test.toml
	configFileName := "test.toml"

	// Define search paths (current directory)
	searchPathFunc := func() []string {
		return []string{"."}
	}

	// Create the basic config
	baseConfig := cli_toml.NewConfigFile(&configFileName, searchPathFunc)

	// Wrap it with the typed wrapper
	config := cli.NewTypedConfigFile(baseConfig)

	// Load the configuration
	if err := config.LoadData(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Using config file: %s\n\n", config.FileUsed())

	// Example 1: Simple typed values
	fmt.Println("=== Simple Configuration Values ===")
	timeout := config.GetInt("timeout")
	debug := config.GetBool("debug")
	retryCount := config.GetInt64("retry_count")
	maxConnections := config.GetUint64("max_connections")

	fmt.Printf("Timeout: %d seconds\n", timeout)
	fmt.Printf("Debug mode: %t\n", debug)
	fmt.Printf("Retry count: %d\n", int(retryCount))
	fmt.Printf("Max connections: %d\n", maxConnections)

	// Example 2: Array of strings
	fmt.Println("\n=== API Keys ===")
	apiKeys := config.GetStringSlice("api_keys")
	if apiKeys != nil {
		fmt.Printf("Found %d API keys:\n", len(apiKeys))
		for i, key := range apiKeys {
			fmt.Printf("  %d: %s\n", i+1, maskApiKey(key))
		}
	}

	// Example 3: Array of objects (API endpoints)
	fmt.Println("\n=== API Endpoints Configuration ===")
	endpoints := config.GetObjectSlice("endpoints")
	if endpoints != nil {
		fmt.Printf("Found %d endpoints:\n", len(endpoints))
		for i, endpoint := range endpoints {
			name := endpoint.GetString("name")
			url := endpoint.GetString("url")
			timeout := endpoint.GetInt("timeout")
			retries := endpoint.GetInt("retries")
			authRequired := endpoint.GetBool("auth_required")
			rateLimit := endpoint.GetInt64("rate_limit")

			fmt.Printf("\n  Endpoint #%d:\n", i+1)
			fmt.Printf("    Name: %s\n", name)
			fmt.Printf("    URL: %s\n", url)
			fmt.Printf("    Timeout: %d seconds\n", timeout)
			fmt.Printf("    Retries: %d\n", retries)
			fmt.Printf("    Auth Required: %t\n", authRequired)
			fmt.Printf("    Rate Limit: %d requests/min\n", rateLimit)
		}
	}

	// Example 4: Servers configuration
	fmt.Println("\n=== Servers Configuration ===")
	servers := config.GetObjectSlice("servers")
	for _, server := range servers {
		host := server.GetString("host")
		port := server.GetInt("port")
		secure := server.GetBool("secure")
		protocols := server.GetStringSlice("protocols")

		fmt.Printf("\n  Server: %s:%d\n", host, port)
		fmt.Printf("    Secure: %t\n", secure)
		fmt.Printf("    Protocols: %v\n", protocols)
	}

	// Example 5: Nested configuration object (Service)
	fmt.Println("\n=== Service Configuration ===")
	serviceConfig := config.GetObject("service")
	if serviceConfig != nil {
		name := serviceConfig.GetString("name")
		version := serviceConfig.GetString("version")
		enabled := serviceConfig.GetBool("enabled")

		fmt.Printf("Name: %s\n", name)
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Enabled: %t\n", enabled)

		// Access nested database configuration
		dbConfig := config.GetObject("service.database")
		if dbConfig != nil {
			fmt.Printf("\n  Database:\n")
			fmt.Printf("    Host: %s\n", dbConfig.GetString("host"))
			fmt.Printf("    Port: %d\n", dbConfig.GetInt("port"))
			fmt.Printf("    Name: %s\n", dbConfig.GetString("name"))
			fmt.Printf("    Pool Size: %d\n", dbConfig.GetInt("pool_size"))
			fmt.Printf("    Timeout: %d seconds\n", dbConfig.GetInt("timeout"))
			fmt.Printf("    SSL Mode: %s\n", dbConfig.GetString("ssl_mode"))
		}

		// Access nested cache configuration
		cacheConfig := config.GetObject("service.cache")
		if cacheConfig != nil {
			fmt.Printf("\n  Cache:\n")
			fmt.Printf("    Type: %s\n", cacheConfig.GetString("type"))
			fmt.Printf("    Host: %s\n", cacheConfig.GetString("host"))
			fmt.Printf("    Port: %d\n", cacheConfig.GetInt("port"))
			fmt.Printf("    DB: %d\n", cacheConfig.GetInt("db"))
			fmt.Printf("    TTL: %d seconds\n", cacheConfig.GetInt("ttl"))
		}

		// Access nested monitoring configuration
		monitoringConfig := config.GetObject("service.monitoring")
		if monitoringConfig != nil {
			fmt.Printf("\n  Monitoring:\n")
			fmt.Printf("    Enabled: %t\n", monitoringConfig.GetBool("enabled"))
			fmt.Printf("    Metrics Port: %d\n", monitoringConfig.GetInt("metrics_port"))
			fmt.Printf("    Health Check Interval: %d seconds\n", monitoringConfig.GetInt("health_check_interval"))
			alertEndpoints := monitoringConfig.GetStringSlice("alert_endpoints")
			if alertEndpoints != nil {
				fmt.Printf("    Alert Endpoints: %v\n", alertEndpoints)
			}
		}
	}

	// Example 6: Feature flags configuration
	fmt.Println("\n=== Feature Flags Configuration ===")
	features := config.GetObjectSlice("features")
	if features != nil {
		fmt.Printf("Found %d feature flags:\n", len(features))
		for _, feature := range features {
			name := feature.GetString("name")
			enabled := feature.GetBool("enabled")
			percentage := feature.GetInt("percentage")
			description := feature.GetString("description")
			rolloutDate := feature.GetString("rollout_date")

			status := "DISABLED"
			if enabled {
				status = fmt.Sprintf("ENABLED (%d%% rollout)", percentage)
			}

			fmt.Printf("\n  Feature: %s\n", name)
			fmt.Printf("    Status: %s\n", status)
			fmt.Printf("    Description: %s\n", description)
			fmt.Printf("    Rollout Date: %s\n", rolloutDate)
		}
	}

	// Example 7: Dynamic object modification
	fmt.Println("\n=== Dynamic Object Modification ===")
	demonstrateObjectModification(config)

	// Example 8: Error handling demonstration
	fmt.Println("\n=== Error Handling Examples ===")
	demonstrateErrorHandling(config)
}

// demonstrateObjectModification shows how to create and set objects dynamically
func demonstrateObjectModification(config cli.ConfigFileTyped) {
	fmt.Println("Creating new dynamic objects...")

	// Create a new endpoint object
	newEndpoint := cli.NewTypedConfigObject()
	newEndpoint.SetString("name", "dynamic-endpoint")
	newEndpoint.SetString("url", "https://dynamic.example.com/v1")
	newEndpoint.SetInt("timeout", 45)
	newEndpoint.SetInt("retries", 4)
	newEndpoint.SetBool("auth_required", true)
	newEndpoint.SetInt64("rate_limit", 1500)
	newEndpoint.SetStringSlice("protocols", []string{"https", "http2", "grpc"})

	// Add it to the existing endpoints
	fmt.Println("Adding new endpoint to existing list...")
	endpoints := config.GetObjectSlice("endpoints")
	if endpoints == nil {
		// If no endpoints exist, create a new array
		endpoints = []cli.ConfigFileTyped{newEndpoint}
	} else {
		// Append to existing array
		endpoints = append(endpoints, newEndpoint)
	}

	if err := config.SetObjectSlice("endpoints", endpoints); err != nil {
		fmt.Printf("Error setting endpoints: %v\n", err)
	} else {
		fmt.Println("Successfully added new endpoint!")

		// Verify it was added
		updatedEndpoints := config.GetObjectSlice("endpoints")
		if updatedEndpoints != nil && len(updatedEndpoints) > 0 {
			lastEndpoint := updatedEndpoints[len(updatedEndpoints)-1]
			fmt.Printf("  New endpoint: %s (timeout: %d, auth: %t)\n",
				lastEndpoint.GetString("name"),
				lastEndpoint.GetInt("timeout"),
				lastEndpoint.GetBool("auth_required"))
		}
	}

	// Create a completely new service configuration
	fmt.Println("\nCreating new service configuration...")
	newService := cli.NewTypedConfigObject()
	newService.SetString("name", "auth-service")
	newService.SetString("version", "1.0.0")
	newService.SetBool("enabled", true)
	newService.SetInt("port", 9000)
	newService.SetStringSlice("trusted_ips", []string{"127.0.0.1", "10.0.0.0/8"})

	// Create nested database config
	dbConfig := cli.NewTypedConfigObject()
	dbConfig.SetString("driver", "postgres")
	dbConfig.SetString("host", "db.auth.internal")
	dbConfig.SetInt("port", 5432)
	dbConfig.SetString("name", "auth_db")
	dbConfig.SetBool("ssl_enabled", true)
	dbConfig.SetInt("pool_size", 25)

	// Set the database as a nested object
	if err := newService.SetObject("database", dbConfig); err != nil {
		fmt.Printf("Error setting database config: %v\n", err)
	} else {
		fmt.Println("Successfully created nested database configuration!")
	}

	// Set the entire service
	if err := config.SetObject("auth_service", newService); err != nil {
		fmt.Printf("Error setting auth_service: %v\n", err)
	} else {
		fmt.Println("Successfully created auth_service configuration!")

		// Verify it was set
		if authService := config.GetObject("auth_service"); authService != nil {
			fmt.Printf("  Auth Service: %s v%s (port: %d)\n",
				authService.GetString("name"),
				authService.GetString("version"),
				authService.GetInt("port"))

			if db := authService.GetObject("database"); db != nil {
				fmt.Printf("    Database: %s at %s:%d\n",
					db.GetString("driver"),
					db.GetString("host"),
					db.GetInt("port"))
			}
		}
	}

	// Modify an existing server
	fmt.Println("\nModifying existing server configuration...")
	servers := config.GetObjectSlice("servers")
	if servers != nil && len(servers) > 0 {
		// Update the first server
		server0 := servers[0]
		fmt.Printf("Original server: %s:%d\n", server0.GetString("host"), server0.GetInt("port"))

		// Update values
		server0.SetInt("port", 8443)
		server0.SetBool("secure", false)

		// Save the modified servers back
		if err := config.SetObjectSlice("servers", servers); err != nil {
			fmt.Printf("Error updating servers: %v\n", err)
		} else {
			fmt.Printf("Updated server: %s:%d\n", server0.GetString("host"), server0.GetInt("port"))
			fmt.Printf("  Secure changed from true to: %t\n", server0.GetBool("secure"))
		}
	}
}

// Helper functions to mask sensitive information
func maskApiKey(key string) string {
	if key == "" {
		return "[empty]"
	}
	if len(key) <= 8 {
		return "[***]"
	}
	return key[:4] + "***" + key[len(key)-4:]
}

// demonstrateErrorHandling shows how to handle missing or invalid data gracefully
func demonstrateErrorHandling(config cli.ConfigFileTyped) {
	// Try to access a non-existent slice
	missingSlice := config.GetObjectSlice("non_existent")
	if missingSlice == nil {
		fmt.Println("✓ Handled missing slice gracefully")
	}

	// Try to access a non-existent object
	missingObject := config.GetObject("config.does_not_exist")
	if missingObject == nil {
		fmt.Println("✓ Handled missing object gracefully")
	}

	// Try to access with invalid type conversion
	// Even if the value exists but is not a string, GetString returns ""
	nonString := config.GetString("debug") // debug is a bool
	if nonString == "" {
		fmt.Printf("✓ Got empty string when accessing bool 'debug' as string\n")
	}

	// Check if a key exists without getting the value
	if _, exists := config.GetValue("timeout"); exists {
		fmt.Println("✓ 'timeout' key exists")
	}

	// Handle potentially missing keys with default values
	optionalValue := config.GetString("optional.setting")
	if optionalValue == "" {
		fmt.Printf("✓ Optional setting not found, using default: [default_value]\n")
	}
}
