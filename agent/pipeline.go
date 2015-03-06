package agent

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/vektra/cypress"
)

func NewSpooolFile(dir string) *SpoolFile {
	sf := new(SpoolFile)
	os.MkdirAll(dir, 0755)

	if err := sf.Start(dir); err != nil {
		panic(err)
	}

	return sf
}

func ParseSource(s string, r cypress.Reciever) Source {
	if s == "local" {
		return LocalCollector(r)
	} else {
		uri, err := url.Parse(s)

		if err == nil {
			switch uri.Scheme {
			case "redis":
				ri := &RedisInput{}
				ri.Init(uri.Host, uri.Path, r)
				return ri
			default:
				panic("Unknown source")
			}
		} else {
			panic("Unknown source")
		}
	}
}

func ParseSink(s string) (cypress.Reciever, error) {
	switch s {
	case "spool":
		return NewSpooolFile(DefaultSpoolDir), nil
	default:
		if s[0:6] == "spool:" {
			return NewSpooolFile(s[7:]), nil
		} else {
			uri, err := url.Parse(s)

			if err == nil {
				switch uri.Scheme {
				case "redis":
					ro := &RedisOutput{}
					ro.Start(uri.Host, uri.Path)
					return ro, nil
				default:
					return nil, errors.New(fmt.Sprintf("Invalid uri: %s", s))
				}
			} else {
				return nil, errors.New(fmt.Sprintf("Invalid sink: %s", s))
			}
		}
	}
}

type Pipeline struct {
	Recievers *ManyReciever
	Sources   []Source
}

type Latch chan error

func (p *Pipeline) Start() error {
	gates := []Latch{}

	for _, s := range p.Sources {
		l := make(Latch)

		go func() {
			l <- s.Start()
		}()

		gates = append(gates, l)
	}

	var outer error

	for _, l := range gates {
		err := <-l

		if err != nil {
			outer = err
		}
	}

	return outer
}

func (p *Pipeline) Close() {
	for _, s := range p.Sources {
		s.Close()
	}
}

func MakePipeline(srcs, sinks string) (*Pipeline, error) {
	recvs := []cypress.Reciever{}

	for _, s := range strings.Split(sinks, ",") {
		r, err := ParseSink(s)
		if err != nil {
			return nil, err
		}

		recvs = append(recvs, r)
	}

	pi := &Pipeline{Recievers: ManyRecievers(recvs...)}

	for _, s := range strings.Split(srcs, ",") {
		src := ParseSource(s, pi.Recievers)

		pi.Sources = append(pi.Sources, src)
	}

	return pi, nil
}
