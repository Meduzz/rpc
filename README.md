# rpc

Project to experiment with how far rpc over nats can be taken.

    POST /rpc?action=a.b.c
    .... body

Means put body on a.b.c queue, and expect an answer.

# Run
    
    Terminal1
    NATS_URL=nats://localhost:4222 go run proxy.go

    Terminal2
    go run example/echo.go

    post requests to localhost:8080 with action=echo
