# Test DSL syntax

Test cases are written in Domain Specific Language (DSL) based on Golang. 

It is a vanilla GO with helper functions for common operations plus the
simplified string templates described below, which aim to improve readability
for non-GO developers that want to read or contribute tests.

## String templates

golang's default string formating package is similar to C. Format strings might look like `"this is a %s"` where `%s` is a verb that will be replaced at runtime.

These verbs collides with URL-escaping a lot, strings like `/ipfs/Qm.../%c4%85/%c4%99` might trigger weird errors. We implemented a minimal templating library that is used almost everywhere in the test.

It uses `{{name}}` as a replacement for `%s`. Other verbs are not supported.


```golang
Fmt("{{action}} the {{target}}", "pet", "cat") // => "pet the cat"
```

Backticks enable use of verbatim strings, without having to deal with golang-specific escaping of things like double quotes:

```golang
Fmt(`Etag: W/"{{etag-value}}"`, "weak-key") // => "ETag: W/\"weak-key\""
```

It is required to always provide a meaningful `{{name}}`:

```golang
Fmt(`/ipfs/{{cid}}/%c4%85/%c4%99`, fixture.myCID) // => "/ipfs/Qm..../%c4%85/%c4%99"
```

Values are replaced in the order they are defined, and you may reuse named values

```golang
Fmt(`<a href="{{cid}}">{{label}}}</a><a href="{{cid}}/index.html">index</a>`, fixture.myCID, "Link Title!") // => '<a href="Qm...">Link Title!</a><a href="Qm..../index.html">index</a>'
```

You may escape `{{}}` by using more than two opening or closing braces,

```golang
Fmt("{foo}") // => "{foo}"
Fmt("{{{foo}}}") // => "{{foo}}"
Fmt("{{{{foo}}}}") // => "{{{foo}}}"
Fmt("{{{foo}}}") // => {{foo}}
```

This templating is used almost everywhere in the test sugar, for example in request Path:

```golang
Request().Path("ipfs/{{cid}}", myCid) // will use "ipfs/Qm...."
```

