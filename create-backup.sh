#!/bin/bash
if [ -z "$1" ]
then
  echo "Usage: $0 <zfsbackup-executable>"
  exit 1
fi
echo "Enter a tag for the backup:"
read backup_tag
if [ -z "$backup_tag" ]; then
  backup_name="$(date +'%Y-%m-%d-%H:%M')"
else
  backup_name="$(date +'%Y-%m-%d-%H:%M')-$backup_tag"
fi
$1 create "$backup_name"
if [ $? -ne 0 ]; then
  echo "Backup failed"
  exit 1
fi
echo "Backup created: $backup_name"