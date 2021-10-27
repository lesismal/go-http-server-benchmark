# go-http-server-benchmark

- The more connections, nbio cost the less memory, and performance the better than other frameworks.

- We can serve for [1000k or more connections](https://github.com/lesismal/nbio_examples/tree/master/websocket_1m) using nbio while other frameworks may have been OOM.


## 1k connections
<img width="676" alt="c1000" src="https://user-images.githubusercontent.com/40462947/138992121-5d83f82b-a75d-4968-a6dc-46a4afd57ba2.PNG">

## 5k connections
<img width="702" alt="c5000" src="https://user-images.githubusercontent.com/40462947/138992136-337b1e5f-03e8-435a-acee-11fa5c33350f.PNG">

## 20k connections
<img width="679" alt="c20000" src="https://user-images.githubusercontent.com/40462947/138992146-2e4d1574-4283-466c-b656-0f0c258a1df2.PNG">

## 30k connections
<img width="717" alt="c30000" src="https://user-images.githubusercontent.com/40462947/138992154-551f5571-4cc7-49ba-b1ac-8f4b9dec3df8.PNG">
