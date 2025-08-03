package util

import (
    "os"
    "path/filepath"
)

func OpenOrCreate(path string) (*os.File, error) {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return nil, err
    }

    dir := filepath.Dir(absPath)
    err = os.MkdirAll(dir, os.ModePerm)
    if err != nil {
        return nil, err
    }

    file, err := os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
    if err != nil {
        return nil, err
    }
    return file, nil
}

func MakeDirs(path string) (string, error) {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return "", err
    }

    dir := filepath.Dir(absPath)
    err = os.MkdirAll(dir, os.ModePerm)
    if err != nil {
        return "", err
    }

    return absPath, nil
}