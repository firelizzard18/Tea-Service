package main

import (
   "io"
   "os"
   "log"
   "time"
   "os/exec"
   "errors"
   "sync"
)

type Process struct {
   cmd *exec.Cmd
   exported *DBusServer
   started *sync.WaitGroup

   inProvider *Provider
   outListeners *ListenerList
   errListeners *ListenerList
   
   Description string
}

func StartProcess(name string, desc string, bus string, arg ...string) (p *Process, err error) {
   p = new(Process)
   p.cmd = exec.Command(name, arg...)
   p.started = new(sync.WaitGroup)
   p.Description = desc

   p.started.Add(1)

   inPipe, err := p.cmd.StdinPipe()
   if err != nil {
      return nil, errors.New("Could not open process stdin: " + err.Error())
   }

   outPipe, err := p.cmd.StdoutPipe()
   if err != nil {
      return nil, errors.New("Could not open process stdout: " + err.Error())
   }

   errPipe, err := p.cmd.StderrPipe()
   if err != nil {
      return nil, errors.New("Could not open process stderr: " + err.Error())
   }

   p.exported, err = ExportToDBus(p, bus)
   if err != nil {
      return nil, errors.New("Could not open dbus: " + err.Error())
   }
   
   p.inProvider = New()
   p.outListeners = NewListenerList()
   p.errListeners = NewListenerList()

   go p.runInputProvider(inPipe, p.inProvider)
   go p.runOutputListener(outPipe, p.outListeners)
   go p.runOutputListener(errPipe, p.errListeners)
   
   return
}

func (p *Process) runInputProvider(pipe io.Writer, provider *Provider) {
   p.started.Wait()
   p.Start(pipe)
}

func (p *Process) runOutputListener(pipe io.Reader, list *ListenerList) {
   p.started.Wait()
   data := make([]byte, 256)
   for {
      n, err := pipe.Read(data)
      if err != nil {
         panic(err)
      }

      for node := range list.Traverse() {
         node.Write(data[:n])
      }
   }
}

func (p *Process) Start() (err error) {
   err = p.cmd.Start()
   if err == nil {
      p.started.Done()
   }
   return
}

func (p *Process) Wait() error {
   return p.cmd.Wait()
}

func (p *Process) forwardFileSource(r *os.File, c chan []byte) {
   _ = "breakpoint"
   var n int
   var err error
   
   data := make([]byte, 256)
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
   var writeCount int
   var err error

outer:
   for data := range node.Read() {
      totalCount := len(data)
      offset := 0
      for offset < totalCount {
         writeCount, err = w.Write(data[offset:totalCount])
         if err != nil {
            break outer
         }
         offset += writeCount
      }
   }
   
   log.Print(err)
   node.Remove()
   
   err = w.Close()
   if err != nil {
      log.Print(err)
   }
   return
}

func (p *Process) ConnectInput(source *os.File) error {
   c, ok := p.inProvider.Write()
   if !ok {
         return errors.New("The command interface is occupied")
   }
   
   go p.forwardFileSource(source, c)
   return nil
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
   if otype == None || otype >= Invalid {
      return nil, errors.New("Invalid output type")
   }
   
   outRead, outWrite, err := os.Pipe()
   if err != nil {
      return nil, err
   }
   go func () {
      time.Sleep(time.Duration(1) * time.Second)
      outWrite.Close()
   }()

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
   go func () {
      time.Sleep(time.Duration(1) * time.Second)
      inRead.Close()
   }()
   
   if otype != None {
      outRead, err = p.RequestOutput(otype)
      if err != nil {
         goto fail2
      }
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
   var outRead *os.File
   var err error

   c, ok := p.inProvider.Write()
   if !ok {
         return nil, errors.New("The command interface is occupied")
   }

   if otype != None {
      outRead, err = p.RequestOutput(otype)
      if err != nil {
         return nil, err
      }
   }

   go func () {
      c <- []byte(command + "\n")
      close(c)
   }()

   return outRead, err
}