package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type ConfigFileSource interface {
	GetValue(string) (any, bool)            // Get the value from the configuration file at the specified path.
	GetKeys(string) []string                // Get the keys from the configuration file at the specified path.
	SetValue(string, any) error             // Set a value in the configuration file at the specified path.
	DeleteKey(string) error                 // Delete a key from the configuration file at the specified path.
	Save() error                            // Save the configuration file.
	OnChange(ConfigFileChangeHandler) error // Track changes to the configuration file.
	FileUsed() string                       // Get the file used for the configuration.
	LoadData() error                        // Load the configuration data from the file.
}

type SearchPathFunc func() []string
type ConfigFileUnmarshal func(data []byte, v any) error
type ConfigFileMarshal func(v any) ([]byte, error)
type ConfigFileChangeHandler func()

type ConfigFileBase struct {
	FileName      *string                 // Point to the configuration file name
	SearchPath    SearchPathFunc          // Function to define the search paths for the config file
	Unmarshal     ConfigFileUnmarshal     // Function to decode the configuration file content
	Marshal       ConfigFileMarshal       // Function to encode the configuration file content
	data          map[string]any          // Parsed configuration data
	isLoaded      bool                    // Indicates if the configuration file has been loaded
	mutex         sync.Mutex              // Mutex for thread-safe access to the configuration data
	fileUsed      string                  // The file that was used to load the configuration
	watcher       *fsnotify.Watcher       // File system watcher for monitoring changes
	changeHandler ConfigFileChangeHandler // Change handler for config file changes
}

var _ ConfigFileSource = (*ConfigFileBase)(nil)

func (c *ConfigFileBase) InitConfigFile() {
	c.data = make(map[string]any)
	c.isLoaded = false
	c.mutex = sync.Mutex{}
	c.fileUsed = ""
}

// searchForConfigFile searches for the configuration file in the defined search paths and returns the file name including the path, if not found it returns an empty string.
func (c *ConfigFileBase) searchForConfigFile() string {
	var paths []string

	if c.SearchPath == nil {
		paths = []string{"."}
	} else {
		paths = c.SearchPath()
	}

	fileName := *c.FileName

	// If the filename given is to a file then accept it and don't search for other possible matches
	if info, err := os.Stat(fileName); !os.IsNotExist(err) && !info.IsDir() {
		return fileName
	}

	// Search for the configuration file in the search paths
	dotFileRegex := regexp.MustCompile(`^\.[^./]`)
	lookForDotOnly := dotFileRegex.MatchString(fileName)

	for _, path := range paths {
		var candidates []string
		if lookForDotOnly {
			candidates = []string{filepath.Join(path, fileName)}
		} else {
			candidates = []string{
				filepath.Join(path, fileName),
				filepath.Join(path, "."+fileName),
			}
		}
		for _, candidate := range candidates {
			info, err := os.Stat(candidate)
			if err == nil && !info.IsDir() {
				return candidate
			}
		}
	}

	return ""
}

func (c *ConfigFileBase) LoadData() error {
	if !c.isLoaded {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		if !c.isLoaded {
			filename := c.searchForConfigFile()
			if filename == "" {
				return fmt.Errorf("configuration file not found")
			}

			contentBytes, err := os.ReadFile(filename)
			if err != nil {
				fmt.Println(err)
				return err
			}

			if err := c.Unmarshal(contentBytes, &c.data); err != nil {
				fmt.Println(err)
				return err
			}

			c.isLoaded = true
			c.fileUsed = filename
		}
	}

	return nil
}

func (c *ConfigFileBase) Save() error {
	if !c.isLoaded || c.fileUsed == "" {
		// Assume the filename points to where the file should be created
		c.fileUsed = *c.FileName
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	contentBytes, err := c.Marshal(c.data)
	if err != nil {
		return err
	}

	return os.WriteFile(c.fileUsed, contentBytes, 0644)
}

func (c *ConfigFileBase) FileUsed() string {
	c.LoadData()
	if !c.isLoaded {
		return ""
	}

	return c.fileUsed
}

func (c *ConfigFileBase) traversePath(keys []string, current map[string]any) (map[string]any, bool) {
	for _, key := range keys {
		// Navigate deeper into the nested structure
		if val, exists := current[key]; exists {
			if nextMap, ok := val.(map[string]any); ok {
				current = nextMap
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}

	return current, true
}

func (c *ConfigFileBase) GetValue(path string) (any, bool) {
	if err := c.LoadData(); err != nil {
		return nil, false
	}

	// Extract the value based on the provided path
	keys := strings.Split(path, ".")
	current := c.data

	var exists bool
	if current, exists = c.traversePath(keys[:len(keys)-1], current); exists {
		if val, exists := current[keys[len(keys)-1]]; exists {
			return val, true
		}
	}

	return nil, false
}

func (c *ConfigFileBase) GetKeys(path string) []string {
	if err := c.LoadData(); err != nil {
		return nil
	}

	current := c.data
	if path != "" {
		var exists bool
		current, exists = c.traversePath(strings.Split(path, "."), current)
		if !exists {
			return nil
		}
	}

	// Return all keys in the current map
	result := make([]string, 0, len(current))
	for k := range current {
		result = append(result, k)
	}

	return result
}

func (c *ConfigFileBase) SetValue(path string, value any) error {
	// Extract the keys from the path
	keys := strings.Split(path, ".")
	current := c.data

	// Traverse to the second last key
	for _, key := range keys[:len(keys)-1] {
		if nextMap, exists := current[key]; exists {
			if nextMap, ok := nextMap.(map[string]any); ok {
				current = nextMap
			} else {
				return fmt.Errorf("path %s is not a valid map", path)
			}
		} else {
			// Create a new map if it doesn't exist
			newMap := make(map[string]any)
			current[key] = newMap
			current = newMap
		}
	}

	// Set the value at the final key
	current[keys[len(keys)-1]] = value

	return nil
}

func (c *ConfigFileBase) DeleteKey(path string) error {
	if err := c.LoadData(); err != nil {
		return err
	}

	// Extract the value based on the provided path
	keys := strings.Split(path, ".")
	current := c.data

	var exists bool
	if current, exists = c.traversePath(keys[:len(keys)-1], current); exists {
		delete(current, keys[len(keys)-1])
	}

	return nil
}
