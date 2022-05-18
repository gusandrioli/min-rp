# **min-rp**
Minimal Reverse Proxy with different modes implemented in Golang

## Quick Setup
Check full examples at `examples/` directory
```go
rp.SetReverseProxyAndServe(
    &rp.SetReverseProxyAndServeOpts{
        Config: &rp.Config{
            ReverseProxy: rp.ReverseProxy{
                Port: "8080",
            },
            Workers: []*rp.Worker{
                {URL: "http://localhost:8081/"},
                {URL: "http://localhost:8082/"},
                {URL: "http://localhost:8083/"},
            },
            Type: ReverseProxyTypePathPrefix,
        },
    },
)
```

## Reverse Proxy Types
1. Path Matching - ReverseProxyTypePathPrefix
2. Round Robin   - ReverseProxyTypeRoundRobin

## Bugs
Bugs or suggestions? Open an issue [here](https://github.com/gusandrioli/min-rp/issues/new).