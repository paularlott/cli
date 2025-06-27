package cli

type ConfigFileTyped interface {
	ConfigFileSource

	GetString(string) string                 // Get the string value from the configuration file at the specified path.
	GetInt(string) int                       // Get the int value from the configuration file at the specified path.
	GetInt64(string) int64                   // Get the int64 value from the configuration file at the specified path.
	GetInt32(string) int32                   // Get the int32 value from the configuration file at the specified path.
	GetInt16(string) int16                   // Get the int16 value from the configuration file at the specified path.
	GetInt8(string) int8                     // Get the int8 value from the configuration file at the specified path.
	GetUint(string) uint                     // Get the uint value from the configuration file at the specified path.
	GetUint64(string) uint64                 // Get the uint64 value from the configuration file at the specified path.
	GetUint32(string) uint32                 // Get the uint32 value from the configuration file at the specified path.
	GetUint16(string) uint16                 // Get the uint16 value from the configuration file at the specified path.
	GetUint8(string) uint8                   // Get the uint8 value from the configuration file at the specified path.
	GetFloat32(string) float32               // Get the float32 value from the configuration file at the specified path.
	GetFloat64(string) float64               // Get the float64 value from the configuration file at the specified path.
	GetBool(string) bool                     // Get the bool value from the configuration file at the specified path.
	GetStringSlice(string) []string          // Get the string slice value from the configuration file at the specified path.
	GetIntSlice(string) []int                // Get the int slice value from the configuration file at the specified path.
	GetInt64Slice(string) []int64            // Get the int64 slice value from the configuration file at the specified path.
	GetInt32Slice(string) []int32            // Get the int32 slice value from the configuration file at the specified path.
	GetInt16Slice(string) []int16            // Get the int16 slice value from the configuration file at the specified path.
	GetInt8Slice(string) []int8              // Get the int8 slice value from the configuration file at the specified path.
	GetUintSlice(string) []uint              // Get the uint slice value from the configuration file at the specified path.
	GetUint64Slice(string) []uint64          // Get the uint64 slice value from the configuration file at the specified path.
	GetUint32Slice(string) []uint32          // Get the uint32 slice value from the configuration file at the specified path.
	GetUint16Slice(string) []uint16          // Get the uint16 slice value from the configuration file at the specified path.
	GetUint8Slice(string) []uint8            // Get the uint8 slice value from the configuration file at the specified path.
	GetFloat32Slice(string) []float32        // Get the float32 slice value from the configuration file at the specified path.
	GetFloat64Slice(string) []float64        // Get the float64 slice value from the configuration file at the specified path.
	SetString(string, string) error          // Set the string value in the configuration file at the specified path.
	SetInt(string, int) error                // Set the int value in the configuration file at the specified path.
	SetInt64(string, int64) error            // Set the int64 value in the configuration file at the specified path.
	SetInt32(string, int32) error            // Set the int32 value in the configuration file at the specified path.
	SetInt16(string, int16) error            // Set the int16 value in the configuration file at the specified path.
	SetInt8(string, int8) error              // Set the int8 value in the configuration file at the specified path.
	SetUint(string, uint) error              // Set the uint value in the configuration file at the specified path.
	SetUint64(string, uint64) error          // Set the uint64 value in the configuration file at the specified path.
	SetUint32(string, uint32) error          // Set the uint32 value in the configuration file at the specified path.
	SetUint16(string, uint16) error          // Set the uint16 value in the configuration file at the specified path.
	SetUint8(string, uint8) error            // Set the uint8 value in the configuration file at the specified path.
	SetFloat32(string, float32) error        // Set the float32 value in the configuration file at the specified path.
	SetFloat64(string, float64) error        // Set the float64 value in the configuration file at the specified path.
	SetBool(string, bool) error              // Set the bool value in the configuration file at the specified path.
	SetStringSlice(string, []string) error   // Set the string slice value in the configuration file at the specified path.
	SetIntSlice(string, []int) error         // Set the int slice value in the configuration file at the specified path.
	SetInt64Slice(string, []int64) error     // Set the int64 slice value in the configuration file at the specified path.
	SetInt32Slice(string, []int32) error     // Set the int32 slice value in the configuration file at the specified path.
	SetInt16Slice(string, []int16) error     // Set the int16 slice value in the configuration file at the specified path.
	SetInt8Slice(string, []int8) error       // Set the int8 slice value in the configuration file at the specified path.
	SetUintSlice(string, []uint) error       // Set the uint slice value in the configuration file at the specified path.
	SetUint64Slice(string, []uint64) error   // Set the uint64 slice value in the configuration file at the specified path.
	SetUint32Slice(string, []uint32) error   // Set the uint32 slice value in the configuration file at the specified path.
	SetUint16Slice(string, []uint16) error   // Set the uint16 slice value in the configuration file at the specified path.
	SetUint8Slice(string, []uint8) error     // Set the uint8 slice value in the configuration file at the specified path.
	SetFloat32Slice(string, []float32) error // Set the float32 slice value in the configuration file at the specified path.
	SetFloat64Slice(string, []float64) error // Set the float64 slice value in the configuration file at the specified path.
}

type ConfigFileTypedWrapper struct {
	inner ConfigFileSource
}

var _ ConfigFileTyped = (*ConfigFileTypedWrapper)(nil)

func NewTypedConfigFile(inner ConfigFileSource) *ConfigFileTypedWrapper {
	return &ConfigFileTypedWrapper{
		inner: inner,
	}
}

func (w *ConfigFileTypedWrapper) GetValue(path string) (any, bool)  { return w.inner.GetValue(path) }
func (w *ConfigFileTypedWrapper) GetKeys(path string) []string      { return w.inner.GetKeys(path) }
func (w *ConfigFileTypedWrapper) SetValue(path string, v any) error { return w.inner.SetValue(path, v) }
func (w *ConfigFileTypedWrapper) DeleteKey(path string) error       { return w.inner.DeleteKey(path) }
func (w *ConfigFileTypedWrapper) Save() error                       { return w.inner.Save() }
func (w *ConfigFileTypedWrapper) OnChange(h ConfigFileChangeHandler) error {
	return w.inner.OnChange(h)
}
func (w *ConfigFileTypedWrapper) FileUsed() string { return w.inner.FileUsed() }

func getAs[T any](c ConfigFileSource, path string) T {
	var zero T
	if value, exists := c.GetValue(path); exists {
		if cast, ok := value.(T); ok {
			return cast
		}
	}
	return zero
}

func (c *ConfigFileTypedWrapper) GetString(path string) string {
	return getAs[string](c.inner, path)
}

func (c *ConfigFileTypedWrapper) GetInt(path string) int {
	return getAs[int](c, path)
}

func (c *ConfigFileTypedWrapper) GetInt64(path string) int64 {
	return getAs[int64](c, path)
}

func (c *ConfigFileTypedWrapper) GetInt32(path string) int32 {
	return getAs[int32](c, path)
}

func (c *ConfigFileTypedWrapper) GetInt16(path string) int16 {
	return getAs[int16](c, path)
}

func (c *ConfigFileTypedWrapper) GetInt8(path string) int8 {
	return getAs[int8](c, path)
}

func (c *ConfigFileTypedWrapper) GetUint(path string) uint {
	return getAs[uint](c, path)
}

func (c *ConfigFileTypedWrapper) GetBool(path string) bool {
	return getAs[bool](c, path)
}

func (c *ConfigFileTypedWrapper) GetUint64(path string) uint64 {
	return getAs[uint64](c, path)
}

func (c *ConfigFileTypedWrapper) GetUint32(path string) uint32 {
	return getAs[uint32](c, path)
}

func (c *ConfigFileTypedWrapper) GetUint16(path string) uint16 {
	return getAs[uint16](c, path)
}

func (c *ConfigFileTypedWrapper) GetUint8(path string) uint8 {
	return getAs[uint8](c, path)
}

func (c *ConfigFileTypedWrapper) GetFloat32(path string) float32 {
	return getAs[float32](c, path)
}

func (c *ConfigFileTypedWrapper) GetFloat64(path string) float64 {
	return getAs[float64](c, path)
}

func (c *ConfigFileTypedWrapper) GetStringSlice(path string) []string {
	return getAs[[]string](c, path)
}

func (c *ConfigFileTypedWrapper) GetIntSlice(path string) []int {
	return getAs[[]int](c, path)
}

func (c *ConfigFileTypedWrapper) GetInt64Slice(path string) []int64 {
	return getAs[[]int64](c, path)
}

func (c *ConfigFileTypedWrapper) GetInt32Slice(path string) []int32 {
	return getAs[[]int32](c, path)
}

func (c *ConfigFileTypedWrapper) GetInt16Slice(path string) []int16 {
	return getAs[[]int16](c, path)
}

func (c *ConfigFileTypedWrapper) GetInt8Slice(path string) []int8 {
	return getAs[[]int8](c, path)
}

func (c *ConfigFileTypedWrapper) GetUintSlice(path string) []uint {
	return getAs[[]uint](c, path)
}

func (c *ConfigFileTypedWrapper) GetUint64Slice(path string) []uint64 {
	return getAs[[]uint64](c, path)
}

func (c *ConfigFileTypedWrapper) GetUint32Slice(path string) []uint32 {
	return getAs[[]uint32](c, path)
}

func (c *ConfigFileTypedWrapper) GetUint16Slice(path string) []uint16 {
	return getAs[[]uint16](c, path)
}

func (c *ConfigFileTypedWrapper) GetUint8Slice(path string) []uint8 {
	return getAs[[]uint8](c, path)
}

func (c *ConfigFileTypedWrapper) GetFloat32Slice(path string) []float32 {
	return getAs[[]float32](c, path)
}

func (c *ConfigFileTypedWrapper) GetFloat64Slice(path string) []float64 {
	return getAs[[]float64](c, path)
}

func (c *ConfigFileTypedWrapper) SetString(path string, value string) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt(path string, value int) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt64(path string, value int64) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt32(path string, value int32) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt16(path string, value int16) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt8(path string, value int8) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint(path string, value uint) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint64(path string, value uint64) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint32(path string, value uint32) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint16(path string, value uint16) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint8(path string, value uint8) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetFloat32(path string, value float32) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetFloat64(path string, value float64) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetBool(path string, value bool) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetStringSlice(path string, value []string) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetIntSlice(path string, value []int) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt64Slice(path string, value []int64) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt32Slice(path string, value []int32) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt16Slice(path string, value []int16) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetInt8Slice(path string, value []int8) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUintSlice(path string, value []uint) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint64Slice(path string, value []uint64) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint32Slice(path string, value []uint32) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint16Slice(path string, value []uint16) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetUint8Slice(path string, value []uint8) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetFloat32Slice(path string, value []float32) error {
	return c.SetValue(path, value)
}

func (c *ConfigFileTypedWrapper) SetFloat64Slice(path string, value []float64) error {
	return c.SetValue(path, value)
}
