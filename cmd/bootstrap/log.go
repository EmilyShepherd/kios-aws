package main

import "fmt"

var Levels = []string{"ERROR", "WARN", "INFO"}

// Outputs a generic log message to stdout
func log(level int, msg string) {
	fmt.Printf("[%s]\t%s\n", Levels[level], msg)
}

// Outputs an info message to stdout
func info(msg string) {
	log(2, msg)
}

// Outputs a warning message to stdout
func warn(msg string) {
	log(1, msg)
}

// Outputs an error message to stdout
func err(msg string) {
	log(0, msg)
}
