package worker

import (
	"check-outdated-deps/pkg/npm"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"sync"
	"sync/atomic"
)

type Pool struct {
	sem          chan bool
	wg           sync.WaitGroup
	currentCount *int64
	pkgManager   npm.PackageManager
}

type ProgressCallback func(pkgName, current, latest string, isOutdated bool, count int64)

var maxWorkers = runtime.NumCPU() * 2

func NewPool(pkgManager npm.PackageManager) *Pool {
	return &Pool{
		sem:          make(chan bool, maxWorkers),
		currentCount: new(int64),
		pkgManager:   pkgManager,
	}
}

func (p *Pool) ProcessPackages(packages []npm.Package, progressCallback ProgressCallback) {
	for _, pkg := range packages {
		p.wg.Add(1)

		go func(pkg npm.Package) {
			defer p.wg.Done()

			// Acquire semaphore (blocks if maxWorkers are already running)
			p.sem <- true
			// Release semaphore when done
			defer func() { <-p.sem }()

			// Process the package
			latest := p.processPackage(pkg)

			isOutdated := pkg.Version != latest

			// Update counter and notify progress
			count := atomic.AddInt64(p.currentCount, 1)
			if progressCallback != nil {
				progressCallback(pkg.Name, pkg.Version, latest, isOutdated, count)
			}
		}(pkg)
	}
}

// Wait blocks until all workers have completed
func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) GetCurrentCount() int64 {
	return atomic.LoadInt64(p.currentCount)
}

// Resets the worker pool counters (useful for reusing the pool)
func (p *Pool) Reset() {
	atomic.StoreInt64(p.currentCount, 0)
}

func (p *Pool) processPackage(pkg npm.Package) string {
	packageInfo, err := p.getPackageInfo(pkg)
	if err != nil {
		log.Printf("Error processing package %s: %v", pkg.Name, err)
	}

	if latest, ok := packageInfo["dist-tags"].(map[string]interface{})["latest"].(string); ok {
		return latest
	}

	return ""
}

func (p *Pool) getPackageInfo(pkg npm.Package) (map[string]interface{}, error) {
	var data map[string]interface{}

	var output []byte
	var err error

	switch p.pkgManager {
	case npm.NPM:
		output, err = p.runCommand("npm", "info", pkg.Name, "--json")
	case npm.YARN:
		output, err = p.runCommand("yarn", "npm", "info", pkg.Name, "--json")
	case npm.PNPM:
		output, err = p.runCommand("pnpm", "info", pkg.Name, "--json")
	default:
		return nil, fmt.Errorf("unsupported package manager: %s", p.pkgManager)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get package info for %s: %w", pkg.Name, err)
	}

	err = json.Unmarshal(output, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package info for %s: %w", pkg.Name, err)
	}

	return data, nil
}

func (p *Pool) runCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command failed: %w, output: %s", err, string(output))
	}
	return output, nil
}
