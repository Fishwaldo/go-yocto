package backends

import (
	"github.com/Fishwaldo/go-yocto/backends/kde"
	"github.com/Fishwaldo/go-yocto/utils"
)


type Backend interface {
	GetName() string
	Init() error
	LoadCache() error
	LoadSource() error
	RefreshSource() error
}

var Backends map[string]Backend

func init() {
	Backends = make(map[string]Backend)
	Backends["kde"] = kde.NewBackend()
}

