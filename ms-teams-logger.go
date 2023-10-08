package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/intevel/ms-teams-status-log/logger"
)

var statusStrings = []string{"Added Available", "Added Busy", "Added DoNotDisturb", "Added BeRightBack", "Added OnThePhone", "Added Presenting", "Added InAMeeting", "Added Away", "Offline"}

func main() {
	filePath := flag.String("file", "", "Path to the text file to watch for changes")
	outputPath := flag.String("output", "", "Path to the text file to write the log to")
	flag.Parse()

	if *filePath == "" {
		fmt.Println("Usage: ms-teams-status-log -file <file_path> -output <output_path>")
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.LError(err.Error())
	}

	defer watcher.Close()

	if err := watcher.Add(*filePath); err != nil {
		logger.LError(err.Error())
	}

	var lastLine string

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if err := processFile(*filePath, &lastLine, *outputPath); err != nil {
						logger.LError("Error:" + err.Error())
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.LError("Error:" + err.Error())
			}
		}
	}()

	logger.LInfo("Watching for changes. Press Ctrl+C to exit.")
	select {}
}

func processFile(filePath string, lastLine *string, outputPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		*lastLine = line
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	for _, str := range statusStrings {
		if strings.Contains(*lastLine, str) {
			logger.LInfo("Status changed to " + strings.Split(str, " ")[1])
			WriteLineToOutput(strings.Split(str, " ")[1], "output.txt")
			return err
		}
	}

	return nil
}

func WriteLineToOutput(line string, path string) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	_, err = writer.WriteString(time.Now().Format("2006-01-02 15:04:05") + " " + line + "\n")
	if err != nil {
		panic(err)
	}

	err = writer.Flush()
	if err != nil {
		panic(err)
	}
}
