package main

import (
   "io"
   "os"
   "log"
   "os/exec"
   "errors"
   "sync"
)

type Process struct {
   cmd *exec.Cmd
   exported *DBusServer
   started *sync.WaitGroup

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

func StartProcess(name string, desc string, bus string, arg ...string) (p *Process, err error) {
   p = new(Process)
   p.cmd = exec.Command(name, arg...)
   p.started = new(sync.WaitGroup)
   p.Description = desc

   p.started.Add(1)

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

   p.exported, err = ExportToDBus(p, bus)
   if err != nil {
      return nil, errors.New("Could not open dbus: " + err.Error())
   }
   
   p.inProvider = make(chan chan []byte, 1)
   p.outListeners = NewListenerList()
   p.errListeners = NewListenerList()

   go p.forwardCmdReader(p.outPipe, p.outListeners)
   go p.forwardCmdReader(p.errPipe, p.errListeners)

   // do something about input
   
   return
}

func (p *Process) Run() (err error) {
   err = p.cmd.Run()
   if err == nil {
      p.started.Done()
   }
   return
}

func (p *Process) Start() (err error) {
   err = p.cmd.Start()
   if err == nil {
      p.started.Done()
   }
   return
}

func (p *Process) forwardCmdReader(pipe io.Reader, list *ListenerList) {
   p.started.Wait()
   data := make([]byte, 0, 256)
   for {
      n, err := pipe.Read(data)
      if err != nil {
         panic(err)
      }

      for node := range list.Traverse() {
         node.sink <- data[:n]
      }
   }
}

func (p *Process) forwardFileSource(r *os.File, c chan []byte) {
   var n int
   var err error
   
   data := make([]byte, 0, 256)
   for {
      n, err = r.Read(data)
      if err != nil {
         break
      }
      
      c <- data[:n]
   }
   
   log.Print(err)
   close(c)
   
   err = r.Close()
   if err != nil {
      log.Print(err)
   }
   return
}

func (p *Process) forwardFileSink(w *os.File, node *ListenerNode) {
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
   
   err = w.Close()
   if err != nil {
      log.Print(err)
   }
   return
}

func (p *Process) ConnectInput(source *os.File) error {
   c := make(chan []byte)
   
   select {
      case p.inProvider <- c:
         // the channel was successfully sent
         go p.forwardFileSource(source, c)
         return nil
         
      default:
         return errors.New("The command interface is occupied")
   }
}

func (p *Process) ConnectOutput(sink *os.File) {
   node := p.outListeners.Append()
   go p.forwardFileSink(sink, node)
}

func (p *Process) ConnectError(sink *os.File) {
   node := p.errListeners.Append()
   go p.forwardFileSink(sink, node)
}

func (p *Process) RequestOutput(otype OutputType) (*os.File, error) {
   if (otype >= Invalid) {
      return nil, errors.New("Invalid output type")
   }
   
   outRead, outWrite, err := os.Pipe()
   if err != nil {
      return nil, err
   }
   
   if otype == Output || otype == OutAndErr {
      p.ConnectOutput(outWrite)
   }
   
   if otype == Error || otype == OutAndErr {
      p.ConnectError(outWrite)
   }
   
   return outRead, nil
}

func (p *Process) RequestCommand(otype OutputType) (*os.File, *os.File, error) {
   var inRead, inWrite, outRead *os.File
   var err error
   
   inRead, inWrite, err = os.Pipe()
   if err != nil {
      goto fail1
   }
   
   outRead, err = p.RequestOutput(otype)
   if err != nil {
      goto fail2
   }
   
   err = p.ConnectInput(inRead)
   if err != nil {
      goto fail2
   }
   
   return inWrite, outRead, nil
   
fail2:
   inRead.Close()
   inWrite.Close()
fail1:
   return nil, nil, err
}

func (p *Process) SendCommand(otype OutputType, command string) (*os.File, error) {
   return nil, errors.New("not implemented")
}