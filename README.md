# lake-info

CLI to fetch data from the US Army Corps of Engineers on Lake Levels. Currently only supports Table Rock Lake.

## Installation

```shell
go install github.com/patrickjmcd/lake-info@latest
```

## Configuration

The following environment variables are used to configure the application:

-   ATLAS_CONNECTION_URI - mongodb URI

## Usage

### Setup

The `setup` command will create the appropriate mongo collections for the data.

```shell
lake-info setup
```

### Scrape

The `scrape` command will gather the data from the US Army Corps of Engineers website and store it in the database. 
An optional `-A` or `--all` flag will force the storage of all available data on the specified lake.

```shell
lake-info scrape <lake name>
```

Example:

```shell
lake-info scrape tablerock -A
```

### Serve

The `serve` command will run the Connect API server to provide the data to a frontend application.
An optional `-P` or `--port` flag allows changing the port from the default of `8080`.

```shell
lake-info serve
```

