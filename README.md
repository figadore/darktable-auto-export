# darktable-auto-export
Keep exported jpgs in sync with the latest xmp files from darktable

## Usage
```bash
go build && ./darktable-auto-export -i ~/smb-share/photo/raw -o ~/smb-share/photo/jpg
```

In case of error, the db lock may not be cleaned up. Remove it for flatpak with `rm ~/.var/app/org.darktable.Darktable/config/darktable/data.db.lock`
