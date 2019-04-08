# Go Test Task `Notifier`

## Notify: The executable

### Installation

``` bash
$ cd notify
$ go install
```

### Help
``` bash
$ notify
usage: notify --url=URL [<flags>]

Flags:
        --help                  Shows context-sensitive help
        -i, --interval=5s       Notification interval
```


``` bash
$ notify --help
Usage of notify:
  -i duration
        Notification interval (shorthand) (default 5s)
  -interval duration
        Notification interval (default 5s)
  -url string
        URL where to send notifications
```