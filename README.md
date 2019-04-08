# Go Test Task `Notifier`

This go challenge has 2 components: 

- The library `notilib` responsible for sending notifications to a provided URL. More details [here](./notilib/README.md).
- The executable `notify` that uses the library `notilib` for sending notifications read from the stdin. More details [here](./notify/README.md).


Moreover, we added a test server for receiving the notifications. More details [here](./server/README.md).

![General overview](./images/overview.jpg)

