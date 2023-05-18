package backends

import (
	"github.com/Fishwaldo/go-yocto/backends/kde"
	"github.com/Fishwaldo/go-yocto/source"
	"github.com/Fishwaldo/go-yocto/utils"

)


type Backend interface {
	GetName() string
	Init() error
	LoadCache() error
	LoadSource() error
	SearchSource(keyword string) (source []source.RecipeSource, err error)	
	Ready() bool
}

var Backends map[string]Backend

func init() {
	Backends = make(map[string]Backend)
	Backends["kde"] = kde.NewBackend()
}

func Init() (err error) {
	utils.Logger.Trace("Initializing Backends")

	for _, be := range Backends {
		if err := be.Init(); err != nil {
			utils.Logger.Error("Failed to Initialize Backend", utils.Logger.Args("backend", be.GetName(), "error", err))
		}
	}
	return nil
}

func LoadCache() (err error) {
	utils.Logger.Trace("Loading Cache")
	for _, be := range Backends {
		if be.Ready() {
			if err := be.LoadCache(); err != nil {
				utils.Logger.Error("Failed to Load Cache", utils.Logger.Args("backend", be.GetName(), "error", err))
			}
		} else {
			utils.Logger.Trace("LoadCache: Backend not ready", utils.Logger.Args("backend", be.GetName()))
		}
	}
	return nil
}

func LoadSource() (err error) {
	utils.Logger.Trace("Loading Source")
	for _, be := range Backends {
		if be.Ready() {
			if err := be.LoadSource(); err != nil {
				utils.Logger.Error("Failed to Load Source", utils.Logger.Args("backend", be.GetName(), "error", err))
			}
		} else {
			utils.Logger.Trace("LoadSource: Backend not ready", utils.Logger.Args("backend", be.GetName()))
		}
	}
	return nil
}

func SearchSource(be string, keyword string) (sources []source.RecipeSource, err error) {
	utils.Logger.Trace("Searching Source")
	for _, be := range Backends {
		if be.Ready() {
			if source, err := be.SearchSource(keyword); err != nil {
				utils.Logger.Error("Failed to Search Source", utils.Logger.Args("backend", be.GetName(), "error", err))
			} else {
				sources = append(sources, source...)
			}
		} else {
			utils.Logger.Trace("SearchSource: Backend not ready", utils.Logger.Args("backend", be.GetName()))
		}
	}
	return sources, nil
}