# photo-db-fs
[![Release](https://github.com/anitschke/photo-db-fs/actions/workflows/release.yml/badge.svg)](https://github.com/anitschke/photo-db-fs/actions/workflows/release.yml) [![CI](https://github.com/anitschke/photo-db-fs/actions/workflows/ci.yml/badge.svg)](https://github.com/anitschke/photo-db-fs/actions/workflows/ci.yml) ![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/anitschke/photo-db-fs) [![Go Report Card](https://goreportcard.com/badge/github.com/anitschke/photo-db-fs)](https://goreportcard.com/report/github.com/anitschke/photo-db-fs)

`photo-db-fs` is a FUSE virtual file system for Linux that exposes a photo database as a file system. At the moment it only supports digiKam but is built to be extensible so as to support other photo management programs in the future.

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
  -db-source string
        source of the database photo-db-fs will use for querying, ie for local databases this is the path to the database
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

[anitschk@localhost ~]$ photo-db-fs --mount-point /tmp/myPhotos --db-type digikam-sqlite --db-source ~/Pictures/digikam4.db &
[1] 12910

[anitschk@localhost ~]$ ls /tmp/myPhotos/tags/DesktopBackground/photos
009318790e574d9764679ec1b8f0a987.jpg   5c23b47abdb881acee6b1f5313ecbc71.JPG   ae1628e3571980c2f8909fff5a8d860c.JPG
01586b0bc31424ab3b156ffdac8ef58f.JPG   5c4918818f3f14575f76c78ad1cf62d6.JPG   ae4beaabd272f22fbc04dd57c42d9826.jpg
0338a700e5e602c496abdcad2deaa133.jpg   5da619a984339d0e1bd4b80de91a79cd.jpg   ae5cd09cafe0f53a737d6f53d3567314.JPG

[anitschk@localhost ~]$ kill %1
```
