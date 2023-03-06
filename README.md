# photo-db-fs
[![Release](https://github.com/anitschke/photo-db-fs/actions/workflows/release.yml/badge.svg)](https://github.com/anitschke/photo-db-fs/actions/workflows/release.yml) [![CI](https://github.com/anitschke/photo-db-fs/actions/workflows/ci.yml/badge.svg)](https://github.com/anitschke/photo-db-fs/actions/workflows/ci.yml) ![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/anitschke/photo-db-fs) [![Go Report Card](https://goreportcard.com/badge/github.com/anitschke/photo-db-fs)](https://goreportcard.com/report/github.com/anitschke/photo-db-fs)

`photo-db-fs` is a FUSE virtual file system for Linux that exposes a photo database as a file system. At the moment it only supports digiKam but is built to be extensible so as to support other photo management programs in the future. It currently supports exposing the entire tag hierarchy as a file system and also supports adding custom queries where the results of the query are exposed as a file system.

## Installation
### GitHub Release
Visit the [releases page](https://github.com/anitschke/photo-db-fs/releases) to download one of the prebuilt binaries.

### go install
Build and install from source:
```
go install github.com/anitschke/photo-db-fs@latest
```

## Usage
```
[anitschk@localhost ~]$ photo-db-fs -h
Usage of photo-db-fs:
  -config-file string
        photo-db-fs config file
  -db-source string
        source of the database photo-db-fs will use for querying, 
        ie for local databases this is the path to the database
  -db-type string
        type of photo database photo-db-fs will use for querying
  -log-level string
        debugging logging level
  -mount-point string
        location where photo-db-fs file system will be mounted
```

example:
```
[anitschk@localhost ~]$ mkdir /tmp/myPhotos

[anitschk@localhost ~]$ photo-db-fs \
                              --mount-point /tmp/myPhotos \
                              --db-type digikam-sqlite \
                              --db-source ~/Pictures/digikam4.db &
[1] 12910

[anitschk@localhost ~]$ ls /tmp/myPhotos/tags/DesktopBackground/photos
009318790e574d9764679ec1b8f0a987.jpg   5c23b47abdb881acee6b1f5313ecbc71.JPG
01586b0bc31424ab3b156ffdac8ef58f.JPG   5c4918818f3f14575f76c78ad1cf62d6.JPG
0338a700e5e602c496abdcad2deaa133.jpg   5da619a984339d0e1bd4b80de91a79cd.jpg

[anitschk@localhost ~]$ kill %1
```

Note that all flags may also be specified in the json config file specified by the `-config-file` flag.

## Custom Queries
By default `photo-db-fs` exposes the entire tag hierarchy as a file system, but it can also be configured to expose custom queries as a filesystem that can query the database for photos that match any arbitrary set operations of tags. These custom queries must be written in a json config file.

For example the following config can be used to show a directory full of photos of kayaking in New York state, a second with photos of kayaking NOT in New York state, and a third with photos of kayaking AND canoeing.
```json
{
    "queries" : [
        {
            "name": "KayakingNewYork",
            "selector": {
                "type": "and",
                "properties": {
                    "operands": { 
                        "selectors": [
                            {
                                "type": "hasTag",
                                "properties": {
                                    "tag": {
                                        "strings": [
                                            "Activity", "Kayak"
                                        ]
                                    }
                                }
                            },
                            {
                                "type": "hasTag",
                                "properties": {
                                    "tag": {
                                        "strings": [
                                            "Location", "US", "NY"
                                        ]
                                    }
                                }
                            },
                        ]
                    }
                }
            }
        },
        {
            "name": "KayakingNotNewYork",
            "selector": {
                "type": "difference",
                "properties": {
                    "starting": { 
                        "selector": {
                            "type": "hasTag",
                            "properties": {
                                "tag": {
                                    "strings": [
                                        "Activity", "Kayak"
                                    ]
                                }
                            }
                        }
                    },
                    "excluding": {
                        "selector": {
                            "type": "hasTag",
                            "properties": {
                                "tag": {
                                    "strings": [
                                        "Location", "US", "NY"
                                    ]
                                }
                            }
                        }
                    }
                }
            }
        },
        {
            "name": "KayakingOrCanoeing",
            "selector": {
                "type": "or",
                "properties": {
                    "operands": { 
                        "selectors": [
                            {
                                "type": "hasTag",
                                "properties": {
                                    "tag": {
                                        "strings": [
                                            "Activity", "Kayak"
                                        ]
                                    }
                                }
                            },
                            {
                                "type": "hasTag",
                                "properties": {
                                    "tag": {
                                        "strings": [
                                            "Activity", "Canoe"
                                        ]
                                    }
                                }
                            },
                        ]
                    }
                }
            }
        },
    ]
}
```

## Automatically Mounting
The current recommendation to automatically mount is to use a systemd service file to automatically run `photo-db-fs`. For example see we could write the following [`photo-db-fs.service`](./photo-db-fs.service) file. 
```ini
[Unit]
Description=photo-db-fs FUSE file system server

[Service]
Type=simple
ExecStart=/home/anitschk/go/bin/photo-db-fs --mount-point /srv/media/photo-db-fs --db-type digikam-sqlite --db-source /srv/media/digiKamDB/digikam4.db

[Install]
WantedBy=default.target
```

Then install as a user service using 
```bash
mkdir -p ~/.config/systemd/user/
cp photo-db-fs.service ~/.config/systemd/user/
systemctl --user enable photo-db-fs.service
systemctl --user start photo-db-fs.service
```

<details>
  <summary>Correct way automatically mount</summary>

A systemd .service file isn't really the right way to do this. The correct way would be to hook it up via /etc/fstab or a systemd .mount file instead. There is some good discussion on how rclone (also written in Go) does this and it sounds like they need to do some interpreting/translating of how  `mount` passes it the options because it is a little non-standard. I did some looking in their source but don't see anything obvious as to how they are doing it. see [rclone doc](https://rclone.org/commands/rclone_mount/#rclone-as-unix-mount-helper)
</details>
