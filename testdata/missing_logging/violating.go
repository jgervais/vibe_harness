package main

import "os"

func example() {
    data, err := os.ReadFile("file.txt")
    if err != nil {
        return err
    }
    _ = data
}