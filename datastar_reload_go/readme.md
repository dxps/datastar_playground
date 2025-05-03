## Reload on (code) changes

This is a simple and effective example (taken from [here](https://github.com/starfederation/datastar/blob/main/sdk/go/examples/hotreload/main.go)) of doing page reload on detected code changes.

Hot reload requires a file system watcher and a refresh script.

As file system watcher (of changes), two popular tools for this purpose are:

-   [reflex](https://github.com/cespare/reflex)
-   [air](https://github.com/air-verse/air)

The latter is used in this case (see `run.sh` script) simply because it's still maintained.

The refresh script is a Datastar handler that emits<br/>
a page refresh event only once for each server start.

When the the file watcher forces the server to restart,<br/>
Datastar client will lose the network connection to the<br/>
server and attempt to reconnect. Once the connection is<br/>
established, the client will receive the refresh event.
