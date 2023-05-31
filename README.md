# **ZFSE** -- (Zone File Search Engine)

**Personalized Search, Infinite Exploration**

ZFSE is a single binary independent search engine that can be self-hosted on the cheapest & most budget friendly
servers.

## Menu

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Architecture & Configuration](#architecture--configuration)
- [Task Handlers](#task-handlers)
- [Development Guide](#development-guide)
- [Roadmap](#roadmap)

## Overview

Have you ever dreamed of building your own search engine with it's very own search index, but felt held back by the
complexities and expenses involved?

Fear not! ZFSE provides an effortless way to construct a self-hosted search engine through the power of a single binary,
putting the control of search indexing in your hands.

Building your own customized search engine is as easy as running `zfse run`

The main objectives of this project are:

- Developing a straightforward, single binary search engine complete with its own crawler, indexer, and rankers.
- Minimizing RAM & CPU requirements to ensure affordability and flexibility. ZFSE is capable of running on $5/month
  budget friendly servers.
- Offering a seamless customization experience through a single `config.toml` file, empowering users to quickly create
  their personalized search indexes.
- Presenting an intuitive plugin system that allows users to effortlessly incorporate new filters, indexers, and rankers
  into their own search engine. (🚧 Under Development)

## Quick Start

To initialize `zfse`, simply use the following command:

```bash
zfse init
```

This command creates all the necessary files and subdirectories for ZFSE.

Your folder structure should now look like this:

    .
    ├── zfse                        # Base ZFSE executable
    ├── config.toml                 # ZFSE configuration
    ├── cache                       # Various cache and output files produced by ZFSE
    ├   └── ...
    ├── plugins                     # Filter, Indexer, Ranker plugins for ZFSE
    └── zone-files                  # ICANN zone files to bootstrap ZFSE
        └── example_zone_file.txt   # Example zone file

Now, simply start ZFSE by running:

```bash
zfse run -query "MMORPG"
```

Command above will initiate a web crawl using the [example zone file](data/zone_files/example_zone_file.txt).

Once ZFSE is finished with the crawling and random ranking, you will have the results recorded
under `./cache/ranking/output.txt`

You are now ready to customize ZFSE and build your own search index!

ZFSE makes use of ICANN zone files to bootstrap the search index. First, you need to access and download the zone file
you're interested in from ICANN, using their Centralized Zone Data Service [here](https://czds.icann.org/home).

A good TLD to start with is `.dev`. Once you have access, extract the `tar.gz` archive from ICANN to the `./zone-files`
folder.

You can add as many TLDs as you want to `./zone-files`, and ZFSE will utilize them all automatically. However, keep in
mind that larger TLDs will require more disk space, particularly after crawling and indexing are complete.

Edit the `config.toml` to customize your search index. Refer to [configuration](#configuration).

**NOTE**: ZFSE will crawl the entire `.dev` TLD and record any `<meta description=.../>` tags it encounters. Therefore,
ensure that your host has enough available space (~30-40GB).

**WARNING**: It is strongly advised against running the complete crawl on your personal computer. Some Internet Service
Providers (ISPs) might create problems since ZFSE will initiate large amount of connections while crawling the entire
Top-Level Domain (TLD), so it is preferable to host ZFSE on a cloud provider.

Additionally, be aware of security concerns. ZFSE employs default GoLang library parsers to analyze crawled websites,
and security vulnerabilities within GoLang standard libraries could lead to the compromise of your host during the
crawl.

## Architecture & Configuration

ZFSE is configured by a single `config.toml` file.

ZFSE operates by reading the configuration files and executing the specified Task Handlers sequentially. The process is
divided into four main components:

1. **Pre-Crawl Filtering**: ZFSE initiates all pre-crawl filters defined by the `[[PreCrawlFilters]]` tag. Each
   pre-crawl filter processes the TLD zone file line by line, filtering the content and forwarding the output to the
   subsequent pre-crawl filters. These filters have the ability to discard or append new fields to a domain. Added
   fields can be accessed and utilized by subsequent Task Handlers.


2. **Crawling**: At the moment, ZFSE concentrates on crawling only the index page of websites. The crawler will
   initially verify the validity of the DNS record and determine if the website has a `robots.txt` file available. Next,
   it will parse the `robots.txt` file to see if the ZFSE agent is allowed to index the website. Subsequently, the
   crawler will capture the headers and HTML content of the website's index page and forward it to post-crawl filters.


3. **Post-Crawl Filtering**: Similar to pre-crawl filtering, post-crawl filters are designated by
   the `[[PostCrawlFilters]]` tag. The filters are executed sequentially, passing their output to the next post-crawl
   filter in line.


4. **Indexing**: Unlike filters, indexers create independent index databases without sending their output to the next
   indexer in line. Indexers can be configured using the `[Indexer]` tag.


5. **Ranking**: Rankers receive a user query and leverage the collected data from filters and indexers to determine the
   search result ranking. Rankers can be customized using the `[[Rankers]]` tag.

After the indexing process is complete, ZFSE will rely solely on Indexers and Rankers to produce search results.
Therefore, indexing a Top-Level Domain (TLD) just once is sufficient, unless there is a need to modify ZFSE's pre-index
configuration.

A default configuration file can be found [here](data/config/config.toml).

Most important options are:

- `concurrent_connections`: Controls the number of concurrent connections to use during the crawl. Adjust this setting
  according to the system's CPU, RAM, and the `ulimits` imposed by the operating system.

- `content_read_limit_in_bytes`: Specifies the amount of data the crawler should read and record. Adjust this setting to
  manage disk usage. ZFSE is capable of parsing half-way read HTML content.

## Task Handlers

* [Task Handlers](docs/task_handlers.md)

## Development Guide

* [Development Guide](docs/development.md)

## Roadmap

Please note that ZFSE is in its early stages of development (🚧). It is recommended to wait for the v0.7 release before
handling large TLDs like `.com`. The planned milestones are as follows:

- `v0.2`: Web UI
- `v0.3`: ICANN Zone File Downloader
- `v0.4`: `docker-compose.yml`
- `v0.5`: Additional Indexers & Rankers
- `v0.6`: Plugins
- `v0.7`: Ability to back up & restore unfinished worksets
- `v0.8`: Customizable configuration via Web UI
- `v0.9`: CPU Profilers (runtime/pprof) and final optimization pass
- `v1.0`: Increased unit test coverage & release
- `v1.x`: TBD
