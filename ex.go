package ex

type (
	Target    = string
	FlagValue = string
)

type Flag struct {
	ServiceName string
	ExploitName string
	Value       FlagValue
	From        Target
	Status      SubmittedFlagStatus
}
