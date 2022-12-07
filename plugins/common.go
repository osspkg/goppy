package plugins

// Plugin plugin structure
type Plugin struct {
	Config  interface{}
	Inject  interface{}
	Resolve interface{}
}

// Defaulter interface for setting default values for a structure
type Defaulter interface {
	Default()
}
