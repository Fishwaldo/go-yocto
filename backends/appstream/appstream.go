package appstream

import (
	"github.com/Fishwaldo/go-yocto/source"
	"github.com/Fishwaldo/go-yocto/utils"
)

type AsSummary struct {
	Lang string `xml:"lang,attr"`
	Summary string `xml:",chardata"`
}

type AsDescription struct {
	Lang string `xml:"lang,attr"`
	Description string `xml:",chardata"`
}

type AsReleases struct {
	Version string `xml:"version,attr"`
	Date string `xml:"date,attr"`
}

type AppStream struct {
	Component xml.Name `xml:"component"`
	Name string `xml:"name"`
	Summary []AsSummary `xml:"summary"`
	Description []AsDescription `xml:"description>p"`
	Releases []AsReleases `xml:"releases>release"`
}

type ASProject struct {
	Backend string
	BackendID string
	Project AppStream
}



type AppStreamBE struct {
	ready bool
	projects map[string]ASProject
}

func NewBackend() *AppStreamBE {
	return &AppStreamBE{}
}

func (k *AppStreamBE) GetName() string {
	return "AppStream"
}

func (k *AppStreamBE) Init() error {
	utils.Logger.Trace("Initializing AppStream Backend")
	k.ready = true
	return nil
}

func (k *AppStreamBE) LoadCache() error {
	utils.Logger.Trace("Loading AppStream Cache")
	return nil
}

func (k *AppStreamBE) LoadSource() error {
	utils.Logger.Trace("Loading AppStream Source")
	return nil
}

func (k *AppStreamBE) RefreshSource() error {
	utils.Logger.Trace("Refreshing AppStream Source")
	return nil
}

func (k *AppStreamBE) SearchSource(keyword string) (source []source.RecipeSource, err error) {
	utils.Logger.Trace("Searching AppStream Source", utils.Logger.Args("keyword", keyword))
	return nil, nil
}

func (k *AppStreamBE) Ready() bool {
	return k.ready
}
