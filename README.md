<h1 align="center">mailbox</h1>
<p align="center">
  <img alt="GitHub Actions Workflow Status" src="https://img.shields.io/github/actions/workflow/status/zeim839/mailbox/go.yml?label=Go%20Build"> <img alt="GitHub License" src="https://img.shields.io/github/license/zeim839/mailbox"> <img alt="GitHub issues" src="https://img.shields.io/github/issues/zeim839/mailbox">
</p>
<p align="center">
  <img src="https://i.imgur.com/zubHn5Y.png" width=700/>
</p>

Mailbox is a simple Go program for controlling website contact form submissions. It implements an API server backend for accepting form submissions and an intuitive, authenticated terminal UI for webmasters.
It currently works with MongoDB, but a modular database interface is provided for integrating with other backends. CloudFlare [Turnstile](https://www.cloudflare.com/en-gb/products/turnstile/) captchas are
also supported. Mailbox was designed as a minimal, non-properietary system for serving dynamic forms on static websites (e.g. GitHub pages).

## Install
> [!WARNING]
> Mailbox is work-in-progress. There are currently no stable releases.

This project uses the [Go compiler](https://go.dev/). To get started, clone the GitHub repository:
```bash
git clone https://github.com/zeim839/mailbox.git
cd mailbox
```

## Quick Start
Mailbox implements both a server and client executable. The server hosts a [CRUD](https://en.wikipedia.org/wiki/Create,_read,_update_and_delete) API for accepting and handling form submissions and the client
is a [TUI](https://en.wikipedia.org/wiki/Text-based_user_interface) for interacting with the server.

To install the server:
```bash
cd server
go build .
./server
```

The server executable must be ran with a present working directory which contains a configuration file (see [configure](#configure)).

To install the client:
```bash
cd cmd
go build .
./cmd
```

**Docker**

To build a dedicated docker container for the server, run:
```bash
docker build --tag 'mailbox' .
```

## Configure
The server is configured through a `config.env` file in the present working directory. The client application does not require any configuration. The following configuration options are defined for the server:
 * `MONGO_URI`: the MongoDB connection URI.
 * `GIN_MODE`: The Gin server mode (one of `DEBUG`, `RELEASE`, or `TEST`)
 * `PORT`: server port.
 * `USERNAME`: an optional username for implementing Basic http auth.
 * `PASSWORD`: the password for basic http auth.
 * `CAPTCHA_SECRET`: an optional secret API key for configuring Cloudflare Turnstile captcha.

A minimal configuration is illustrated below:
```env
# config.env
MONGO_URI = "mongodb+srv:// ... "
USERNAME  = "admin"
PASSWORD  = "strong-password-123"
```

Currently, only MongoDB is supported. We are working on implementing additional database backends.

## License
[MIT](LICENSE)

Copyright (C) 2024 Michail Zeipekki
