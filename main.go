package main

import (
	"check-outdated-deps/internal/config"
	"check-outdated-deps/internal/parser"
	"check-outdated-deps/internal/worker"
	"check-outdated-deps/pkg/npm"
	"check-outdated-deps/pkg/version"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"sync/atomic"

	"github.com/fatih/color"
	"github.com/pterm/pterm"
	"github.com/rodaine/table"
)

var outdated int64
var outdatedDev int64

func main() {
	// Flags
	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Version: %s\nCommit: %s\nBuilt: %s\n", version.Version, version.GitCommit, version.BuildTime)
		return
	}

	var dependencies npm.Dependencies
	var devDependencies npm.DevDependencies

	parsedFile, err := config.LoadPackageJson("package.json")
	if err != nil {
		errMsg := fmt.Sprintf("error while loading package.json: %s", err)
		log.Fatal(errMsg)
	}

	pkgManager, err := config.DeterminePackageManager("package.json", parsedFile)

	if err != nil {
		log.Fatal(err)
	}

	for pkg, version := range parsedFile.Dependencies {
		dependencies = append(dependencies, npm.Package{Name: pkg, Version: parser.SanitizeVersion(version)})
	}

	for pkg, version := range parsedFile.DevDependencies {
		devDependencies = append(devDependencies, npm.Package{Name: pkg, Version: parser.SanitizeVersion(version)})
	}

	pkgManagerExists := config.CheckPkgManagerExists(string(pkgManager))

	if !pkgManagerExists {
		errMessage := fmt.Sprintf("package manager '%s' not found", pkgManager)
		log.Fatal(errMessage)
	}

	pool := worker.NewPool(pkgManager)

	// TEST TABLE
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()

	depsTable := table.New("DEPENDENCY", "CURRENT", "LATEST")
	devDepsTable := table.New("DEV DEPENDENCY", "CURRENT", "LATEST")
	depsTable.WithHeaderFormatter(headerFmt).WithPadding(2)
	devDepsTable.WithHeaderFormatter(headerFmt).WithPadding(2)
	depRows := [][]string{}
	devDepRows := [][]string{}

	// Progress bar
	fmt.Println()
	p, _ := pterm.DefaultProgressbar.WithTotal(len(dependencies) + len(devDependencies)).WithTitle("Checkin latest versions").Start()

	getProgressCallback := func(isDevDependency bool) func(pkgName, current, latest string, isOutdated bool, count int64) {
		return func(pkgName, current, latest string, isOutdated bool, count int64) {
			p.Increment()
			formattedLatest := latest

			if isOutdated {

				if formattedVersion, err := parser.FormatVersionComparison(current, latest); err == nil {
					formattedLatest = formattedVersion
				}

				if isDevDependency {
					atomic.AddInt64(&outdatedDev, 1)
					devDepRows = append(devDepRows, []string{pkgName, current, formattedLatest})
				} else {
					atomic.AddInt64(&outdated, 1)
					depRows = append(depRows, []string{pkgName, current, formattedLatest})
				}
			}
		}
	}

	pool.ProcessPackages(dependencies, getProgressCallback(false))
	pool.ProcessPackages(devDependencies, getProgressCallback(true))

	// Wait for goroutines to finish
	pool.Wait()

	sort.Slice(depRows, func(i, j int) bool {
		return depRows[i][0] < depRows[j][0]
	})
	sort.Slice(devDepRows, func(i, j int) bool {
		return devDepRows[i][0] < devDepRows[j][0]
	})

	depsTable.SetRows(depRows)
	devDepsTable.SetRows(devDepRows)

	if outdated > 0 {
		fmt.Println()
		depsTable.Print()
	}
	if outdatedDev > 0 {
		fmt.Println()
		devDepsTable.Print()
	}

	if outdated > 0 || outdatedDev > 0 {
		fmt.Printf("\n⚠️ %d outdated packages out of %d packages.", outdated+outdatedDev, len(dependencies)+len(devDependencies))
		os.Exit(1)
	} else {
		fmt.Printf("\n✅ All %d packages are up-to-date.", len(dependencies)+len(devDependencies))
	}
}
