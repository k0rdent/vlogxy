# Server Group

Server group configuration defines backend VictoriaLogs instances that vlogxy aggregates queries from. Each server group represents one VictoriaLogs backend with connection settings, authentication, and HTTP client options.

## Configuration File

The configuration file location can be specified via:

* Environment variable `CONFIG_PATH`
* Command-line flag `--config` (overrides environment variable)

### Example

```yaml
server_groups:
  - target: "vlogs-1.example.com:9428"
    cluster_name: "us-east"
    scheme: "https"
    path_prefix: ""
    http_client:
      dial_timeout: "30s"
      tls_config:
        insecure_skip_verify: false
      basic_auth:
        username: "user"
        password: "pass"
```

### Reload Config

Configuration can be reloaded using the `/reload` endpoint, or automatically when using Helm chart after config file changes.
