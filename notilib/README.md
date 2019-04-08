# notilib

`notilib` is a library that implements an HTTP notification client. Once configured the URL in the constructor, all notifications are sent there.

The client of this library uses the method `Notify` for adding new messages to be sent, they are first buffered in the `Message Channel` before being sent to the URL. It is an asynchronous operation so the client will not wait until the message is sent.

At a predefined interval, the library reads the content of the `Message Channel` and sends it to the URL provided using an HTTP POST method.

Notilib exposes the `Error Channel` for reporting those messages that has failed and give the oportunity to the client to handle those errors. One mechanism could be retry sending the message, for this purpose the `Retry` method is available.

![Notifier diagram](../Images/notifier.png)

## Usage

### Constructor

First we need to contruct a `notilib` entity using the `New` method:
```go
func New(url string, client *http.Client, conf *Config) (Notilib, error) 
```

Here is an example:
```go
conf := &Config{ 
    MsgChanCap: 100, 
    MaxErrChCap: 2, 
    BurstLimit: 50, 
    NumMessagesPerSecond: 10,
}

client := &http.Client{
    Timeout: time.Second * 10,
}

notilib, err = notilib.New(url, client, conf)
if err != nil {
    log.Errorf("unable to start the client: %v", err)
    return
}
```
where `client` is an optional parameter, by default `http.DefaultClient`; `conf` is also optional and those are the default values:
```go
const defaultMsgChCap = 1000
const defaultMaxErrChCap = 500
const defaultBurstLimit = 1000
const defaultNumMessagesPerSecond = 1000
```

### Start service

Once we have the `notilib` instance, we are ready to listen to the `Message Channel` for new incoming notifications to be forwarded:
```go
notilib.Listen()
```

### Send notifications

Now we can send notifications to be sent to the URL:
```go
guid, err := notilib.Notify(messages)
if err != nil {
    log.Errorf("notifier client has reported a failure: %v", err)
}
```
this returns a `GUID` assigned to all the messages and useful to track errors from the `Error Channel`, this ID has this format `0e527ed5-45a3-4c48-8b96-6fdc709da90d`.

### Handle errors
The client of notilib can handle the errors retrieving first the `Error Channel`:

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


When a new error is received, depending on the current number of retrials, the client can retry sending the same failed message using the method `notilib.Retry`.


## Components

### Notifier
The `notifier` converts the slice of strings passed as input parameter of method `Notify` to a `message` struct and insert them into the `Message Channel`. It is like a buffering, all notifications are stored in the `Message Channel` pending to be processed by the listener.

```go
type message struct {
	content     string // notification text message
	guid        string // GUID: Unique identifier
	index       int    // Index of the message from the []string passed as parameter to the notilib.Notify method
	numRetrials int    // Current number of retrials for this notification
}
```
When calling `Notify`, all the messages from the slice will have the same `guid` but different `index`.

### Listener
The `listener` is responsible for reading the messages from the `Message Channel` and pass them to the `sender` calling `sender.send(msg)`. This process uses a rate limiter to avoid exceeding the server rate limit.

### Sender
The `sender` is initialized with the URL where all notifications have to be sent.

When calling `sender.send(msg)`, it transforms the `message` struct passed as input parameter into an `*http.Request`, setting the HTTP method to POST, and pass the resulting request to the client handler.

The sender is also responsible for checking the HTTP Code of the response and if it is different than `200 OK` or `201 Created`, it will publish a new `NError` into the `Error Channel`.


### Client Handler
The client handler is responsible for sending over the network the notifications to the specified URL. 

We could call directly the function `http.Client.Do(http.Request)` but for testing purposes we have created the client handler since it allows us to mock the call `Do(req *http.Request) (*http.Response, error)`.

### Retrialer
The `retrialer` is responsible for inserting the `message` struct of a failed notification into the `Message Channel` increasing the `numRetrials` by one.


## Testing