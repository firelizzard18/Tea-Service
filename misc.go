package main



type OutputType int

const (
   Output_None OutputType = iota
   Output_Out
   Output_Err
   Output_All
   Output_Invalid
)



type empty struct{}

type ServerInfo struct {
	Sender      string
	Description string
}