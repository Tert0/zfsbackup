# ZFSBackup
Little Backup Tool for ZFS Pools

## Installation with curl
(needs write permissions to /usr/bin)
```bash
curl -L https://github.com/Tert0/zfsbackup/releases/download/v0.8/create-backup.sh -o /usr/bin/create-backup
chmod +x /usr/bin/create-backup
curl -L https://github.com/Tert0/zfsbackup/releases/download/v0.8/linux-amd64-zfsbackup -o /usr/bin/zfsbackup
chmod +x /usr/bin/zfsbackup
```

## Uninstall
(needs write permissions to /usr/bin)
```bash
rm -rf /usr/bin/create-backup
rm -rf /usr/bin/zfsbackup
```

## Configuration
The config file is located at `~/.config/zfsbackup/config.json`.
Example Config:
```json
{
    "zfs_backup_pool": "my_backup_pool",
    "zfs_pool": "my_data_pool"
}
```
## Usage
### Create a backup
```bash
./zfsbackup create <backup or snapshot name>
```
This will create the snapshot and sync it to the backup pool.
### Delete a Backup
```bash
./zfsbackup remove <pool>@<snapshot>
```
Same as `zfs destroy <pool>@<snapshot>`.
### List Backups
```bash
./zfsbackup list
```
Example Output:
```
Name     Date                      Pool             Backup Pool
test1    Mon, 01 Jan 2022 01:01    storage@test1    backup@test1
test2    Mon, 02 Jan 2022 02:02    storage@test2    backup@test2
test3    Mon, 03 Jan 2022 03:03    storage@test3
```
### Restore a Backup
```bash
./zfsbackup restore-backup <pool>@<snapshot name>
```
This will restore the snapshot from the pool to the main pool.

**IMPORTANT: This will destroy all data and snapshots in the main pool.**
### Restore a Snapshot
```bash
./zfsbackup restore-snapshot <snapshot name>
```
This will roll back the main pool to the snapshot.
### Manual Sync
```bash
./zfsbackup sync
```
This will sync all snapshots to the backup pool.

## Create Backup with Script
```bash
./create-backup.sh <path to executable>
```
### Create without Tag automatically
```bash
echo "" | ./create-backup.sh <path to executable>
```
### Create with Tag automatically
```bash
echo "<tag-name>" | ./create-backup.sh <path to executable>
```

## License
This project is licensed under the MIT License.