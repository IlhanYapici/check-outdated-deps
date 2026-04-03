package worker

import (
	"log"
	"runtime"
	"sync"
	"sync/atomic"

	"check-outdated-deps/internal/npm"
)

type Pool struct {
	sem          chan bool
	wg           sync.WaitGroup
	currentCount *int64
}

type ProgressCallback func(pkgName, current, latest string, isOutdated bool, count int64)

var maxWorkers = runtime.NumCPU() * 2

func NewPool() *Pool {
	return &Pool{
		sem:          make(chan bool, maxWorkers),
		currentCount: new(int64),
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

func (p *Pool) processPackage(pkg npm.Package) string {
	packageInfo, err := npm.GetPackageMetadata(pkg.Name)
	if err != nil {
		log.Printf("Error processing package %s: %v", pkg.Name, err)
	}

	if latest, ok := packageInfo["dist-tags"].(map[string]interface{})["latest"].(string); ok {
		return latest
	}

	return ""
}
