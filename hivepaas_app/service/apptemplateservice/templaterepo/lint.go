package templaterepo

import (
	"fmt"
	"path"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
)

// lintVolumeID and lintSecret stand in for what a user supplies, so a template
// can be rendered without one.
const (
	lintVolumeID = "lint-volume"
	lintSecret   = "lint-secret-value"
)

// Lint checks what Load could read: each template on its own, against the
// vocabularies and its icon, and by rendering every version and variant with
// default parameters - the same render HivePaaS runs.
func Lint(repo *Repo) []Problem {
	problems := lintVocabularies(repo)
	categories := categoryRefs(repo.Categories.Categories)
	tags := map[string]bool{}
	for _, tag := range repo.Tags.Tags {
		tags[tag.ID] = true
	}

	for _, file := range repo.Templates {
		tmpl := file.Template
		report := func(format string, args ...any) {
			problems = append(problems, Problem{Path: file.Path, Message: fmt.Sprintf(format, args...)})
		}

		name := strings.TrimSuffix(path.Base(file.Path), ".yaml")
		if err := tmpl.Validate(name); err != nil {
			report("%s", ErrorText(err))
			continue
		}
		for _, category := range tmpl.Metadata.Categories {
			if !categories[category] {
				report("category %q is not in %s", category, CategoriesFile)
			}
		}
		for _, tag := range tmpl.Metadata.Tags {
			if !tags[tag] {
				report("tag %q is not in %s", tag, TagsFile)
			}
		}
		if repo.Icons[tmpl.Metadata.Icon] == nil {
			report("icon %q is missing", tmpl.Metadata.Icon)
		}
		if err := templaterender.ValidateDefaults(tmpl.Parameters); err != nil {
			report("%s", ErrorText(err))
			continue
		}
		problems = append(problems, lintRenders(file)...)
	}
	return problems
}

func lintVocabularies(repo *Repo) []Problem {
	var problems []Problem
	parents := map[string]bool{}
	for _, category := range repo.Categories.Categories {
		if parents[category.ID] {
			problems = append(problems, Problem{Path: CategoriesFile,
				Message: fmt.Sprintf("category %q is declared twice", category.ID)})
		}
		parents[category.ID] = true
		if len(category.Children) == 0 {
			problems = append(problems, Problem{Path: CategoriesFile,
				Message: fmt.Sprintf("category %q has no children to put templates in", category.ID)})
		}
	}
	tags := map[string]bool{}
	for _, tag := range repo.Tags.Tags {
		if tags[tag.ID] {
			problems = append(problems, Problem{Path: TagsFile, Message: fmt.Sprintf("tag %q is declared twice", tag.ID)})
		}
		tags[tag.ID] = true
	}
	return problems
}

func categoryRefs(categories []*templatemodel.Category) map[string]bool {
	refs := map[string]bool{}
	for _, parent := range categories {
		for _, child := range parent.Children {
			refs[parent.ID+"/"+child.ID] = true
		}
	}
	return refs
}

func lintRenders(file *TemplateFile) []Problem {
	tmpl := file.Template
	variants := []string{""}
	if len(tmpl.Variants) > 0 {
		variants = variants[:0]
		for _, variant := range tmpl.Variants {
			variants = append(variants, variant.Name)
		}
	}

	var problems []Problem
	for _, version := range tmpl.Versions {
		for _, variant := range variants {
			if variant != "" && version.ImageFor(variant) == "" {
				continue
			}
			_, err := templaterender.Render(&templaterender.Request{
				Template:        tmpl,
				Version:         version.Name,
				Variant:         variant,
				Params:          lintParams(tmpl),
				AllowDeprecated: true,
			})
			if err != nil {
				label := "version " + version.Name
				if variant != "" {
					label += ", variant " + variant
				}
				problems = append(problems, Problem{Path: file.Path, Message: label + ": " + ErrorText(err)})
			}
		}
	}
	return problems
}

// lintParams fills in what defaults cannot: a volume, and a secret that has no
// generate. Everything else renders from its default, and a required parameter
// without one fails the render - which is the point.
func lintParams(tmpl *templatemodel.Template) map[string]any {
	params := map[string]any{}
	for _, param := range tmpl.Parameters {
		switch param.Type {
		case templatemodel.ParamTypeVolume:
			params[param.Name] = lintVolumeID
		case templatemodel.ParamTypeSecret:
			if param.Generate == nil {
				params[param.Name] = lintSecret
			}
		case templatemodel.ParamTypeString, templatemodel.ParamTypeInt, templatemodel.ParamTypeSize,
			templatemodel.ParamTypeBool, templatemodel.ParamTypeSelect:
		}
	}
	return params
}
