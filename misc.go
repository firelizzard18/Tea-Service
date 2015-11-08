package main

type OutputType int

const (
   None OutputType = iota
   Output
   Error
   OutAndErr
   Invalid
)

type empty struct{}

type ServerInfo struct {
	Sender      string
	Description string
}