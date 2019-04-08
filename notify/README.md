#  Notify


![Notify executable](../images/notify.jpg)


## Installation
Execute the following commands from a terminal to install the `notify` executable:

``` bash
$ cd notify
$ go install
```


### Help 

If we type just the executable name `notify` we obtain a usage help response:

``` bash
$ notify
usage: notify --url=URL [<flags>]

Flags:
        --help                  Shows context-sensitive help
        -i, --interval=5s       Notification interval
        -c, --chcap=500         Channel capacity for reading from stdin
        -r, --retrials=2        Maximal number of retrials when receives an error sending a notification
        -m, --messages=100      Maximal number of messages to be processed per interval
```

We can also use the `--help` flag to obtain more help:

``` bash
$ notify --help
Usage of notify:
  -c int
        Channel capacity for reading from stdin (shorthand) (default 500)
  -chcap int
        Channel capacity for reading from stdin (default 500)
  -i duration
        Notification interval (shorthand) (default 5s)
  -interval duration
        Notification interval (default 5s)
  -m int
        Maximal number of messages to be processed per interval (shorthand) (default 100)
  -messages int
        Maximal number of messages to be processed per interval (default 100)
  -r int
        Maximal number of retrials when receives an error sending a notification (shorthand) (default 2)
  -retrials int
        Maximal number of retrials when receives an error sending a notification (default 2)
  -u string
        URL where to send notifications (shorthand)
  -url string
        URL where to send notifications
```

## Usage
To execute the `notify` program, at least we have to provide the mandatory flag `url` with the URL where to send the notifications using HTTP POST method:
```bash
$ notify --url=http://localhost:9090/api/notifications 
INFO[2019-04-08T21:30:16+02:00] HTTP Notification client started. Listening for new messages from stdin... 
```

We could also use the flags for changing the default configuration values:

```bash
$ notify --url=http://localhost:9090/api/notifications --loglevel=debug --interval=10s --chcap=100 --retrials=5 --messages=50
INFO[2019-04-08T21:29:19+02:00] HTTP Notification client started. Listening for new messages from stdin... 
DEBU[2019-04-08T21:29:19+02:00] Notify configuration: 
{
  url: "http://localhost:9090/api/notifications",
  interval: "10s",
  channelCapacity: 100,
  maxNumRetrials: 5,
  maxNumMessagesToProcess: 50,
}
 
DEBU[2019-04-08T21:29:19+02:00] Notilib configuration: 
{
  BurstLimit: 1000,
  NumMessagesPerSecond: 1000,
  MsgChanCap: 1000,
  ErrChanCap: 500,
  LogLevel: debug,
}
```

Once the program is running we can write sentences using the keyboard. Each time we press Return the message is inserted into the `Stdin Channel` and they will be sent to the URL specified by the notilib:
```bash
zeus:notify danielgil$ notify --url=http://localhost:9090/api/notifications -loglevel=debug -i=10s
INFO[2019-04-08T21:55:05+02:00] HTTP Notification client started. Listening for new messages from stdin... 
hello world
my name is
Daniel
how are you?
DEBU[2019-04-08T21:55:15+02:00] new tick. Num messages in channel: 4         
DEBU[2019-04-08T21:55:15+02:00] queuing new messages: [hello world my name is Daniel how are you?] 
INFO[2019-04-08T21:55:15+02:00] messages received: GUID=b97c73c3-02e0-4945-9247-a81f4f208c71 
DEBU[2019-04-08T21:55:15+02:00] message[0] added, content=hello world        
DEBU[2019-04-08T21:55:15+02:00] message[1] added, content=my name is         
DEBU[2019-04-08T21:55:15+02:00] message[2] added, content=Daniel             
DEBU[2019-04-08T21:55:15+02:00] message[3] added, content=how are you?       
DEBU[2019-04-08T21:55:15+02:00] 4 messages inserted into the msgCh           
DEBU[2019-04-08T21:55:15+02:00] Message sent correctly: HttpCode=200 OK, GUID=[b97c73c3-02e0-4945-9247-a81f4f208c71], index=3 
DEBU[2019-04-08T21:55:15+02:00] Message sent correctly: HttpCode=200 OK, GUID=[b97c73c3-02e0-4945-9247-a81f4f208c71], index=1 
DEBU[2019-04-08T21:55:15+02:00] Message sent correctly: HttpCode=200 OK, GUID=[b97c73c3-02e0-4945-9247-a81f4f208c71], index=2 
DEBU[2019-04-08T21:55:15+02:00] Message sent correctly: HttpCode=200 OK, GUID=[b97c73c3-02e0-4945-9247-a81f4f208c71], index=0 
```

## Processing messages
Each `interval` (value that can be configured using a flag, by default is 5 seconds) the program reads the messages from the `Stdin Channel`, create an slice of strings and pass them to the notilib by calling `notilib.Notify(messages)`. 

To avoid that this task takes too much time, we limit the maximal number of messages to be read from the `Stdin Channel`, by default this value is `defaultMaxNumMessagesToProcess=100` but it can be changed using the flag `messages`.

## Error Handler
The program starts an error handler which is responsible for detecting errors (notifications not delivered) and tries to send them again.

First of all it retrieves the `Error Channel` provided by the notilib:

```go
errCh := notilib.GetErrorChannel()
```

The value type used in `Error Channel` is `NError` with those fields:
```go 
type NError struct {
	Error       string // Error message
	Message     string // Original failed notification
	NumRetrials int    // Number of retrials
	GUID        string // Unique identifier
	Index       int    // Index of the failed message from the []string passed as parameter to the notilib.Notify method
}
```

Once the client have access to the `Error Channel`,  it can create a goroutine that stays blocked just listening for new errors:
```go
go func(errCh <-chan nl.NError) {
    for {
        select {
        case e := <-errCh:
            if e.NumRetrials <= maxNumRetrials {
                // retry to send this failed notification
                notilib.Retry(e.Message, e.GUID, e.Index, e.NumRetrials)
            }
        }
    }
}(errCh)
```

When a new error is received, depending on the current number of retrials, the client can decide to send the same failed message using the method `notilib.Retry` (if it hasn't exceed the maximal number of retrials allowed).

## Test redirecting stdin

Create a file:
```bash
$ vi text.txt
```

Add some content to the file:
```
Hello World!
My name is Daniel.
Go Gophers
Visual Studio Code
```

Execute the program redirecting the standard input from the file:
```bash
$ notify --url=http://localhost:9090/api/notifications -loglevel=debug  < text.txt
INFO[2019-04-08T22:13:11+02:00] HTTP Notification client started. Listening for new messages from stdin... 
DEBU[2019-04-08T22:13:11+02:00] Notify configuration: 
{
  url: "http://localhost:9090/api/notifications",
  interval: "5s",
  channelCapacity: 500,
  maxNumRetrials: 2,
  maxNumMessagesToProcess: 100,
}
 
DEBU[2019-04-08T22:13:11+02:00] Notilib configuration: 
{
  BurstLimit: 1000,
  NumMessagesPerSecond: 1000,
  MsgChanCap: 1000,
  ErrChanCap: 500,
  LogLevel: debug,
}
 
DEBU[2019-04-08T22:13:16+02:00] new tick. Num messages in channel: 4         
DEBU[2019-04-08T22:13:16+02:00] queuing new messages: [Hello World! My name is Daniel. Go Gophers Visual Studio Code] 
INFO[2019-04-08T22:13:16+02:00] messages received: GUID=7158340c-0526-4cf9-9f7b-231d15a31990 
DEBU[2019-04-08T22:13:16+02:00] message[0] added, content=Hello World!       
DEBU[2019-04-08T22:13:16+02:00] message[1] added, content=My name is Daniel. 
DEBU[2019-04-08T22:13:16+02:00] message[2] added, content=Go Gophers         
DEBU[2019-04-08T22:13:16+02:00] message[3] added, content=Visual Studio Code 
DEBU[2019-04-08T22:13:16+02:00] 4 messages inserted into the msgCh           
DEBU[2019-04-08T22:13:16+02:00] Message sent correctly: HttpCode=200 OK, GUID=[7158340c-0526-4cf9-9f7b-231d15a31990], index=3 
DEBU[2019-04-08T22:13:16+02:00] Message sent correctly: HttpCode=200 OK, GUID=[7158340c-0526-4cf9-9f7b-231d15a31990], index=0 
DEBU[2019-04-08T22:13:16+02:00] Message sent correctly: HttpCode=200 OK, GUID=[7158340c-0526-4cf9-9f7b-231d15a31990], index=1 
DEBU[2019-04-08T22:13:16+02:00] Message sent correctly: HttpCode=200 OK, GUID=[7158340c-0526-4cf9-9f7b-231d15a31990], index=2 
```