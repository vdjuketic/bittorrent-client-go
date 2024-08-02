# BitTorrent Client in Go

This is a simple BitTorrent client implemented in Go. It allows you to download files using the BitTorrent protocol.

## Features

[x] Download files from a BitTorrent torrent file
[x] Tracker communication to find peers
[x] Handling of peer connections and data exchange
[] Upload files to peers
[] Support for multiple torrents
[] CLI interface for easy usage
[] Efficient piece selection and download strategy


## Installation

To install the BitTorrent client, you need to have [Go](https://golang.org/doc/install) installed on your machine.

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/bittorrent-client-go.git
    ```
2. Navigate to the project directory:
    ```sh
    cd bittorrent-client-go
    ```
3. Build the project:
    ```sh
    go build
    ```

## Usage

The BitTorrent client can be used via the command line. Below are some basic commands to get you started:

### Downloading a Torrent

To download a file using a torrent, use the following command:

```sh
./bittorrent-client-go.exe download -output="/path/to/output" -torrent="/path/to/torrent/file"
