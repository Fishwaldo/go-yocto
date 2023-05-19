package recipe

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/Fishwaldo/go-yocto/backends"
	"github.com/Fishwaldo/go-yocto/source"
	"github.com/Fishwaldo/go-yocto/utils"
	"github.com/spf13/viper"
	"github.com/pterm/pterm"
//	"github.com/davecgh/go-spew/spew"
)

var (
	funcs     = template.FuncMap{"join": strings.Join}
	parserecipename = regexp.MustCompile(`^(.*)_(.*)\.bb$`)
	existingRecipes map[string]*source.RecipeSource = make(map[string]*source.RecipeSource)

)


func init() {
}


func CreateRecipe(be string, name string) (error) {
	utils.Logger.Trace("Creating Recipe", utils.Logger.Args("backend", be, "name", name))
	b, ok := backends.Backends[be]
	if !ok {
		utils.Logger.Error("Backend not found", utils.Logger.Args("backend", be))
		return errors.New("backend not found")
	}
	if !b.Ready() {
		utils.Logger.Error("Backend not ready", utils.Logger.Args("backend", be))
		return errors.New("backend not ready")
	}


	layers := viper.GetStringSlice("yocto.layers")
	for _, layer := range layers {
		spinnerInfo, _ := pterm.DefaultSpinner.Start("Scanning Existing Recipe Files in " + layer)
		scanRecipes(layer, 0);
		spinnerInfo.Success()
	}
	s, err := b.GetRecipe(name)
	if err != nil {
		utils.Logger.Error("Failed to get Recipe", utils.Logger.Args("backend", be, "name", name, "error", err))
		return err
	}

	/* if it already exists, bail out */
	if _, ok := existingRecipes[s.Identifier]; ok {
		utils.Logger.Warn("Recipe already exists", utils.Logger.Args("recipe", s.Name))
		return errors.New("recipe already exists")
	}

	/* check out dependancies */
	for _, dep := range s.Depends {
		if _, ok := existingRecipes[dep]; !ok {
			utils.Logger.Warn("Recipe Dependancy not found", utils.Logger.Args("recipe", s.Name, "dependancy", dep))
			return errors.New("recipe dependancy not found")
		}
	}

	if err := writeRecipeFiles(s); err != nil {
		utils.Logger.Error("Failed to write Recipe Files", utils.Logger.Args("error", err))
		return err
	}
	return nil
}

func writeRecipeFiles(s *source.RecipeSource) (error) {
	dir := path.Join(viper.GetString("yocto.layerdirectory"), "recipes-" + s.Section)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			utils.Logger.Error("Failed to create directory", utils.Logger.Args("error", err, "dir", dir))
			return err
		}
	}

	dir = path.Join(dir, s.Identifier)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			utils.Logger.Error("Failed to create directory", utils.Logger.Args("error", err))
			return err
		}
	}

	utils.Logger.Info("Publishing to " + dir)

	maintmpl, err := ioutil.ReadFile("templates/recipe-main.tmpl")
	if err != nil {
		utils.Logger.Error("Failed to read main template", utils.Logger.Args("error", err))
	} else {

		/* create recipe_version.bb file */
		masterTmpl, err := template.New("master").Funcs(funcs).Parse(string(maintmpl))
		if err != nil {
			utils.Logger.Error("Failed to parse template", utils.Logger.Args("error", err))
			return err
		}

		fname := path.Join(dir, fmt.Sprintf("%s_%s.bb", s.Identifier, s.Version))

		if f, err := os.Create(fname); err != nil {
    		utils.Logger.Error("Failed to create file", utils.Logger.Args("error", err))
			return err	
		} else {
			if err := masterTmpl.Execute(f, s); err != nil {
				utils.Logger.Error("Failed to execute template", utils.Logger.Args("error", err))
				return err
			} else {
				utils.Logger.Info("Created " + fname)
			}
			f.Close();
		}

		/* create recipe.inc file */
		maintmpl, err := ioutil.ReadFile("templates/recipe-include.tmpl")
		if err != nil {
			utils.Logger.Error("Failed to read main template", utils.Logger.Args("error", err))
		} else {
			overlayTmpl, err := template.Must(masterTmpl.Clone()).Parse(string(maintmpl))
			if err != nil {
				utils.Logger.Error("Failed to parse template", utils.Logger.Args("error", err))
				return err
			}
			fname = path.Join(dir, fmt.Sprintf("%s.inc", s.Identifier))

			if f, err := os.Create(fname); err != nil {
				utils.Logger.Error("Failed to create file", utils.Logger.Args("error", err))
				return err	
			} else {
				if err := overlayTmpl.Execute(f, s); err != nil {
					utils.Logger.Error("Failed to execute template", utils.Logger.Args("error", err))
					return err
				} else {
					utils.Logger.Info("Created " + fname)
				}
				f.Close();
			}
		}
	}
	return nil
}


func scanRecipes(dir string, level int) (error) {

	fsrecipessections, err := os.ReadDir(dir)
	if err != nil {
		utils.Logger.Error("Failed to read directory", utils.Logger.Args("error", err))
		return err
	}

	for _, file := range fsrecipessections {
		if file.IsDir() {
			if (level == 0) { 
				ok, _ := filepath.Match("recipes-*", file.Name())
				if !ok {
					continue
				}
			}
			if err := scanRecipes(path.Join(dir, file.Name()), level + 1); err != nil {
				utils.Logger.Error("Failed to scan directory", utils.Logger.Args("error", err))
				continue
			}
		} else if file.Type().IsRegular() {
			if strings.EqualFold(path.Ext(file.Name()), ".bb") {
				if err := parseRecipeFile(file, path.Join(dir, file.Name())); err != nil {
					utils.Logger.Error("Failed to parse recipe file", utils.Logger.Args("error", err, "file", path.Join(dir, file.Name())))
					continue
				}
			}
		}
	}
	return nil
}

func parseRecipeFile(file fs.DirEntry, dir string) (error) {
	if parserecipename.Match([]byte(file.Name())) { 
		match := parserecipename.FindSubmatch([]byte(file.Name()))
		s := source.RecipeSource{
			Identifier: string(match[1]),
			Version: string(match[2]),
			Section: strings.TrimPrefix(file.Name(), "-"),
			BackendID: "existing",
			Location: path.Join(dir, file.Name()),
		}
		if _, ok := existingRecipes[s.Identifier]; !ok {
			existingRecipes[s.Identifier] = &s
		}
	}
	return nil
}