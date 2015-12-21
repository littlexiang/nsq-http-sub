# nsq-http-sub

an nsqd & nsqlookupd http interface for 
- http polling style sub for nsq
- broadcast operations to all nsqd servers of cluster

originally from official tool nsq_pubsub with extended config and bug fix

https://github.com/nsqio/nsq/tree/master/apps/nsq_pubsub

## api
### pub
* POST /pub?topic=topicName
```curl -d "<message>" http://127.0.0.1:8080/pub?topic=topicName```

### sub
* GET /sub?topic=topicName&channel=channelName
* GET /stats

### channel operations
* POST /channel/create?topic=topicName&channel=channelName
* POST /channel/pause?topic=topicName&channel=channelName
* POST /channel/unpause?topic=topicName&channel=channelName
* POST /channel/empty?topic=topicName&channel=channelName
* POST /channel/delete?topic=topicName&channel=channelName

## usage
```
➜  nsq-http-sub git:(master) ✗ ./nsq-http-sub --help
Usage of ./nsq-http-sub:
  -http-address string
    	<addr>:<port> to listen on for HTTP clients (default "0.0.0.0:8080")
  -lookupd-http-address value
    	lookupd HTTP address (may be given multiple times)
  -max-in-flight int
    	max number of messages to allow in flight (default 100)
  -max-messages int
    	return if got N messages in a single poll (default 1)
  -timeout int
    	return within N seconds if maxMessages not reached (default 10)

```