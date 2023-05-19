package source

type RecipeSource struct {
	Name string
	Identifier string
	Description string
	Summary string
	Version string
	Url string
	Section string
	BackendID string
	Inherits []string
	Depends []string
	SrcURI string
	SrcSHA256 string
	Licenses []string
	Location string
}

