# Overview

A lightweight observable lib. The problem of go channel is that is doesn't support unlimited buffer size,
it's a pain to decide what size to use, this lib will dynamically handle it.

- unlimited buffer size
- one publisher to multiple subscribers
- thread-safe
- subscribers never block each other
- publishing order is the same as listening order

## Examples

See [examples_test.go](examples_test.go).
