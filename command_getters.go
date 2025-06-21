package cli

// Flag getters
func (c *Command) GetString(name string) string {
	if v, ok := c.parsedFlags[name]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c *Command) GetInt64(name string) int64 {
	if v, ok := c.parsedFlags[name]; ok {
		if i, ok := v.(int64); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetInt(name string) int {
	if v, ok := c.parsedFlags[name]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetInt8(name string) int8 {
	if v, ok := c.parsedFlags[name]; ok {
		if i, ok := v.(int8); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetInt16(name string) int16 {
	if v, ok := c.parsedFlags[name]; ok {
		if i, ok := v.(int16); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetInt32(name string) int32 {
	if v, ok := c.parsedFlags[name]; ok {
		if i, ok := v.(int32); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetUint64(name string) uint64 {
	if v, ok := c.parsedFlags[name]; ok {
		if u, ok := v.(uint64); ok {
			return u
		}
	}
	return 0
}

func (c *Command) GetUint(name string) uint {
	if v, ok := c.parsedFlags[name]; ok {
		if u, ok := v.(uint); ok {
			return u
		}
	}
	return 0
}

func (c *Command) GetUint8(name string) uint8 {
	if v, ok := c.parsedFlags[name]; ok {
		if u, ok := v.(uint8); ok {
			return u
		}
	}
	return 0
}

func (c *Command) GetUint16(name string) uint16 {
	if v, ok := c.parsedFlags[name]; ok {
		if u, ok := v.(uint16); ok {
			return u
		}
	}
	return 0
}

func (c *Command) GetUint32(name string) uint32 {
	if v, ok := c.parsedFlags[name]; ok {
		if u, ok := v.(uint32); ok {
			return u
		}
	}
	return 0
}

func (c *Command) GetFloat32(name string) float32 {
	if v, ok := c.parsedFlags[name]; ok {
		if f, ok := v.(float32); ok {
			return f
		}
	}
	return 0.0
}

func (c *Command) GetFloat64(name string) float64 {
	if v, ok := c.parsedFlags[name]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0.0
}

func (c *Command) GetBool(name string) bool {
	if v, ok := c.parsedFlags[name]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// Slice getters
func (c *Command) GetStringSlice(name string) []string {
	if v, ok := c.parsedFlags[name]; ok {
		if s, ok := v.([]string); ok {
			return s
		}
	}
	return nil
}

func (c *Command) GetIntSlice(name string) []int {
	if v, ok := c.parsedFlags[name]; ok {
		if s, ok := v.([]int); ok {
			return s
		}
	}
	return nil
}

func (c *Command) GetInt8Slice(name string) []int8 {
	if v, ok := c.parsedFlags[name]; ok {
		if s, ok := v.([]int8); ok {
			return s
		}
	}
	return nil
}

func (c *Command) GetInt16Slice(name string) []int16 {
	if v, ok := c.parsedFlags[name]; ok {
		if s, ok := v.([]int16); ok {
			return s
		}
	}
	return nil
}

func (c *Command) GetInt32Slice(name string) []int32 {
	if v, ok := c.parsedFlags[name]; ok {
		if s, ok := v.([]int32); ok {
			return s
		}
	}
	return nil
}

func (c *Command) GetInt64Slice(name string) []int64 {
	if v, ok := c.parsedFlags[name]; ok {
		if s, ok := v.([]int64); ok {
			return s
		}
	}
	return nil
}

func (c *Command) GetUintSlice(name string) []uint {
	if v, ok := c.parsedFlags[name]; ok {
		if s, ok := v.([]uint); ok {
			return s
		}
	}
	return nil
}

func (c *Command) GetUint8Slice(name string) []uint8 {
	if v, ok := c.parsedFlags[name]; ok {
		if s, ok := v.([]uint8); ok {
			return s
		}
	}
	return nil
}

func (c *Command) GetUint16Slice(name string) []uint16 {
	if v, ok := c.parsedFlags[name]; ok {
		if s, ok := v.([]uint16); ok {
			return s
		}
	}
	return nil
}

func (c *Command) GetUint32Slice(name string) []uint32 {
	if v, ok := c.parsedFlags[name]; ok {
		if s, ok := v.([]uint32); ok {
			return s
		}
	}
	return nil
}

func (c *Command) GetUint64Slice(name string) []uint64 {
	if v, ok := c.parsedFlags[name]; ok {
		if s, ok := v.([]uint64); ok {
			return s
		}
	}
	return nil
}

func (c *Command) GetFloat32Slice(name string) []float32 {
	if v, ok := c.parsedFlags[name]; ok {
		if s, ok := v.([]float32); ok {
			return s
		}
	}
	return nil
}

func (c *Command) GetFloat64Slice(name string) []float64 {
	if v, ok := c.parsedFlags[name]; ok {
		if s, ok := v.([]float64); ok {
			return s
		}
	}
	return nil
}

// Argument getters
func (c *Command) GetStringArg(name string) string {
	if v, ok := c.parsedArgs[name]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c *Command) GetBoolArg(name string) bool {
	if v, ok := c.parsedArgs[name]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func (c *Command) GetInt64Arg(name string) int64 {
	if v, ok := c.parsedArgs[name]; ok {
		if i, ok := v.(int64); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetIntArg(name string) int {
	if v, ok := c.parsedArgs[name]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetInt32Arg(name string) int32 {
	if v, ok := c.parsedArgs[name]; ok {
		if i, ok := v.(int32); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetInt16Arg(name string) int16 {
	if v, ok := c.parsedArgs[name]; ok {
		if i, ok := v.(int16); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetInt8Arg(name string) int8 {
	if v, ok := c.parsedArgs[name]; ok {
		if i, ok := v.(int8); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetUint64Arg(name string) uint64 {
	if v, ok := c.parsedArgs[name]; ok {
		if i, ok := v.(uint64); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetUintArg(name string) uint {
	if v, ok := c.parsedArgs[name]; ok {
		if i, ok := v.(uint); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetUint32Arg(name string) uint32 {
	if v, ok := c.parsedArgs[name]; ok {
		if i, ok := v.(uint32); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetUint16Arg(name string) uint16 {
	if v, ok := c.parsedArgs[name]; ok {
		if i, ok := v.(uint16); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetUint8Arg(name string) uint8 {
	if v, ok := c.parsedArgs[name]; ok {
		if i, ok := v.(uint8); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetFloat64Arg(name string) float64 {
	if v, ok := c.parsedArgs[name]; ok {
		if i, ok := v.(float64); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetFloat32Arg(name string) float32 {
	if v, ok := c.parsedArgs[name]; ok {
		if i, ok := v.(float32); ok {
			return i
		}
	}
	return 0
}
