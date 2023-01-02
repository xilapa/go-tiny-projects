# Go Tiny Projects

A repository with small projects I'm doing to learn Golang. It's a work in progress.

The projects on this repository are based/inspired on the following contents:
 - Wesley Willians lessons from [Full Cycle Youtube channel](https://www.youtube.com/c/FullCycle/).
 - Table Driven Tests from [Golang Wiki](https://github.com/golang/go/wiki/TableDrivenTests)
 - Writing Web Applications from [Go Blog](https://go.dev/doc/articles/wiki/)

Also I've tried to follow some guidelines from [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md), like the interface compliance at compile time, function grouping and ordering.

# Project List

## order-processor
An order price calculator that consume messages from a RabbitMq queue and persists the order on a Sqlite database.
On this project I've implemented unity and integration tests using the "Test Suite" concept to have a way to do a Setup/Teardown for a group of tests, something that is familiar to me as TestFixtures on XUnit/C#.

To launch the app, set the rabbitmq url at `config/config.yml`. Then start the producer and the consumer:

    To start the consumer go to order-processor/cmd/consumer and run: `go run .`

    To start the producer go to order-processor/cmd/producer and run: `go run .`

To get the total orders processed make a get request to `localhost:8080/status`:

    `curl -i http://localhost:8080/status`

## test-assertions
Simple test assertions that check if two values are equals or that a value is not an error. I've created this package after start using the standard tests lib on the "order-processor" project, to follow the DRY principle.

## strong-rabbit
While working on the "order-processor" project I missed the auto-reconnect capability present on the official RabbitMq C# client. So I've made this package, it is a wrapper around the official one that gives the capability to auto-reconnect on failures and also to multiplex channels on connections.

The connections are stored in a pool and separated by their usage. Calling 'Connect()' multiple times with the same connection group will return a pointer to the same connection, it'll not open a new one.

Making it easy to have many channels on a single connection, while making it possible to have different connections to publish and consume, as the [official docs recommends](https://pkg.go.dev/github.com/rabbitmq/amqp091-go#Channel.Consume)
It can be used as the official one.

The StrongChannel can be of type Consumer or Publisher, each type is optimized to use less resources. The Publisher channel uses a background go-routine for reconnection while the Consumer channel don't. The Consumer channel listen's simultaneously to channel/connection errors while listening for new messages.

- Call 'Connect()' to ge a connection, with the connection call 'Channel()' passing a name and it's type to get an auto-reconnecting channel. Call 'Consume()' on the channel to start consuming or call any Publish method to publish a message.
- Publish confirms and channel QoS are restored on reconnection.
- Topology is not restored on reconnection.
- To stop the channel just call 'Close()', it's idempotent as the official.
- To stop a connection and remove it from the pool, call 'Close()' on the connection.

## go-wiki
A simple wiki server that store pages on a txt file. If the page doesn't exists, user is redirect to the edit page.
    To start the wiki go to go-wiki and run: `go run .` 
The wiki make use of the below lfu-cache to keep the most accessed pages in memory. These pages are listed at `localhost:8080/home`.


## lfu-cache
A in-memory cache with Least Frequently Used (LFU) eviction algorithm with O(1) time complexity. The algorithm is inspired on [this paper](https://www.researchgate.net/publication/355493987_An_O1_algorithm_for_implementing_the_LFU_cache_eviction_scheme), that proposes the use of a linked list of linked lists to store the cache items frequency. 
On this package, instead of using linked list of linked lists, a map with linked lists and one variable is used to keep track of items usage frequency.  Linked lists of cache items keys, are stored on the map, having the frequency as key. The lower usage frequency is stored on a variable.
