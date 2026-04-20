package cmd

type ExecuteFlags struct {
	IsDefineInfo    bool
	IsDefineAllInfo bool
	IsSuggest       bool
	IsAllType       bool
	IsExtends       bool
	IsHover         bool
	IsVersion       bool
	IsHelp          bool
	IsStdin         bool
}

func NewExecuteFlags() *ExecuteFlags {
	return &ExecuteFlags{}
}
