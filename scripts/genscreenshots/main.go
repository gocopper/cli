package main

import (
	"flag"
	"log"
	"sync"
)

func main() {
	const (
		HTTPPort = 5001
		VitePort = 6001
	)

	var (
		cliPkgPath        string
		screenshotsOutDir string

		wg sync.WaitGroup
	)

	flag.StringVar(&cliPkgPath, "pkg", "", "Path to copper cli's main pkg")
	flag.StringVar(&screenshotsOutDir, "out", "screenshots", "Path to a directory where screenshots will be saved")

	flag.Parse()

	var stacks = []Stack{
		{Name: "go"},
		{Name: "go:tailwind", RunNPM: true},
		{Name: "vite:react", RunNPM: true, HasVite: true},
		{Name: "vite:react:tailwind", RunNPM: true, HasVite: true},
		{Name: "none"},
	}

	wg.Add(len(stacks))

	for i := range stacks {
		var (
			stack    = stacks[i]
			httpPort = HTTPPort + i
			vitePort = VitePort + i
		)

		go func() {
			defer wg.Done()

			err := runTask(&ScreenGrabber{
				Stack:             stack,
				CLIPkgPath:        cliPkgPath,
				HTTPPort:          httpPort,
				VitePort:          vitePort,
				ScreenshotsOutDir: screenshotsOutDir,
			})
			if err != nil {
				log.Printf("Failed to grab screenshot for stack %s because \n%+v\n", stack.Name, err)
			}
		}()
	}

	wg.Wait()
}
