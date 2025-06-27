package cli

type Argument interface {
	name() string
	usage() string
	isRequired() bool
	typeText() string
	validateArg(*Command) error
}

type ArgumentTyped[T any] struct {
	Name        string               // Name of the argument
	Usage       string               // Usage description for the argument
	Required    bool                 // Whether this flag is required
	AssignTo    *T                   // Optional pointer to the variable where the value should be stored
	ValidateArg func(*Command) error // Optional validation for the argument
}

func (a *ArgumentTyped[T]) name() string {
	return a.Name
}

func (a *ArgumentTyped[T]) usage() string {
	return a.Usage
}

func (a *ArgumentTyped[T]) isRequired() bool {
	return a.Required
}

// Add a typeText method to ArgumentTyped similar to FlagTyped
func (a *ArgumentTyped[T]) typeText() string {
	var zero T
	return GetTypeText(zero)
}

func (a *ArgumentTyped[T]) validateArg(c *Command) error {
	if a.ValidateArg != nil {
		return a.ValidateArg(c)
	}
	return nil
}

type StringArg = ArgumentTyped[string]
type IntArg = ArgumentTyped[int]
type Int8Arg = ArgumentTyped[int8]
type Int16Arg = ArgumentTyped[int16]
type Int32Arg = ArgumentTyped[int32]
type Int64Arg = ArgumentTyped[int64]
type UintArg = ArgumentTyped[uint]
type Uint8Arg = ArgumentTyped[uint8]
type Uint16Arg = ArgumentTyped[uint16]
type Uint32Arg = ArgumentTyped[uint32]
type Uint64Arg = ArgumentTyped[uint64]
type Float32Arg = ArgumentTyped[float32]
type Float64Arg = ArgumentTyped[float64]
type BoolArg = ArgumentTyped[bool]
