package client

import (
	"io"
	"log"
	"os"

	"github.com/firelizzard18/Tea-Service/common"
)

var CmdDefaults = Command{OutType: "both"}

type Command struct {
	OutType string `long:"out-type" description:"the output {type} when connecting to a server (none, out, err, both)"`
}

func (m *Command) GetOutType() common.OutputType {
	switch m.OutType {
	case "none":
		return common.OutputNone
	case "out":
		return common.OutputOut
	case "err":
		return common.OutputErr
	case "both":
		return common.OutputAll

	default:
		return common.OutputInvalid
	}
}

func DumpOutput(source *os.File) {
	if source == nil {
		return
	}

	defer source.Close()

	b := make([]byte, 256)
	for {
		n, err := source.Read(b)
		if err == io.EOF {
			log.Print("The output file descriptor has been closed")
			return
		}

		for o := 0; o < n; {
			m, err := os.Stdout.Write(b[o:n])
			if err != nil {
				log.Panic(err)
				return
			}
			o += m
		}
	}
}

func SendInput(sink *os.File) {
	defer sink.Close()

	b := make([]byte, 256)
	for {
		n, err := os.Stdin.Read(b)
		if err == io.EOF {
			log.Panic("Stdin has been closed")
			return
		}

		for o := 0; o < n; {
			m, err := sink.Write(b[o:n])
			if err != nil {
				log.Print(err)
				return
			}
			o += m
		}
	}
}
