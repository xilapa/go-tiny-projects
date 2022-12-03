# Go Tiny Projects

A repository with small projects I'm doing to learn Golang.

The projects on this repository are based/inspired on the following contents:
 - Wesley Willians lessons from [Full Cycle Youtube channel](https://www.youtube.com/c/FullCycle/).
 - Table Driven Tests from [Golang Wiki](https://github.com/golang/go/wiki/TableDrivenTests)

Also I've tried to follow some guidelines from [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md), like the interface compliance at compile time, function grouping and ordering.

# Project List

### order-processor
An order price calculator that consume messages from a RabbitMq queue and persists the order on a Sqlite database.
On this project I've implemented unity and integration tests using the "Test Suite" concept to have a way to do a Setup/Teardown for a group of tests, something that is familiar to me as TestFixtures on XUnit/C#.

### test-assertions
Simple test assertions that check if two values are equals or that a value is not an error. I've created this package after start using the standard tests lib on the "order-processor" project, to follow the DRY principle.