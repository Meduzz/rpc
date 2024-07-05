# RPC

This wrapper will let you do simple json rpc over nats.

## Good news

With v2 the focus is not only RPC, but also Events. This change brings one type of context for each type. With oppinions on what you would want to do in each situation.

## The tools

*One... is fine*
The quickstart relies on that everything needed to connect to Nats is provided by you. That entails a NATS connection and a codec (`codec.Json()`). With that set you can go straight for desert by using the default RPC instance with `rpc.HandleRPC` and `rpc.HandleEvent`.

*Manual, with full control*
If you're more of a manual setup kind of person, then bring a nats connection. Call `rpc.NewRpc` with the connection and a codec. The only provided codec is JSON (`encoding.Json()`) which is also used in the default RPC instance.