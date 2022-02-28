package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

type Snapshot struct {
	Name string
	Date time.Time
}

func GetAllSnapshots(pool string, verbose bool) []Snapshot {
	output, err := ExecuteCommand(verbose, "zfs", "list", "-t", "snapshot", pool, "-H", "-o", "name,creation", "-p")
	if err != nil {
		panic(err)
	}
	lines := strings.Split(output, "\n")
	var snapshots []Snapshot
	for line := range lines {
		if lines[line] == "" {
			continue
		}
		timestamp, err := strconv.ParseInt(strings.Split(lines[line], "\t")[1], 10, 64)
		if err != nil {
			return nil
		}
		date := time.Unix(timestamp, 0)
		snapshot := Snapshot{
			Name: strings.TrimPrefix(strings.Split(lines[line], "\t")[0], pool+"@"),
			Date: date,
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots
}

func snapshotsContain(snapshots []Snapshot, snapshot Snapshot) bool {
	for _, s := range snapshots {
		if s.Name == snapshot.Name && s.Date == snapshot.Date {
			return true
		}
	}
	return false
}

func snapshotsContainName(snapshots []Snapshot, name string) bool {
	for _, s := range snapshots {
		if s.Name == name {
			return true
		}
	}
	return false
}

func getLastSyncedSnapshot(snapshots []Snapshot, backupSnapshots []Snapshot) Snapshot {
	var lastSnapshot Snapshot

	for _, snapshot := range backupSnapshots {
		if snapshotsContain(snapshots, snapshot) {
			if lastSnapshot.Date.IsZero() || snapshot.Date.After(lastSnapshot.Date) {
				lastSnapshot = snapshot
			}
		}
	}
	return lastSnapshot
}

func getNotSyncedSnapshot(lastSyncedSnapshot Snapshot, snapshots []Snapshot, backupSnapshots []Snapshot) Snapshot {
	var notSyncedSnapshot Snapshot
	date := lastSyncedSnapshot.Date
	for _, snapshot := range snapshots {
		if snapshot.Date.After(date) && !snapshotsContain(backupSnapshots, snapshot) {
			if notSyncedSnapshot.Date.IsZero() || snapshot.Date.Before(notSyncedSnapshot.Date) {
				notSyncedSnapshot = snapshot
			}
		}
	}

	return notSyncedSnapshot
}

func syncSnapshot(pool string, newPool string, lastSnapshot Snapshot, snapshot Snapshot, incremental bool) error {
	var err error
	var sendCmd *exec.Cmd
	if incremental {
		sendCmd = exec.Command("zfs", "send", "-I", pool+"@"+lastSnapshot.Name, pool+"@"+snapshot.Name)
	} else {
		sendCmd = exec.Command("zfs", "send", pool+"@"+snapshot.Name)
	}
	recvCmd := exec.Command("zfs", "receive", "-F", newPool)
	sendCmdStdout, err := sendCmd.StdoutPipe()
	if err != nil {
		return err
	}
	recvStdin, err := recvCmd.StdinPipe()
	if err != nil {
		return err
	}
	err = sendCmd.Start()
	if err != nil {
		return err
	}
	err = recvCmd.Start()
	if err != nil {
		return err
	}
	dataWrote := 0
	buffer := make([]byte, 1024*1024*16)
	for {
		read, err := sendCmdStdout.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("read error")
			return err
		}
		if read == 0 {
			break
		}
		write, err := recvStdin.Write(buffer[:read])
		if err != nil {
			fmt.Println("write error")
			return err
		}
		if read != write {
			return fmt.Errorf("read %d bytes, wrote %d bytes", read, write)
		}
		dataWrote += read

		func() {
			fmt.Print("\r")
			fmt.Printf("Wrote %.2f MB", float64(dataWrote)/1024.0/1024.0)
		}()
	}
	fmt.Println("")
	err = recvCmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func syncPools(config Config) {
	firstIteration := true
	for {
		snapshots := GetAllSnapshots(config.ZFSPool, false)
		backupSnapshots := GetAllSnapshots(config.ZFSBackupPool, false)
		lastSyncedSnapshot := getLastSyncedSnapshot(snapshots, backupSnapshots)
		notSyncedSnapshots := getNotSyncedSnapshot(lastSyncedSnapshot, snapshots, backupSnapshots)
		if firstIteration {
			fmt.Println("Last Synced Snapshot:", lastSyncedSnapshot.Name)
			fmt.Println("Last not Synced Snapshot:", notSyncedSnapshots.Name)
		}
		firstIteration = false
		if notSyncedSnapshots.Name == "" {
			break
		}

		if len(backupSnapshots) > 0 && lastSyncedSnapshot.Name != "" { // syncSnapshot incremental
			err := syncSnapshot(config.ZFSPool, config.ZFSBackupPool, lastSyncedSnapshot, notSyncedSnapshots, true)
			if err != nil {
				panic(err)
			}
		} else { // first syncSnapshot
			err := syncSnapshot(config.ZFSPool, config.ZFSBackupPool, lastSyncedSnapshot, notSyncedSnapshots, false)
			if err != nil {
				panic(err)
			}
		}
	}
	GetAllSnapshots(config.ZFSPool, true)
	GetAllSnapshots(config.ZFSBackupPool, true)
	fmt.Println("Sync Done")
}

func createSnapshot(config Config, name string) error {
	_, err := ExecuteCommand(false, "zfs", "snapshot", config.ZFSPool+"@"+name)
	if err == nil {
		fmt.Println("Snapshot Created:", name)
	}
	return err
}

func SyncPools(config Config) {
	fmt.Println("IMPORTANT: This is a one-way sync: ZFSPool -> ZFSBackupPool")
	syncPools(config)
}

func CreateBackup(name string, config Config) error {
	snapshots := GetAllSnapshots(config.ZFSPool, false)
	if snapshotsContainName(snapshots, name) {
		return fmt.Errorf("snapshot %s@%s already exists", config.ZFSPool, name)
	}
	snapshots = GetAllSnapshots(config.ZFSBackupPool, false)
	if snapshotsContainName(snapshots, name) {
		return fmt.Errorf("snapshot %s@%s already exists", config.ZFSBackupPool, name)
	}
	syncPools(config)
	err := createSnapshot(config, name)
	if err != nil {
		return err
	}
	syncPools(config)
	return nil
}

func mergeSnapshotLists(snapshots1 []Snapshot, snapshots2 []Snapshot) []Snapshot {
	snapshots := make([]Snapshot, 0)
	for _, snapshot := range snapshots1 {
		if !snapshotsContain(snapshots, snapshot) {
			snapshots = append(snapshots, snapshot)
		}
	}
	for _, snapshot := range snapshots2 {
		if !snapshotsContain(snapshots, snapshot) {
			snapshots = append(snapshots, snapshot)
		}
	}
	return snapshots
}

func ListSnapshots(config Config) {
	type SnapshotEntry struct {
		Snapshot
		Pool       bool
		BackupPool bool
	}
	entries := make([]SnapshotEntry, 0)
	snapshots := GetAllSnapshots(config.ZFSPool, false)
	backups := GetAllSnapshots(config.ZFSBackupPool, false)
	for _, snapshot := range mergeSnapshotLists(snapshots, backups) {
		onPool := snapshotsContain(snapshots, snapshot)
		onBackupPool := snapshotsContain(backups, snapshot)
		entries = append(entries, SnapshotEntry{snapshot, onPool, onBackupPool})
	}
	writer := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)
	_, err := fmt.Fprintln(writer, "Name\tDate\tPool\tBackup Pool")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		entryPool := ""
		entryBackupPool := ""
		if entry.Pool {
			entryPool = config.ZFSPool + "@" + entry.Name
		}
		if entry.BackupPool {
			entryBackupPool = config.ZFSBackupPool + "@" + entry.Name
		}
		_, err := fmt.Fprintln(writer, entry.Name+"\t"+entry.Date.Format("Mon, 02 Jan 2006 15:04")+"\t"+entryPool+"\t"+entryBackupPool)
		if err != nil {
			panic(err)
		}
	}
	err = writer.Flush()
	if err != nil {
		panic(err)
	}
}

func getSnapshotByName(snapshots []Snapshot, name string) Snapshot {
	for _, snapshot := range snapshots {
		if snapshot.Name == name {
			return snapshot
		}
	}
	return Snapshot{}
}

func deleteSnapshot(pool string, name string) error {
	_, err := ExecuteCommand(false, "zfs", "destroy", pool+"@"+name)
	if err == nil {
		fmt.Println("Snapshot Deleted:", name)
	}
	return err
}

func RestoreBackup(config Config, name string) error {
	var err error
	pool := strings.Split(name, "@")[0]
	if pool == config.ZFSPool {
		return fmt.Errorf("to restore a snapshot of the same pool please use the restore-snapshot command")
	}
	snapshotName := strings.Split(name, "@")[1]
	snapshots := GetAllSnapshots(pool, false)
	if !snapshotsContainName(snapshots, snapshotName) {
		return fmt.Errorf("snapshot %s@%s does not exist", pool, snapshotName)
	}
	snapshot := getSnapshotByName(snapshots, snapshotName)
	if snapshot.Name == "" {
		return fmt.Errorf("snapshot %s@%s does not exist", pool, snapshotName)
	}
	if len(GetAllSnapshots(config.ZFSPool, false)) > 0 {
		fmt.Print("THIS WILL DESTROY ALL DATA ON THE POOL \"" + config.ZFSPool + "\"! Are you sure? (y/N) ")
		reader := bufio.NewReader(os.Stdin)
		answer, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Canceled restore process!")
			return nil
		}
		err = deleteSnapshot(config.ZFSPool, "%") // zfs destroy pool@% will delete all snapshots of the pool
		if err != nil {
			return err
		}
	}
	fmt.Println("Deleted all snapshots of the pool")
	fmt.Println("Restoring snapshot", snapshot.Name)
	err = syncSnapshot(pool, config.ZFSPool, snapshot, snapshot, false)
	if err != nil {
		return err
	}
	return nil
}

func RestoreSnapshot(config Config, name string) error {
	snapshots := GetAllSnapshots(config.ZFSPool, false)
	if !snapshotsContainName(snapshots, name) {
		return fmt.Errorf("snapshot %s@%s does not exist", config.ZFSPool, name)
	}
	fmt.Print("THIS WILL ROLLBACK ALL DATA ON THE POOL \"" + config.ZFSPool + "\" TO THE SNAPSHOT \"" + config.ZFSPool + "@" + name + "\"! Are you sure? (y/N) ")
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer != "y" && answer != "yes" {
		fmt.Println("Canceled restore process!")
		return nil
	}
	_, err = ExecuteCommand(false, "zfs", "rollback", "-r", config.ZFSPool+"@"+name)
	if err != nil {
		return err
	}
	fmt.Println("Restored snapshot", config.ZFSPool+"@"+name, "to", config.ZFSPool)
	return nil
}
