package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type CLI struct {
	Version          kong.VersionFlag `help:"Show version information."`
	DryRun           bool             `help:"[SAFE MODE] List duplicate files without making changes. Always test with this first!"`
	Delete           bool             `help:"⚠️  WARNING: Permanently delete duplicate files. USE AT YOUR OWN RISK. No warranty provided."`
	Inverse          bool             `help:"Inverse deletion, keeping only the newest file and deleting older ones."`
	InverseAndRename bool             `name:"inverse-and-rename" help:"Inverse deletion and rename, keeping only the newest file and renaming it."`
	Out              string           `name:"out" short:"o" help:"Output file for results." type:"path"`
	Path             []string         `arg:"" name:"path" help:"Path(s) to search for duplicates." type:"path"`
	Regex            string           `name:"regex" help:"⚠️  Custom regex for finding duplicates. USE AT YOUR OWN RISK - test with --dry-run first!" default:"(.+)\\s\\((\\d+)\\)\\.(pdf|mobi|mp4|epub|wav|mp3)$"`
}

var cli CLI

type Context struct {
	*kong.Context
}

func (c *CLI) Run(_ *Context) error {
	if len(c.Path) == 0 {
		return fmt.Errorf("at least one path must be specified")
	}
	re, err := regexp.Compile(c.Regex)
	if err != nil {
		return fmt.Errorf("invalid regex: %w", err)
	}

	// Map to store original files and their duplicates
	files := make(map[string][]string)

	for _, p := range c.Path {
		err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				matches := re.FindStringSubmatch(filepath.Base(path))
				if len(matches) > 0 {
					// Compute the original file's full path
					baseName := matches[1] + "." + matches[3]
					originalPath := filepath.Join(filepath.Dir(path), baseName)
					files[originalPath] = append(files[originalPath], path)
				}
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("error walking path %s: %v", p, err)
		}
	}

	var results []string

	for original, duplicates := range files {
		if len(duplicates) == 0 {
			continue
		}

		// Check if the original file actually exists
		if _, err := os.Stat(original); os.IsNotExist(err) {
			continue
		}

		if c.DryRun {
			results = append(results, fmt.Sprintf("Original: %s", original))
			for _, d := range duplicates {
				results = append(results, fmt.Sprintf("  - Duplicate: %s", d))
			}
			continue
		}

		if c.Delete {
			if c.Inverse || c.InverseAndRename {
				// Keep the newest file
				sort.Slice(duplicates, func(i, j int) bool {
					infoI, _ := os.Stat(duplicates[i])
					infoJ, _ := os.Stat(duplicates[j])
					return infoI.ModTime().After(infoJ.ModTime())
				})

				newest := duplicates[0]
				toDelete := duplicates[1:]
				toDelete = append(toDelete, original)

				for _, f := range toDelete {
					err := os.Remove(f)
					if err != nil {
						results = append(results, fmt.Sprintf("Failed to delete %s: %v", f, err))
					} else {
						results = append(results, fmt.Sprintf("Deleted %s", f))
					}
				}

				if c.InverseAndRename {
					// The original has been deleted, so we can rename the newest to the original's name
					err := os.Rename(newest, original)
					if err != nil {
						results = append(results, fmt.Sprintf("Failed to rename %s to %s: %v", newest, original, err))
					} else {
						results = append(results, fmt.Sprintf("Renamed %s to %s", newest, original))
					}
				} else {
					results = append(results, fmt.Sprintf("Kept newest file: %s", newest))
				}

			} else {
				// Delete all duplicates
				for _, d := range duplicates {
					err := os.Remove(d)
					if err != nil {
						results = append(results, fmt.Sprintf("Failed to delete %s: %v", d, err))
					} else {
						results = append(results, fmt.Sprintf("Deleted %s", d))
					}
				}
			}
		}
	}

	output := strings.Join(results, "\n")

	if c.Out != "" {
		return outputResults(c.Out, output)
	} else if c.Delete {
		return outputResults("results.txt", output)
	}

	fmt.Println(output)
	return nil
}

func outputResults(filename string, results string) error {
	err := os.WriteFile(filename, []byte(results), 0644)
	if err != nil {
		return fmt.Errorf("failed to write results to %s: %v", filename, err)
	}
	fmt.Printf("Results written to %s\n", filename)
	return nil
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("ohman"),
		kong.Description(`⚠️  WARNING: This tool deletes files permanently. USE AT YOUR OWN RISK.

This software is provided "as-is", without warranty of any kind.

Always backup your files and test with --dryrun first.
`),
		kong.UsageOnError(),
		kong.Vars{
			"version": version,
			"commit":  commit,
			"date":    date,
		},
	)
	err := ctx.Run(&Context{Context: ctx})
	ctx.FatalIfErrorf(err)
}
