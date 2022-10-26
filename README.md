# darktable-auto-export
Keep exported jpgs in sync with the latest xmp files from darktable

## Usage
```bash
go build && ./darktable-auto-export -i ~/smb-share/photo/raw -o ~/smb-share/photo/jpg
```

## Roadmap
- [ ] Use [when-changed](https://github.com/joh/when-changed) or similar to monitor files
