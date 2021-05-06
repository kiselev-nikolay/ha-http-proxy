# Example of proxy service

HTTP server for proxying HTTP-requests to 3rd-party services. The server is waiting HTTP-request from client (curl, for example). In request's body there should be message in JSON format. Like:

```json
{
  "method": "GET",
  "url": "http://google.com",
  "headers": {
    "Authentication": "Basic bG9naW46cGFzc3dvcmQ="
  }
}
```

Server forms valid HTTP-request to 3rd-party service with data from client's message and responses to client with JSON object:

```js
{
  "id": "abcdef",  // generated unique id
  "status": 200,   // HTTP status of thrird-party service response
  "headers": {     // headers array from thrird-party service response
    "X-Example": "Basic bG9naW46cGFzc3dvcmQ="
  },
  "length": 255    // content length of thrird-party service response
}
```

## Contents

+ Implementation using only the standard library ➡ [kiselev-nikolay/ha-http-proxy/pure](http://github.com/kiselev-nikolay/ha-http-proxy/pure)
+ Implementation using only the gin gonic framework ➡ [kiselev-nikolay/ha-http-proxy/gin](http://github.com/kiselev-nikolay/ha-http-proxy/gin)