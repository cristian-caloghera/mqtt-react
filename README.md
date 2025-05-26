# mqtt-react

Small Go application that reacts on MQTT message by executing an arbitrary command.

## How to run

```
mqtt-react [path/to/config/file]
```

## Configuration

The application takes a single optional parameter, that is the path to the configuration file. If not given it defaults to `./mqtt-react.yaml`.
The configuration supports multiple brokers, and multiple topics per broker.

The string `$payload` is replaced by the payload received.

For an example look at the provided sample configuration file.

### Known limitations

The current implementation:
 - converts the payload to a string. That means binary payloads are not supported.
 - executes the command inside a shell, whatever `sh` defaults to in your environment
