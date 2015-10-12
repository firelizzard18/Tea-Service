package teasvc

import (
   "io"
   "os"
   "log"
   "os/exec"
   "errors"
//   "sync"
)

type Process struct {
   cmd *exec.Cmd
   exported *DBusServer

   inPipe io.WriteCloser
   outPipe io.ReadCloser
   errPipe io.ReadCloser

   inProvider chan chan []byte
   outListeners *ListenerList
   errListeners *ListenerList
   
   Description string
}

type OutputType int

const (
   Output OutputType = iota
   Error
   OutAndErr
   Invalid
)

func StartProcess(name string, desc string, arg ...string) (p *Process, err error) {
   p = new(Process)
   p.cmd = exec.Command(name, arg...)
   p.Description = desc

   p.inPipe, err = p.cmd.StdinPipe()
   if err != nil {
      return nil, errors.New("Could not open process stdin: " + err.Error())
   }

   p.outPipe, err = p.cmd.StdoutPipe()
   if err != nil {
      return nil, errors.New("Could not open process stdout: " + err.Error())
   }

   p.errPipe, err = p.cmd.StderrPipe()
   if err != nil {
      return nil, errors.New("Could not open process stderr: " + err.Error())
   }

   p.exported, err = ExportToDBus(p)
   if err != nil {
      return nil, errors.New("Could not open dbus: " + err.Error())
   }
   
   p.inProvider = make(chan chan []byte, 1)
   p.outListeners = NewListenerList()
   p.errListeners = NewListenerList()

   go p.forwardOut(p.outPipe, p.outListeners)
   go p.forwardOut(p.errPipe, p.errListeners)

   // do something about input
   
   return
}

func (p *Process) forwardOut(pipe io.Reader, list *ListenerList) {
   data := make([]byte, 0, 256)
   for {
      n, err := pipe.Read(data)
      if (err != nil) {
         panic(err)
      }

      for node := range list.Traverse() {
         node.sink <- data[:n]
      }
   }
}

func (p *Process) forwardOutPipe(w *os.File, node *ListenerNode) {
   var n int
   var err error
   
outer:
   for data := range node.sink {
      s := 0
      for s < len(data) {
         n, err = w.Write(data[n:])
         if err != nil {
            break outer
         }
         s += n
      }
   }
   
   log.Print(err)
   node.Remove()
   close(node.sink)
   return
}

func (p *Process) forwardInPipe(r *os.File, c chan []byte) {
   var n int
   var err error
   
   data := make([]byte, 0, 256)
   for {
      n, err = r.Read(data)
      if (err != nil) {
         break
      }
      
      c <- data[:n]
   }
   
   log.Print(err)
   close(c)
   return
}

func (p *Process) RequestOutput(otype OutputType) (*os.File, error) {
   if (otype >= Invalid) {
      return nil, errors.New("Invalid output type")
   }
   
   outRead, outWrite, err := os.Pipe()
   if (err != nil) {
      return nil, err
   }
   
   if otype == Output || otype == OutAndErr {
      node := p.outListeners.Append()
      go p.forwardOutPipe(outWrite, node)
   }
   
   if otype == Error || otype == OutAndErr {
      node := p.errListeners.Append()
      go p.forwardOutPipe(outWrite, node)
   }
   
   return outRead, nil
}

func (p *Process) RequestCommand(otype OutputType) (*os.File, *os.File, error) {
   c := make(chan []byte)
   
   select {
      case p.inProvider <- c:
         // the channel was successfully sent
         
      default:
         return nil, nil, errors.New("The command interface is occupied")
   }
   
   var inRead, inWrite, outRead *os.File
   var err error
   
   inRead, inWrite, err = os.Pipe()
   if (err != nil) {
      return nil, nil, err
   }
   
   outRead, err = p.RequestOutput(otype)
   if (err != nil) {
      inRead.Close()
      inWrite.Close()
      return nil, nil, err
   }

   go p.forwardInPipe(inRead, c)
   
   return inWrite, outRead, nil
}

func (p *Process) SendCommand(otype OutputType, command string) (*os.File, error) {
   return nil, errors.New("not implemented")
}