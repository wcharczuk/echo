package web

// State is the collection of state objects on a context.
type State map[interface{}]interface{}

// Value returns a value from the state bag.
func (s State) Value(key interface{}) interface{} {
	return s[key]
}

// StateProvider provide states, an example is Ctx
type StateProvider interface {
	State() State
}

// StateValueProvider is a type that provides a state value.
type StateValueProvider interface {
	StateValue(key string) interface{}
}
