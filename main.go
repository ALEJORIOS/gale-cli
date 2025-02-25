package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	// "time"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No commands found. For help -h")
		log.Println(os.Args[0])
		return
	}

	args := os.Args[0:]
	log.Println("args: ", args)

	switch {
	case args[1] == "help" || args[1] == "-h":
		help()
	case args[1] == "new" || args[1] == "-n":
		createNewProject(args[2])
	case args[1] == "run":
		run()
	default:
		fmt.Println("Eror: unknown command")
	}
}

func createNewProject(name string) error {
	if name == "" {
		return fmt.Errorf("the project name is needed")
	}
	fmt.Println("Creating a new gale project...")
	return nil
	// Aquí podríamos copiar una estructura base de proyecto
}

func run() {
	fmt.Println()
	fmt.Println("Executing Go Gale with hot reloading...")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	var cmd *exec.Cmd

	restart := func() {
		if cmd != nil {
			cmd.Process.Kill()
		}
		cmd = exec.Command("go", "run", ".")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Start()
	}

	restart()

	var lastEventTime time.Time
	var lastEventPath string
	var mu sync.Mutex

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					mu.Lock()
					now := time.Now()

					// Si el mismo archivo se modificó en menos de 500ms, ignorarlo
					if event.Name == lastEventPath && now.Sub(lastEventTime) < 500*time.Millisecond {
						mu.Unlock()
						continue
					}

					lastEventTime = now
					lastEventPath = event.Name
					mu.Unlock()
					fmt.Println("\x1b[35mreloading...\x1b[0m")
					restart()
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error in watcher:", err)
			}
		}
	}()

	path := "."
	// Agregar archivos a observar
	if len(os.Args) > 2 && (os.Args[2] == "--path" || os.Args[2] == "-p") {
		path = os.Args[3]
	}
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

func help() {
	fmt.Println()
	fmt.Println("\x1b[1;34mGale - A tool for developing web applications with Go!\x1b[0m")
	fmt.Println("\x1b[1mCommands:\x1b[0m")
	fmt.Println("gale help \t\t\tnew show help")
	// fmt.Println("gale new <project name> \tnew project")
	fmt.Println("gale run \t\t\trun the current project with hot-reloading")
	fmt.Println()
}
