# Overview

A lightweight observable lib.

- one publisher to multiple subscribers
- thread-safe
- subscribers never block each other
- publishing order is the same as listening order

The trade-off is the subscribers may delay sometime (between 1ns to 1ms) when new events come.
