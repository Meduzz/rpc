# rpc

Project to experiment with how far rpc over nats can be taken.

    POST /rpc?action=a.b.c
    .... body

Means put body on a.b.c queue, and expect an answer.

# Run
    
    Terminal1
    cd example/proxy
    go run app.go

    Terminal2
    cd example/workers
    go run app.go

    post json to localhost:8080/rpc?action=echo
