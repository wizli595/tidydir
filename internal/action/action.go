package action

// ActionType represents the kind of operation.
type ActionType string

const (
	ActionMove   ActionType = "move"
	ActionDelete ActionType = "delete"
	ActionRename ActionType = "rename"
)

// Action represents a single file operation to be executed.
type Action struct {
	Type   ActionType
	Source string
	Dest   string // empty for delete
	Reason string
}
