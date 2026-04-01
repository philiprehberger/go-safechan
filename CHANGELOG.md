# Changelog

## 0.2.1

- Standardize README to 3-badge format with emoji Support section
- Update CI checkout action to v5 for Node.js 24 compatibility
- Add GitHub issue templates, dependabot config, and PR template

## 0.2.0

- Add `Drain` and `DrainCtx` for non-blocking collection of buffered channel values
- Add `Filter` for forwarding only values matching a predicate
- Add `Map` for transforming channel values with type conversion support
- Add `SendTimeout` and `RecvTimeout` for deadline-based send/receive

## 0.1.2

- Consolidate README badges onto single line

## 0.1.1

- Add badges and Development section to README

## 0.1.0

- Initial release
- Safe `Send` and `SendCtx` that never panic on closed channels
- Context-aware `RecvCtx`
- `FanIn`, `FanOut`, and `Broadcast` channel combinators
