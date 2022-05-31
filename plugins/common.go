package plugins

type Plugin struct {
	Config       interface{}
	Inject       interface{}
	Dependencies []Plugin
}
