package common

type OutputType int

const (
	OutputNone OutputType = iota
	OutputOut
	OutputErr
	OutputAll
	OutputInvalid
)
