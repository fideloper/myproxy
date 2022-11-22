# My Proxy
This is my (reverse) proxy. There are many like it, but this one is mine.

I've written about [developing this reverse proxy](https://fideloper.com/go-http).

## TODO

Some ideas that might be fun to incorporate

- [x] multiple targets
    - Decide on upstream target based on hostname, uri, port, etc
- [x] Multiple listeners - port 80, 443, and whatever else we want to configur
- [x] graceful shutdown
- [ ] Dynamic configuration
    - graceful reloading
- [x] Multiple backends (AKA load balancing)
- [x] Passive Health checks
- [ ] Active Health Checks
- [ ] Dynamic behavior on incoming requests (e.g. send job to SQS)
- [ ] Dynamic behavior on returned requests (e.g. read a response header and replay request somewhere else)
- [ ] h2c (backend, or backend AND frontend)?
- [ ] fastcgi?
- [ ] WAF
- [ ] Logging
- [ ] Prometheus metrics
