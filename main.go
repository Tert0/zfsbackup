package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Usage:", os.Args[0], "[create|remove|list|restore-backup|restore-snapshot|sync]")
		os.Exit(1)
	}

	config := ReadConfig()

	switch os.Args[1] {
	case "create":

		if len(os.Args) <= 2 {
			fmt.Println("Please provide a snapshot/backup name")
			os.Exit(1)
		}

		err := CreateBackup(os.Args[2], config)
		if err != nil {
			panic(err)
		}
	case "remove":
		panic("Not implemented")
	case "list":
		ListSnapshots(config)
	case "restore-backup":
		if len(os.Args) <= 2 {
			fmt.Println("Please provide a pool and snapshot. Example: pool@snapshot")
			os.Exit(1)
		}
		err := RestoreBackup(config, os.Args[2])
		if err != nil {
			panic(err)
		}
	case "restore-snapshot":
		if len(os.Args) <= 2 {
			fmt.Println("Please provide snapshot name")
			os.Exit(1)
		}
		err := RestoreSnapshot(config, os.Args[2])
		if err != nil {
			panic(err)
		}
	case "sync":
		SyncPools(config)

	default:
		fmt.Println("Unknown command")
		os.Exit(1)
	}
}
