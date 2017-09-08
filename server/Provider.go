package server

import (
	"io"
	"log"
)

type Provider struct {
	channel chan chan []byte
	// TODO use CompareAndSwapPointer
	mutex chan empty
}

func New() *Provider {
	p := new(Provider)

	p.channel = make(chan chan []byte, 1)
	p.mutex = make(chan empty, 1)

	return p
}

func (p *Provider) Start(pipe io.Writer) {
	select {
	case p.mutex <- empty{}:
		// things worked, cool

	default:
		log.Panic("Somehow the mutex got backed up")
		return
	}

	go func() {
		for dataChan := range p.channel {
			for data := range dataChan {
				totalCount := len(data)

				for offset := 0; offset < totalCount; {
					count, err := pipe.Write(data[offset:totalCount])
					if err != nil {
						panic(err)
					}

					offset += count
				}
			}

			select {
			case p.mutex <- empty{}:
				// things worked, cool

			default:
				log.Panic("Somehow the mutex got backed up")
				return
			}
		}
	}()
}

func (p *Provider) Stop() {
	close(p.mutex)
	close(p.channel)
}

func (p *Provider) Write() (dataChan chan []byte, ok bool) {
	select {
	case <-p.mutex:
		// channel is available

	default:
		ok = false
		return
	}

	dataChan = make(chan []byte)
	select {
	case p.channel <- dataChan:
		// start routine is working
		ok = true
		return

	default:
		log.Panic("Somehow the data channel channel has gotten backed up")
		ok = false
		return
	}
}
