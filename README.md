# Intro

What's the smallest amount of ceremony we can get away with, while still calling it cross language barrier RPC?

The whole point is shooting data between point a and b. Lets call the data "messages". Besides the actual data, we prolly want to send some metadata. We could embedd it into the data itself, but since we're almost certain to always send some... lets make it part of the message.

What's next you ask? Well, we know point a (that's where we stand), but where's point b? Do we care? How do we get to point b? Do we care? All we need is the address of point b, right? And a piece of data to send...

## Message definition

    Message struct {
    	Metadata map[string]string `json:"metadata,omitempty"`
    	Body     json.RawMessage   `json:"body,omitempty"`
    }

# RPC (this lib)

This is a lib designed to send messages from point a to point b. Atm the primary transport for doing so, is over NATS. A very capable messaging system.

In this lib, point b can take a number of forms. One is in the shape of a Worker (request/response) another is Eventer (event). A third shape is called Handler, and that blurs the line between Worker and Eventer. More about that later.

While we're at point a, the api needed to communicate with point b is fairly simple. In RPC it's a RpcClient and that api can trigger or request messages. At point b on the other hand, we need to collect some more information, to be able to handle the incomming message. At point b awaits a RpcServer.

## RpcServer

A RpcServer lets you expose address & methods pairs. An address is any topic you can come up with, ex. (```my.worker.topic```). Your method needs to implement one of the 3 method types to be usable as point b.

    Worker  func(*Message) *Message
    Eventer func(*Message)
    Handler func(Context)

## RpcClient

As long as we know the address (it's topic) of point b we can simply request an answer from it, or trigger a notification to it with the client.

# Nitty gritty...

Have a look in the examples folder.

[example/worker](https://github.com/Meduzz/rpc/blob/master/example/workers/app.go) - has an example of both a Worker method, and a Handler method acting as a Worker.

[example/print](https://github.com/Meduzz/rpc/blob/master/example/print/app.go) - has an example of the power of the Handle method type, as well as an Eventer type method.

# Run the worker example
    
    Terminal1
    cd example/proxy
    go run app.go

    Terminal2
    cd example/workers
    go run app.go

    post json to localhost:8080/rpc?action=echo&rpc=request
