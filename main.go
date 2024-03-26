package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"time"

	"github.com/radovskyb/watcher"
)

type Config struct {
	Output        string   `json:"output"`
	ExcludedPaths []string `json:"excludedPaths"`
	BuildArgs     []string `json:"buildArgs"`
	RunArgs       []string `json:"runArgs"`
	TempDir       string   `json:"tmp_dir"`
}

func main() {
	colorGreen := "\033[32m"
	colorRed := "\033[31m"
	cyanColor := "\033[36m"
	colorReset := "\033[0m"

	logger := log.New(os.Stdout, "[GOIR] ", log.LstdFlags)
	w := watcher.New()
	var config Config

	var outputBinary string = "./tmp"

	var runCmd *exec.Cmd

	var isLinux = runtime.GOOS != "windows"

	if confFile, _ := os.ReadFile("./goir.json"); confFile != nil {
		json.Unmarshal(confFile, &config)
	}
	if config.TempDir != "" {
		outputBinary = "./" + config.TempDir + "/"
		err := os.Mkdir("./"+config.TempDir, 0750)
		if err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}
	}

	if config.Output == "" {
		if !isLinux {
			outputBinary += "/main.exe"
		} else {
			outputBinary += "/main"
		}
	} else {
		if !isLinux {
			outputBinary += config.Output + ".exe"
		} else {
			outputBinary += config.Output
		}
	}

	args := os.Args
	if len(args) > 1 {
		config.RunArgs = append(config.RunArgs, args[1:]...)
	}

	config.BuildArgs = append([]string{"build", "-o", outputBinary}, config.BuildArgs...)

	w.SetMaxEvents(1)

	if len(config.ExcludedPaths) > 0 {
		w.Ignore(config.ExcludedPaths...)
	}

	w.IgnoreHiddenFiles(true)

	fmt.Printf("%s%s%s",

		cyanColor,

		"\n ██████╗  █████╗ ██╗██████╗ \n"+
			"██╔════╝ ██╔══██╗██║██╔══██╗\n"+
			"██║  ██╗ ██║  ██║██║██████╔╝\n"+
			"██║  ╚██╗██║  ██║██║██╔══██╗\n"+
			"╚██████╔╝╚█████╔╝██║██║  ██║\n"+
			" ╚═════╝  ╚════╝ ╚═╝╚═╝  ╚═╝ written in go v0.0.6-beta \n", colorReset)

	r := regexp.MustCompile(".go$")
	w.AddFilterHook(watcher.RegexFilterHook(r, false))

	if err := w.AddRecursive("."); err != nil {
		log.Fatalln(err)
	}

	if open, _ := os.Open(outputBinary); open != nil {
		open.Close()
		if err := os.RemoveAll(outputBinary); err != nil {
			logger.Printf("%sError removing: %v%s\n", colorRed, err, colorReset)
		}
	}

	pids := make([]int, 0)

	go func() {
		for {
			select {
			case event := <-w.Event:

				if !event.IsDir() {

					if event.Path != "-" && event.FileInfo.Name() != "" {
						logger.Printf("%s%s has changed%s\n", cyanColor, event.FileInfo.Name(), colorReset)
					}

					if isLinux {
						if runCmd != nil && runCmd.Process != nil {
							if err := runCmd.Process.Signal(os.Kill); err != nil {
								logger.Printf("%sError killing previous process %v%s\n", colorRed, err, colorReset)
								continue
							}
						}
					} else {
						for _, pid := range pids {
							kill := exec.Command("taskkill", "/pid", fmt.Sprint(pid), "/T", "/F")
							if err := kill.Run(); err != nil {
								logger.Printf("%sError killing previous process %v%s\n", colorRed, err, colorReset)
								continue
							}
						}
					}

					logger.Printf("%sBuilding...%s\n", colorGreen, colorReset)

					buildCmd := exec.Command("go", config.BuildArgs...)
					output, err := buildCmd.CombinedOutput()
					if err != nil {
						logger.Printf("%sError building: %v%s\n", colorRed, err, colorReset)
						logger.Printf("%s %s %s\n", colorRed, string(output), colorReset)
						continue
					}

					runCmd = exec.Command(outputBinary, config.RunArgs...)

					runCmd.Stdout = os.Stdout
					runCmd.Stderr = os.Stderr
					if err := runCmd.Start(); err != nil {
						logger.Printf("%sError running: %v%s\n", colorRed, err, colorReset)
						continue
					}

					if !isLinux {
						pids = append(pids, runCmd.Process.Pid)
					}

					logger.Printf("%sStarted serving...%s\n\n", colorGreen, colorReset)
				}

			case err := <-w.Error:
				log.Fatalln(err)

			case <-w.Closed:
				return
			}
		}
	}()

	go func() {
		for _, file := range w.WatchedFiles() {
			logger.Printf("Watching %s%s%s", cyanColor, file.Name(), colorReset)
		}
	}()

	go func() {
		w.Wait()
		w.TriggerEvent(watcher.Create, nil)
	}()

	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
	}
}
