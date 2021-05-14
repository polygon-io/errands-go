# Echo Errand

This simple errand is useful for testing that everything in your errands setup is hooked up properly.

### Environment Variables

- `ERRANDS_URL` - The URL of the errands server to poll for errands
- `ERRANDS_TOPIC` - The name of the topic to poll the errands server for. Defaults to `echo`

### Errand Parameters

- `echo` - A string that the errand processor will info log when processing your errand
- `fail` A boolean indicating whether or not the errand should fail or complete (`true` to fail, `false` or unset to complete)

### Building the Docker Image

From the root directory of this repository, run:

`docker build --tag <your-tag> -f cmd/echo/Dockerfile`