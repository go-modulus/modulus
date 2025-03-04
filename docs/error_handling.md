# Error Handling

Error handling is a critical part of any software system. It is essential to have a robust error handling mechanism in place to ensure that the system can recover gracefully from errors and continue to function correctly. In this section, we will discuss some best practices for error handling in our framework.

Errors can be separated to several types:
* **System errors** - errors that are caused by internal system issues such as database connection problems, network issues, etc. These errors should be logged and the system should attempt to recover from them.
* **Handled system errors** - it is a system error that has been catch and converted to framework's error.
* **User errors** - errors that are caused by incorrect input from the user or incorrect user behavior. These errors should be handled gracefully and the user should be informed about the issue.
* **Validation errors** - errors that are caused by incorrect input data. These errors should be handled by the validation layer and the user should be informed about the issue.

We have not created the separate error types for each of these categories. Instead, we use the standard Go error type and add additional information to the error message to indicate the type of error.

## System Errors

Let's take a look at the system error. Any error that is created as the default error type will be considered as a system error. For example, if we have a function that reads data from the database and returns an error if the data cannot be read, we can create an error like this:

```go
func ReadData() (string, error) {
    data, err := db.ReadData()
    if err != nil {
        return "", fmt.Errorf("failed to read data: %w", err)
    }
    return data, nil
}
```

By default, this error will be logged and the system returns an error message `Something went wrong`. This type of error needs to be hidden from the user and logged for further investigation. Usually only unchecked errors from libraries can be considered as system errors.

The JSON result returned from API looks like this:

```json
{
  "errors": [
    {
      "message": "Something went wrong (Code: cv2l9acp5ask0v7k2hmg)",
      "path": [
        "loginUser"
      ],
      "extensions": {
        "code": "internal error",
        "meta": {
          "requestId": "cv2l9acp5ask0v7k2hmg"
        }
      }
    }
  ]
}
```

It logs the error message and the request ID. The request ID is a unique identifier for the request that can be used to trace the request through the system. It is set by the RequestID middleware.

The log message looks like this:

```json
{
  "level": "error",
  "ts": "2025-03-03T08:49:11+02:00",
  "msg": "failed to read data: invalid field",
  "app": "modulus",
  "requestId": "cv2l1psp5asjt5p1qbv0",
  "error": {
    "message": "failed to read data: invalid field",
    "trace": null,
    "type": "*fmt.wrapError"
  }
}
```

For the development environment you can enable the `GQL_RETURN_CAUSE=true` environment variable to return the error message from the error cause.

In this case the returning JSON will look like this:

```json
{
  "errors": [
    {
      "message": "Something went wrong (Code: cv2m15cp5ask8340dih0)",
      "path": [
        "loginUser"
      ],
      "extensions": {
        "cause": {
          "code": "failed to read data: invalid field"
        },
        "code": "internal error",
        "meta": {
          "requestId": "cv2m15cp5ask8340dih0"
        }
      }
    }
  ]
}
```
In the result, the end user can see a message without any system details but with the request ID in the message. He can send the message to the support team. The support team can see the details in the logs by the `requestId` field.


## Handled System Errors
To be honest, sending system errors to the user is a lack of design. From our point of view any error from the library should be caught and returned as a handled system error. For example, if we have a function that reads data from the database and returns an error if the data cannot be read, we can create an error like this:

```go
import (
    mErrors "github.com/go-modulus/modulus/errors"
)
func ReadData() (string, error) {
    data, err := db.ReadData()
    if err != nil {
        return "", mErrors.WithCauseHint("Failed to read data. Try again later", err)
    }
    return data, nil
}
```

It was the first approach of handling system errors. When we don't want to think about the error codes and further error checking by outside backend layers or frontend. The JSON result returned from API looks like this:

```json
{
  "errors": [
    {
      "message": "Failed to read data. Try again later (Code: cv2mnmcp5askfr6qfj9g)",
      "path": [
        "loginUser"
      ],
      "extensions": {
        "cause": {
          "code": "invalid field"
        },
        "code": "Failed to read data. Try again later",
        "meta": {
          "requestId": "cv2mnmcp5askfr6qfj9g"
        }
      }
    }
  ]
}
``` 

The logged result looks like this:

```json
{
  "level": "error",
  "ts": "2025-03-03T10:44:09+02:00",
  "msg": "Failed to read data. Try again later",
  "app": "modulus",
  "error": {
    "cause": {
      "message": "invalid field",
      "trace": null,
      "type": "*errors.errorString"
    },
    "message": "Failed to read data. Try again later",
    "trace": null,
    "type": "errors.withCause"
  },
  "requestId": "cv2mnmcp5askfr6qfj9g"
}
```

As you can see in a case of getting such an error by the user, the user will see the message `Failed to read data. Try again later (Code: cv2mnmcp5askfr6qfj9g)` instead of the system error message.
Also, he can send it to the support team, and they will see the error message in the logs by the `message` field.

You have the second option how to handle system errors. It is to create a custom error type and return it from the function. For example, if we have a function that reads data from the database and returns an error if the data cannot be read, we can create an error like this:

```go
import (
    mErrors "github.com/go-modulus/modulus/errors"
)
var ErrCannotReadData = mErrors.NewSysError("cannot read data", "Failed to read data. Try again later")

func ReadData() (string, error) {
    data, err := db.ReadData()
    if err != nil {
        return "", mErrors.WithCause(ErrCannotReadData, err)
    }
    return data, nil
}
```

In this case you return an error from your function that can be checked outside using the regular

```go
if errors.Is(err, ErrCannotReadData) {
    // handle the error
}
```

Our error will be converted to a JSON result as:

```json
{
  "errors": [
    {
      "message": "Failed to read data. Try again later (Code: cv2nhusp5askn87dfs80)",
      "path": [
        "loginUser"
      ],
      "extensions": {
        "cause": {
          "code": "invalid field"
        },
        "code": "cannot read data",
        "meta": {
          "requestId": "cv2nhusp5askn87dfs80"
        }
      }
    }
  ]
}
```

Using the code `"code": "cannot read data"` your frontend developers can handle the error in a different way if needed.

In logs this error will be shown as:

```json
{
  "level": "error",
  "ts": "2025-03-03T11:40:11+02:00",
  "msg": "cannot read data",
  "app": "modulus",
  "error": {
    "cause": {
      "message": "invalid field",
      "trace": null,
      "type": "*errors.errorString"
    },
    "message": "cannot read data",
    "trace": null,
    "type": "errors.withCause"
  },
  "requestId": "cv2nhusp5askn87dfs80"
}
```

As you can see the support team can find an error in logs by the `message` field or by `requestId` obtained from the message.