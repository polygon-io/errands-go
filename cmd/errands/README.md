# CLI Overview

The `errands` CLI facilitates working with the errands API. It provides commands
such as `list` and `delete` and can even port-forward the errands service in our
k8s cluster so you don't have to!

# Installation

If you just want to use the tool as it is, you can install it by running
```bash
go install github.com/polygon-io/errands-go/cmd/errands@latest
errands help
```

If you've got the repo cloned and are dev-ing on the tool, you can build and run it locally:
```bash
go run ./cmd/errands/. help # assuming you're in this directory
```

or install your local version:
```bash
go install ./cmd/errands/. # assuming you're in this directory
errands help
```

# Usage

By default it will `kubectl port-forward` the errands service on `localhost:5555` but you can change the
port via `--port=XXX` or disable the port-forwarding entirely via `--bootstrap=false`.

If you disable bootstrapping then you'll need to provide the endpoint via `--endpoint=http://my-running-errands-endpoint.com`.

```bash
# to list all the failed or inactive sort-pparc errands
errands list --type=sort-pparc --status=failed,inactive

# if you're already port-forwarding the errands service on port 6000
errands list --type=sort-pparc --status=failed,inactive --bootstrap=false --port=6000

# to perform a dry-run delete of all the failed sort-pparc jobs
errands delete --type=sort-pparc --status=failed --dry-run=true

# to delete an errand by its ID
errands delete --id=abc-xyz-123
```
