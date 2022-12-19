# darktable-auto-export
Keep exported jpgs in sync with the latest xmp files from darktable

## Usage
```bash
go build && ./darktable-auto-export -i ~/smb-share/photo/raw -o ~/smb-share/photo/jpg
```

Alpha note: In case of error, the db lock may not be cleaned up. Remove it for flatpak with `rm ~/.var/app/org.darktable.Darktable/config/darktable/data.db.lock && rm ~/.var/app/org.darktable.Darktable/config/darktable/library.db.lock`

### Config
Config is handled by [viper](github.com/spf13/viper) and [cobra](github.com/spf13/cobra), meaning you can you command line flags or a config.yaml or config.json file. See subcommands help for more details


Here is an example config.yml showing the defaults
```
sync:
  delete-missing: false
  in: "./"
  out: "./"
  command: "flatpak run --command=darktable-cli org.darktable.Darktable"
  extension: ".ARW"
  new: false
unlock:
  lockdir: ""
```

## Roadmap
See https://github.com/figadore/darktable-auto-export/labels/roadmap
