package main

import (
	"go/build"
	"sort"
)

type context struct {
	soFar map[string]bool
	ctx   build.Context
}

func (c *context) find(name, dir string) (err error) {
	if name == "C" {
		return nil
	}
	var pkg *build.Package
	pkg, ok := cache[name]
	if !ok {
		pkg, err = c.ctx.Import(name, dir, 0)
		if err != nil {
			return err
		}
	}
	cache[name] = pkg
	if pkg.Goroot {
		return nil
	}

	if name != "." {
		c.soFar[pkg.ImportPath] = true
	}
	imports := pkg.Imports
	for _, imp := range imports {
		if !c.soFar[imp] {
			if err := c.find(imp, pkg.Dir); err != nil {
				return err
			}
		}
	}
	return nil
}

func findDeps(name, dir string) ([]string, error) {
	ctx := build.Default

	c := &context{
		soFar: make(map[string]bool),
		ctx:   ctx,
	}
	if err := c.find(name, dir); err != nil {
		return nil, err
	}
	deps := make([]string, 0, len(c.soFar))
	for p := range c.soFar {
		if p != name {
			deps = append(deps, p)
		}
	}
	sort.Strings(deps)
	return deps, nil
}
