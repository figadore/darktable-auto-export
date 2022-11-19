# darktable-auto-export
Keep exported jpgs in sync with the latest xmp files from darktable

## Usage
```bash
go build && ./darktable-auto-export -i ~/smb-share/photo/raw -o ~/smb-share/photo/jpg
```

Alpha note: In case of error, the db lock may not be cleaned up. Remove it for flatpak with `rm ~/.var/app/org.darktable.Darktable/config/darktable/data.db.lock && rm ~/.var/app/org.darktable.Darktable/config/darktable/library.db.lock`

### Config
To delete jpgs where the raw file is no longer found, ensure that a file called config.yml exists in the directory where this binary is run. It should have the following contents

```yaml
delete-missing: true
```

This is useful for darktable workflows where editing and culling can be done at any time, not just up front.

:warning: This will delete all jpgs in the output directory where a corresponding raw file with the specified extension cannot be found! Only use this for directories that are exclusively for this workflow, and where the source files stay where they are/were.



## Roadmap
See https://github.com/figadore/darktable-auto-export/labels/roadmap
