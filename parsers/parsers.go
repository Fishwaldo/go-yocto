package parsers

import (
	"errors"
	"io"
	"strings"

	"github.com/Fishwaldo/go-yocto/parsers/appstream"
	"github.com/Fishwaldo/go-yocto/utils"
)



type Parsers interface {
	GetName() string
	Init() error
	Parse(data io.Reader) (metadata map[string]interface{}, err error)
}

var parsers map[string]Parsers

func init() {
	parsers = make(map[string]Parsers, 0)
	as := appstream.NewParser()
	parsers[strings.ToLower(as.GetName())] = as
}

func InitParsers() (err error) {
	utils.Logger.Trace("Initializing Parsers")
	for _, p := range parsers {
		p.Init()
	}
	return nil
}

func GetParser(name string) (p Parsers, err error) {
	p, ok := parsers[strings.ToLower(name)]
	if !ok {
		utils.Logger.Error("Parser Not Found", utils.Logger.Args("name", name))
		return nil, errors.New("Parser Not Found")
	}
	return p, nil
}