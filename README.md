# ZFSBackup
Little Backup Tool for ZFS Pools

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

## License
This project is licensed under the MIT License.