# nsq-http-sub

http polling style sub for nsq

originally from official tool nsq_pubsub with extended config and bug fix

https://github.com/nsqio/nsq/tree/master/apps/nsq_pubsub

## api

* GET /sub?topic=topicName&channel=channelName
* GET /stats
* [TODO] GET /length?topic=topicName&channel=channelName


## config
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
  -nsqd-tcp-address value
    	nsqd TCP address (may be given multiple times)
  -timeout int
    	return within N seconds if maxMessages not reached (default 10)

```

