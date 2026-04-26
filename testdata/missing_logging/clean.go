package main

import (
    "os"
    "log"
)

func example() {
    data, err := os.ReadFile("file.txt")
    if err != nil {
        return err
    }
    log.Printf("Read file: file.txt")
    _ = data
}