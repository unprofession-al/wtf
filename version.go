package main

import (
	"fmt"
	"runtime/debug"
)

func VersionInfo() string {
	return fmt.Sprintf("Commit: %s\n", GetCommit())
}

func GetCommit() string {
	out := "unknown"
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return out
	}
	commit := "unknown"
	state := "unknown"
	for _, setting := range info.Settings {
		fmt.Println(setting.Key, setting.Value)
		switch setting.Key {
		case "vcs.revision":
			commit = setting.Value
		case "vcs.modified":
			if setting.Value == "true" {
				state = "dirty"
			} else {
				state = "clean"
			}
		}
	}
	out = fmt.Sprintf("%s (%s)", commit, state)
	return out
}
