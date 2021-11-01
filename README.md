[![Go Report Card](https://goreportcard.com/badge/github.com/smford/motioneye-snapshotter)](https://goreportcard.com/report/github.com/smford/motioneye-snapshotter) [![License: GPL v3](https://img.shields.io/badge/License-Apache%20v2-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)

MotionEye Snapshotter
=====================

A simple tool that allows image snapshots to be made from cameras managed by [MotionEye](https://github.com/ccrisan/motioneye).

Features:
- take snapshots from cameras managed by MotionEye
- stores images
- simple web interface
- simple snapshot viewer
- provides a simple API that a device can call, initiating a snapshot, and returning a url to the snapshot


Usage Scenario
--------------

MotionEye Snapshotter is useful for when you have a lightweight IOT device that is used to trigger snapshots.  For example a door monitoring system detects when a door is opened, which then tells MotionEye Snapshotter to take and store a timestamped snapshot of a camera facing the door.


Installation
------------
### Via go
- `# go get -v github.com/smford/motioneye-snapshotter`


### From Git
Clone git repo and build yourself
- `# git clone git@github.com:smford/motioneye-snapshotter.git`
- `# cd motioneye-snapshotter`
- `# go build`


### Docker
1. Create the docker volume to store configuration
    ```
    # docker volume create motioneye-snapshotter
    ```
1. Copy `config.yaml` and `index.html` to `/var/lib/docker/volumes/motioneye-snapshotter/_data/`
1. Start up motioneye-snapshotter
    ```
    # docker run --name motioneye-snapshotter -d --restart always -p 54036:5757/tcp -v motioneye-snapshotter:/config smford/motioneye-snapshotter:latest
    ```


Configuration
-------------

### Configuring MotionEye Snapshotter
Create a configuration file called `config.yaml` an example is available below:
```
indexfile: /path/to/index.html
listenip: 0.0.0.0
listenport: 5757
meserver: http://127.0.0.1:8765
meuser: admin
snapshoturl: https://snapshots.mydomain.com
outputdir: /path/to/outputdir
cameras:
  1: main-door
  2: lobby
camerasigs:
  1: xxxxXXxxxxxXxxxx
  2: yyYYyYyyYYyYyyYy
```


#### Configuration File Options
| Setting | Details |
|:--|:--|
| indexfile | the name and path to the file that is shown when people visit the main page of motioneye-snapshotter |
| listenip | The IP for simple-canary to listen to, 0.0.0.0 = all IPs |
| listenport | The port that simple-canary should listen on |
| meserver | The URL to the MotionEye Server |
| meuser | The MotionEye user to login and get the snapshot as |
| snapshoturl | Pretty URL for people to access the web interface of motioneye-snapshotter |
| outputdir | Directory to store all snapshots |
| cameras | A list of cameras, with `1` being the camera number within MotionEye, and `main-door` being the name you have chosen to easily identify the camera |
| camerasigs | A list of cameras and their snapshot signitures (details below) |


Starting motioneye-snapshotter
----------------------
### From command line
After creating the config.yaml and the index.html file simply run:

`# motioneye-snapshotter --config /path/to/config.yaml`


### By Docker
See the instructions here https://github.com/smford/motioneye-snapshotter#docker


Configure Clients
-----------------

- IoT Device detects that a door has opened and wants a snapshot of the main-door camera to be taken:
  Configure it to do an http request to: `http://192.168.10.1:54036/snap?camera=main-door`
- Using a cronjob to take snapshots every 5 minutes:
  `*/5 * * * * wget --spider "http://192.168.10.1:54036/snap?camera=main-door" >/dev/null 2>&1`


Command Line Options
--------------------
```
  --config [config file]             Configuration file: /path/to/file.yaml, default = ./config.yaml
  --displayconfig                    Display configuration
  --help                             Display help
```

API Endpoints
-------------
Assuming motioneye-snapshotter is configured to use 192.168.10.1:54036 and there is a camera called `main-door`

| Task | URL | Result |
|:--|:--|:--|
| Display index.html | `http://192.168.10.1:54036/` | |
| Take a snapshot of `main-door` camera | `http://192.168.10.1:54036/snap?camera=main-door` | Returns path to snapshot taken |
| Browse cameras and the snapshots taken | `http://192.168.10.1:54036/files` | |
| See what cameras are configured | `http://192.168.10.1:54036/cameras` | |
| Download a specific snapshot | `http://192.168.10.1:54036/files/main-door?file=20211012_124955.jpg` | Downloads image |
