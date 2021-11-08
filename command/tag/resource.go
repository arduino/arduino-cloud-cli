package tag

// ResourceType specifies which resource the
// tag command refers to.
// Valid resources are the entities of the
// cloud that have a 'tags' field.
type ResourceType int

const (
	None ResourceType = iota
	Device
	Thing
)
