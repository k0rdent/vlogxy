# vlogxy - VictoriaLogs Proxy & Aggregator

**vlogxy** is an aggregator tool for [VictoriaLogs](https://victoriametrics.com/products/victorialogs/). It provides a unified query interface across multiple VictoriaLogs instances, enabling seamless log querying across multiple clusters or data centers. vlogxy intercepts queries, distributes them to all configured backend servers in parallel, aggregates the results, and returns a unified response.

## API

### Supported Endpoints

**Query endpoints:**

- `GET /select/logsql/query` - Stream query results
- `GET /select/logsql/stats_query` - Statistics query
- `GET /select/logsql/stats_query_range` - Time-range statistics
- `GET /select/logsql/hits` - Hit counts
- `POST /select/logsql/field_values` - Field values

**Management endpoints:**

- `GET /health` - Health check
- `GET /reload` - Reload configuration

> **Note:** Stream endpoints (`/stream/*`) and live tailing are not currently supported.

## Links

- [Local Development](/docs/dev.md)
- [Configuration File](/docs/server-group.md)
- [VictoriaLogs Documentation](https://docs.victoriametrics.com/victorialogs/)
- [LogQL Query Language](https://docs.victoriametrics.com/victorialogs/logsql/)
- [VictoriaMetrics](https://victoriametrics.com/)
