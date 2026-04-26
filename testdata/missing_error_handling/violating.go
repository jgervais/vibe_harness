package main

import "os"

func example() {
    data, _ := os.ReadFile("file.txt")
    _ = data
}