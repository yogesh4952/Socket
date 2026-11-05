# Socket

A simple TCP chat server written in Go.

## Overview

`Socket` is a minimal, concurrent TCP chat server that lets multiple clients connect, choose a unique username, and broadcast messages to everyone else in real time.

## Features

- **Concurrent connections** — each client is handled in its own goroutine.
- **Username registration** — clients are prompted for a username on connect; duplicates are rejected.
- **Broadcast messaging** — any message a client sends is forwarded to all other connected clients.
- **Thread-safe state** — client registry (by connection and by name) is guarded by a mutex.
- **Automatic cleanup** — disconnected clients are removed from the registry on exit.

## How it works

1. The server listens on TCP port `:8090`.
2. On connect, each client is asked to enter a username.
3. If the username is already taken, the client is rejected and the connection closes.
4. Otherwise the client is registered, and the server reads messages line-by-line.
5. Every incoming line is broadcast to all *other* registered clients as `username: message`.
6. When a client disconnects (EOF or error), it is removed from the registry.

## Running

```bash
go run .
```

The server starts listening on `:8090`.

## Connecting

Use any TCP client (e.g. `nc`, `telnet`, or a custom client):

```bash
nc localhost 8090
```

You will be prompted to enter a username, then can type messages that are broadcast to all other clients.

## Architecture

- `client` — holds a client's name and network connection.
- `Server` — maintains two maps (`byConn`, `byName`) for O(1) lookup, protected by a `sync.Mutex`.
- `handleConnection` — per-client goroutine: registers the client, then loops reading and broadcasting messages.
- `broadcast` — sends a message to every registered client except the sender.

## Requirements

- Go 1.27+