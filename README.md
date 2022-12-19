# darktable-auto-export
Keep exported jpgs in sync with the latest xmp files from darktable

## Usage
```bash
go build -o dae ./ && ./dae -i ~/smb-share/photo/raw -o ~/smb-share/photo/jpg
```

Alpha note: In case of error, the db lock may not be cleaned up. For flatpak, the command is usually something like `./dae unlock ~/.var/app/org.darktable.Darktable/config/darktable/`, or, once the config file is updated with the `lockdir` parameter, simply `./dae unlock`

### Config
Config is handled by [viper](github.com/spf13/viper) and [cobra](github.com/spf13/cobra), meaning you can you command line flags or a config.yaml or config.json file. See subcommands help for more details


Here is an example config.yml showing the defaults
```
# sync subcommand
delete-missing: false
in: "./"
out: "./"
command: "flatpak run --command=darktable-cli org.darktable.Darktable"
extension: ".ARW"
new: false
# unlock subcommand
lockdir: ""
```

## Roadmap
See https://github.com/figadore/darktable-auto-export/labels/roadmap
