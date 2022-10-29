# Some Proxy
My own reverse proxy. Feelin' cute, might delete later.

I've written about [developing this reverse proxy](https://fideloper.com/go-http).

## TODO

Some ideas that might be fun to incorporate

- [x] multiple targets
    - Decide on upstream target based on hostname, uri, port, etc
- [x] Multiple listeners - port 80, 443, and whatever else we want to configure
- [x] graceful shutdown
- [ ] Dynamic configuration
    - graceful reloading
- [ ] Multiple backends (e.g. load balancing)
- [ ] Health checks
- [ ] Dynamic behavior on incoming requests (e.g. send job to SQS)
- [ ] Dynamic behavior on returned requests (e.g. read a response header and replay request somewhere else)
- [ ] h2c (backend, or backend AND frontend)?
- [ ] fastcgi?
- [ ] WAF