# Go Tiny Projects

A repository with small projects I'm doing to learn Golang.

The projects on this repository are based on the following contents:
 - Wesley Willians lessons from [Full Cycle Youtube channel](https://www.youtube.com/c/FullCycle/).

Also I've tried to follow some guidelines from [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md), like the interface compliance at compile time, function grouping and ordering.

# Project List

### order-processor
An order price calculator that consume messages from a RabbitMq queue and persists the order on a Sqlite database.
On this project I've implemented unity tests with [testify](https://github.com/stretchr/testify) using the "Test Suite" concept to keep "an state" between tests,  something that is familiar to me as TestFixtures on XUnit/C#.