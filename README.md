## Hndlor (a-k-a Handler)

![Push Status](https://github.com/OpenRunic/hndlor/actions/workflows/master-push.yml/badge.svg)

HTTP mux library with utility methods for easy api development

#### Download
```
go get -u github.com/OpenRunic/hndlor
```

#### Example
```go
// login credentials struct
type Credentials struct {
	Username string
	Password string
}

// create primary router
r := hndlor.Router()

// attach middlewares
r.Use(hndlor.Logger(log.Writer()), hndlor.PrepareMux())

// create auth sub router
rAuth := hndlor.SubRouter("/auth")
rAuth.Handle(
  // route pattern
  "POST /login",

  // http.Handler
  hndlor.New(

    // automatically injected into callback
    func(creds Credentials) (hndlor.Json, error) {
      return hndlor.Json{
        "username": creds.Username,
        "password": creds.Password,
      }, nil
    },

    // injectable arguments [hndlor.Value]
    hndlor.Struct[Credentials](),
  ),
)

// mount the auth router
rAuth.MountTo(r)

// print routes info
hndlor.WriteStats(
  r.Mux(), // *http.ServeMux
  log.Writer(), // io.Writer

  // *hndlor.WalkConfig for nested printing since nested mux cannot be accessed
  hndlor.NewWalkConfig().Set(rAuth.Path, rAuth.Mux()),
)

// start server
if err := http.ListenAndServe(":8080", r); err != nil {
  fmt.Printf("failed to start server: %s", err.Error())
}
```

#### Middlewares
```go
// Default Middleware: Logger
hndlor.Logger(io.Writer)

// Default Middleware: PrepareMux
// Parses request and caches body if required
hndlor.PrepareMux(io.Writer)

// Simple middleware that prints message before every request
r.Use(hndlor.M(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
  println("new request!")
  next.ServeHTTP(w, r)
}))

// Middleware that responds with error on fail
r.Use(hndlor.MM(func(w http.ResponseWriter, r *http.Request, next http.Handler) error {
  return errors.New("some validation failed")
}))

// Custom timeout middleware with option(s)
func Timeout(t time.Duration) hndlor.NextHandler {
  return hndlor.M(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
    ctx, cancel := context.WithTimeout(r.Context(), t)
    defer cancel()

    next.ServeHTTP(w, r.WithContext(ctx))
  })
}

// use the timeout middleware
r.Use(Timeout(2 * time.Second))
```

#### Router

Router provides [hndlor.MuxRouter] and exposes methods identitical to [http.Handle] and [http.HandleFunc]

```go
r := hndlor.Router()

sub := hndlor.SubRouter("/nested")

sub.MountTo(r)
```

#### Handler
```go
// handler that responds with error if callback signature is invalid
hn := hndlor.New(

  // resolved values automatically injected into callback
  func(v1 string, v2 int, v3 string) (hndlor.Json, error) {
    return hndlor.Json{}, nil
  },

  // value resolvers: refer to `Values` section
  valueResolver1[string],
  valueResolver2[int],
  valueResolver3[string],
)

// handler that panics on invalid callback signature
hn := hndlor.NewP(
  ...
)

// custom callback for value resolve fail
hn.OnFail(func(hndlor.ValueResolver, error) error)
```

#### Values
```go
// value resolver from http GET
vr := hndlor.Get[string]("q")

// value resolver from http Body
vr := hndlor.Body[string]("first_name")

// value resolver from url path parameters
vr := hndlor.Path[int]("id")

// value resolver from request header
vr := hndlor.Header[string]("X-Api-Token").As("token")

// value resolver from resolved context data
vr := hndlor.Context[string]("gatewayToken").Optional()

// value resolver from custom reader
vr := hndlor.Reader(func(r *http.Request) (string, error) {
  return "user-001-uid", nil
}).As("uid")

// value resolver to struct from defined source and validate
vr := hndlor.Struct[Credentials]().As("credentials").
  Validate(func(r *http.Request, tlc Credentials) error {
    if len(tlc.Username) > 0 {
      return nil
    }
    return errors.New("unable to resolve login credentials")
  })

// collect multiple values at once as [hndlor.Json]
values, err := hndlor.Values(*http.Request,
  vr1,
  vr2,
  vr3,
  ...,
  vrN,
)

// collect multiple values as struct
var creds Credentials
err := hndlor.ValuesAs(*http.Request, &creds,
  vr1,
  vr2,
)

// resolve single value
q, err := vr.Resolve(*http.Request)
```

#### Utility
```go
// get request address [net.Addr]
addr := hndlor.RequestAddr(*http.Request)

// converts from one struct type to other
var data T
err := hndlor.StructToStruct(map[string]any{
  "user": "admin",
}, &data)

// [hndlor.Json] for reference
type Json map[string]any

// write [hndlor.Json] to io.Writer | http.ResponseWriter
hndlor.WriteData(data)

// write error to io.Writer | http.ResponseWriter
hndlor.WriteError(error)

// write error message
hndlor.WriteErrorMessage("authentication failed...")

// Custom Context Value is stored as [hndlor.Json] with key
// hndlor.ContextValueDefault

// write custom context value
req := hndlor.PatchValue(*http.Request, "gatewayToken", "0x010")

// write multiple context value
req := hndlor.PatchMap(*http.Request, hndlor.Json)

// read custom context value
val, err := hndlor.GetData[T](*http.Request, key, fallbackValue)

// read all custom context values saved as [hndlor.Json]
val, err := hndlor.GetAllData(*http.Request)

// get all cached [hndlor.Json] body data from request
bodyJson := hndlor.BodyJson(*http.Request)

// get single body value from cached body data
username, ok := hndlor.BodyRead(*http.Request, "username")

// make struct from cached body data
var creds Credentials
err := hndlor.BodyReadStruct(*http.Request, &creds)

// create error returns [hndlor.ResponseError]
err := hndlor.Error("error message")
err := hndlor.Errorf("error message: %s", name)
```

### Support

You can file an [Issue](https://github.com/OpenRunic/hndlor/issues/new).

### Contribute

To contrib to this project, you can open a PR or an issue.
