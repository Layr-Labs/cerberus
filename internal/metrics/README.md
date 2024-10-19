## Metrics
It exposes the following metrics which can be scraped by Prometheus:

* `rpc_server_request_total`: The total number of RPC requests received by the server.
  * Labels: `method`
* `rpc_server_request_duration_seconds`: The duration of RPC requests in seconds.
  * Labels: `method`
  * Latency buckets for all methods
* `rpc_server_response_total`: The total number of RPC responses sent by the server.
  * Labels: `method` and `status` (e.g. `success`, `failed`).