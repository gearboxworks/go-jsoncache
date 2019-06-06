# go-jsoncache

The package provides an extremely **simple JSON value caching package**
for developers to build CLI apps with Go. JSON values are cached with
keys developers using this package will provide and in a named
subdirectory within the user's home directory of the developer's
choosing. 

## Primary Use-Case
The primary use-case for this package is to cache JSON API responses locally to improve performance. 

Any app that depends on accessing JSON APIs and/or JSON files hosted
remotely on a webserver, such as JSON files found in GitHub
repositories, a.k.a. _"a poor man's JSON API"_ can benefit from
`go-jsoncache` if the app is not currently caching its HTTP requests. 

## Package Dependencies
`go-jsoncache` has two (2) dependencies, one of which we developed too:

1. `github.com/gearboxworks/go-status` — Our "error" package that allows
   for returning success messages as well as error messages. 
2. `github.com/mitchellh/go-homedir`
   — [Hashicorp](https://www.hashicorp.com/)'s defacto-canonical package
   for determining home directories 

## Concurrency/Go Routines
This package does not attempt to support concurrent use, although we are
**not opposed to pull requests** that would address those limitations. 

## Using `go-jsoncache`

First create an instance of a `jsoncache` object:

### `New()`
```
cache := jsoncache.New(".my-app")
```

Next identify a cache key — which must be globally unique with respect
to any and all cache keys for a given computer — and call the `Set()`
method with your key, a byte slice a.k.a. `[]byte` and finally a string
representing a duration that can be parsed with `time.ParseDuration()`: 

```
b,_ := json.Marshal(mydata)
key := "key-for-my-data"
sts := cache.Set(key, b, "600s") // 
```

### `Get()`
Later, to retrieve a cache value create a new cache object with the same
directory you used previous and call `Get()` with the same cache key: 
 
```
cache := jsoncache.New(".my-app")
b,ok,sts := cache.Get("key-for-my-data")
if is.Error(sts) {
   panic(sts.Message())
}
if is.Success(sts) {
   fmt.Printf("Hurray! %s", sts.Message())
}
```

#### `Get()`'s return values
If a cache value can be retrieved it will be in the first return value,
the second will be `true` and the third will be a status value where
`sts.IsSuccess()` will be `true`. 

If an error occurs the first return value will be a zero `[]byte`, the
second will be `false` and `sts.IsError()` will be `true` with
`sts.Message()` providing the error message. To learn more about
statuses vs. errors see
[the `go-status` package](https://github.com/gearboxworks/go-status). 

If the case of a cache expiration then the third return value will be
indicate a success with `sts.IsSuccess()` but the 2nd parameter will be
`false` indicating the cache expired, and of course the 1st return value
will be a zero `[]byte`. 


### `Clear()`
You can clear a key's cached value before it expires with `Clear()`:

```
cache := jsoncache.New(".my-app")
cache.Clear("key-for-my-data")
```

### Type Aliases

The `go-jsoncache` package defines the following
[types as aliases](https://yourbasic.org/golang/type-alias/) to
`string`:

- `Filepath` — Intended to represent a full local file path
- `Path` — Intended to represent a relative file path
- `Dir` — Intended to represent a full local directory path
- `Key` — Intended to be used with cache keys

### Other Cache Methods

There are several other methods provided, but we have not yet documented them:

- `VerifyCacheFile(Key) (Filepath, status.Status)`
- `GetCacheFilepath(Key) Filepath` 

## Cache File Location(s)
By default the cache location is a subdirectory name of your own
choosing either `~/`, or if `go-homedir` fails then either in `/tmp` or
[whatever Window's temporary directory](https://www.askvg.com/where-does-windows-store-temporary-files-and-how-to-change-temp-folder-location/)
is. If you specify an empty subdirectory name it will use
`.go-jsoncache` instead. 

## Cache File Format

JSON values are cached in JSON text files where the developer's JSON
value is wrapped in an outer JSON objects with both `data` and `expires`
properties. The values for the expiration property are generated as
follows using `RFC3339`:

```
duration := "600s"
dur, _ := time.ParseDuration(duration)
wrapper.expires := time.Now().Add(dur).Format(time.RFC3339)
```

## License
The `go-jsoncache` package is licensed via [AGPL-3.0](https://www.gnu.org/licenses/agpl-3.0.en.html).  Please [contact us](mailto:team@gearbox.works) if you would like to use this package under a different license.