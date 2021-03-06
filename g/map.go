package g

// Map is a shortcut for map[string]interface{}
type Map map[string]interface{}

// NewMap Allocate a Map.
func NewMap() Map {
	m := make(map[string]interface{}, 0)
	return m
}

// Map Cast a Map to map[string]interface{}
func (mv Map) Map() map[string]interface{} {
	return mv
}
