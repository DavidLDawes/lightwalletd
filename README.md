# Disclaimer
This is an experimental build and is currently under active development. Please be advised of the following:

- This code currently is not audited by an external security auditor, use it at your own risk
- The code **has not been subjected to thorough review** by engineers at the Electric Coin Company or anywhere else
- We **are actively changing** the codebase and adding features where/when needed on multiple forks

On initial startup lightwalletd starts loading all the blocks starting from block 1. As they are loaded they are added to a levelDB name/value DB store for fast scalable access. Records are stored by block height and by block hash, so blocks can be looked up by either value.

Once all blocks are added, lightwalletd continues waiting for more blocks, adding them as they occur. If lightwalletd is stopped then restrarted, it picks up where it left off using lightwalletd and continues ingesting new blocks as they occur.

🔒 Security Warnings

The Lightwalletd Server is experimental and a work in progress. Use it at your own risk. Developers should familiarize themselves with the [wallet app threat model](https://zcash.readthedocs.io/en/latest/rtd_pages/wallet_threat_model.html), since it contains important information about the security and privacy limitations of light wallets that use Lightwalletd.

---

# Overview

[lightwalletd](https://github.com/Asherda/lightwalletd) is a backend service that provides a bandwidth-efficient interface to the VerusCoin blockchain. Currently, lightwalletd supports the Sapling protocol version as its primary concern. The intended purpose of lightwalletd is to support the development of mobile-friendly shielded light wallets.

The VerusCoin developers ported lightwalletd to the VerusCoin VRSC chain. This version uses verusd rather than zcashd and implements the new VerusCoin hashing algorithms, up to and including V2b2. 

Lightwalletd has not yet undergone audits or been subject to rigorous testing. It lacks some affordances necessary for production-level reliability. We do not recommend using it to handle customer funds at this time (April 2020).

Documentation for lightwalletd clients (the gRPC interface) is in `docs/rtd/index.html`. The current version of this file corresponds to the two `.proto` files; if you change these files, please regenerate the documentation by running `make doc`, which requires docker to be installed. 
# Local/Developer docker-compose Usage
Note: when using docker, map the data directory input via CLI to a location on your file system, so that the data persists even if the containers are destroyed. This avoids reloading everything every time.

[docs/docker-compose-setup.md](./docs/docker-compose-setup.md)

# Local/Developer Usage
Added leveldb support for storing local chain and tx data, replacing the flat file with simple indexing approach. It's included automatically and uses the normal command line options so no change should be needed, existing configurations will continue working.

## Testing
Fixed the unit tests so they all pass. Removed a couple but mostly got them repaired.

Simply run make test to run all the tests:
```
~/levelDB/lightwalletd$ make test
go test -v ./...
# github.com/Asherda/Go-VerusHash
verushash.cxx: In member function ‘void Verushash::initialize()’:
verushash.cxx:21:20: warning: ignoring return value of ‘int sodium_init()’, declared with attribute warn_unused_result [-Wunused-result]
         sodium_init();
         ~~~~~~~~~~~^~
=== RUN   TestHashV2b2
Got the correct v2b2 hash for block 1053660
Got the correct v2b2 hash for block 1053661!
--- PASS: TestHashV2b2 (0.00s)

<deleted lots of passing test results>

=== RUN   TestString_read
--- PASS: TestString_read (0.00s)
PASS
ok  	github.com/Asherda/lightwalletd/walletrpc	(cached)

```
## Code Coverage
If you want to measure unit test coverage of the code run this go test command from the project's root diretory:
```
~/levelDB/lightwalletd$ go test $(go list ./...) -coverprofile coverage.out
# github.com/Asherda/Go-VerusHash
verushash.cxx: In member function ‘void Verushash::initialize()’:
verushash.cxx:21:20: warning: ignoring return value of ‘int sodium_init()’, declared with attribute warn_unused_result [-Wunused-result]
         sodium_init();
         ~~~~~~~~~~~^~
ok  	github.com/Asherda/lightwalletd	0.007s	coverage: 0.0% of statements
ok  	github.com/Asherda/lightwalletd/cmd	0.008s	coverage: 34.1% of statements
ok  	github.com/Asherda/lightwalletd/common	11.213s	coverage: 40.4% of statements
ok  	github.com/Asherda/lightwalletd/common/logging	0.006s	coverage: 91.7% of statements
ok  	github.com/Asherda/lightwalletd/frontend	14.693s	coverage: 49.5% of statements
ok  	github.com/Asherda/lightwalletd/parser	0.520s	coverage: 94.6% of statements
ok  	github.com/Asherda/lightwalletd/parser/internal/bytestring	0.003s	coverage: 100.0% of statements
?   	github.com/Asherda/lightwalletd/testclient	[no test files]
?   	github.com/Asherda/lightwalletd/testtools/genblocks	[no test files]
?   	github.com/Asherda/lightwalletd/testtools/zap	[no test files]
ok  	github.com/Asherda/lightwalletd/walletrpc	0.014s	coverage: 3.1% of statements
```
Once that runs you can take a look at coverage while viewing the source code by running:
```
go tool cover -html=coverage.out
```
## Multichain
We can put chains into separate DB files (effectively a DB per chain) or we can use a single DB and add a chain indication to the key.

We'll work that out, for now this gets us live on levelDB which, even with two writes per record (block by height and max height) is about three times as fast as the prior flat file method.
## LevelDB
Switching from the simple two file index & serialzed compact data approach to using levelDB via [goleveldb](https://github.com/syndtr/goleveldb.git).

This gives us better performance at large scale. We support looking blocks up by block height. With 2 writes per block on my dev machine I'm getting about 3.4K blocks every 4 seconds. The old schem was more like 1.7K.
### Progress - Max Block Height
To simplify housekeeping, we record the highest block cached in leveldb. On restart this allows us to resume where we left off and avoid rescanning. We check that the new block's prior_hash matches our recorded hash for the last cached block, so if we get a reorg we will notice and rewind and re-cache the data.

### Corruption Check
Every block record is prepended by an 8 byte checksum for the block that is calculated when we store it. Each time we get a value we redo the checksum to ensure nothing has been corrupted.
### Reorg and --redownload
If calculated hashes don't line up with prevHash from the next block we assume we hit a chain reorg and rewind, getting the blocks over again, checking the hashes and rewinding up to 100 blocks before giving up.

The key value store is idempotent, so as soon as we write a new value for a given height the old one is gone. There's the usual small risk of data loss due to failures since we do not sync on writes, but the system notices corrupted blocks and hash mismatches and automatically corrects for them, so it's pretty resilient.

Reorgs work on the most recent blocks, no more than 100 of them, presumably due to forks. If you want to simply flush the levelDB data, use the --redownload flag.

The --redownload flag on the command line makes lightwalletd flush the levelDB and reload from scratch. Note that we need to delete all the records previously stored for the block before adding a new one. Since we are single threaded and single process, and there is a single record per key type, this works fine. It takes about 8 seconds to delete them all on the current VerusCoin chain, wkich has a bit over 1M records in August 2020.

We have a utility function to flush ranges of blocks in cache.go called flushBlocks(first int, last int)
### Schema
We ingest the blockchain data and store the results. A siplified view of the result:
An array of blocks
- Each block is serialized into a single []byte array called a compactBlock
- The block contains block details and an array of TXs, each of which is a compactTX. Each TX contains arrays of spends and outputs; all are serialzied into a single array.
- When storing the block, we save it under Bnnnnnnnn where nnnnnnnn is the block height.
- When we store a new "latestBlock" we store the height under the key Icccccccc where cccccccc is the chainID.
## Verusd

You must start a local instance of `verusd`, and its `VRSC.conf` file must include the following entries
(set the user and password strings accordingly):
```
txindex=1
insightexplorer=1
experimentalfeatures=1
rpcuser=xxxxx
rpcpassword=xxxxx
```

verusd can be configured to run `mainnet` or `testnet` (or `regtest`). If you stop `verusd` and restart it on a different network (switch from `testnet` to `mainnet`, for example), you must also stop and restart lightwalletd.

It's necessary to run `verusd --reindex` one time for these options to take effect. This typically takes several hours, and requires more space in the data directory.

Lightwalletd uses the following `verusd` RPCs:
- `getblockchaininfo`
- `getblock`
- `getrawtransaction`
- `getaddresstxids`
- `sendrawtransaction`

We plan on extending it to include identity and token options now that those are available (identity) or becoming available (tokens in may 2020).
## Lightwalletd
Install [Cmake](https://cmake.org/download/)

Install [Boost](https://www.boost.org/)

Install [Go](https://golang.org/dl/#stable) version 1.11 or later. You can see your current version by running `go version`.

Clone the [current repository](https://github.com/zcash/lightwalletd) into a local directory that is _not_ within any component of
your `$GOPATH` (`$HOME/go` by default), then build the lightwalletd server binary by running `make`.

## To run SERVER

Assuming you used `make` to build the server, here's a typical developer invocation:

```
./lightwalletd --log-file /logs/server.log --grpc-bind-addr 127.0.0.1:18232 --verusd-conf-path VRSC.conf --data-dir .
```
Type `./lightwalletd help` to see the full list of options and arguments.

Note that the --zcash-conf-path option is still listed but it doesn't do anything at the moment.
# Production Usage

Run a local instance of `zcashd` (see above), except do _not_ specify `--no-tls-very-insecure`.
Ensure [Go](https://golang.org/dl/#stable) version 1.11 or later is installed.

**x509 Certificates**
You will need to supply an x509 certificate that connecting clients will have good reason to trust (hint: do not use a self-signed one, our SDK will reject those unless you distribute them to the client out-of-band). We suggest that you be sure to buy a reputable one from a supplier that uses a modern hashing algorithm (NOT md5 or sha1) and that uses Certificate Transparency (OID 1.3.6.1.4.1.11129.2.4.2 will be present in the certificate).

To check a given certificate's (cert.pem) hashing algorithm:
```
openssl x509 -text -in certificate.crt | grep "Signature Algorithm"
```

To check if a given certificate (cert.pem) contains a Certificate Transparency OID:
```
echo "1.3.6.1.4.1.11129.2.4.2 certTransparency Certificate Transparency" > oid.txt
openssl asn1parse -in cert.pem -oid ./oid.txt | grep 'Certificate Transparency'
```

To use Let's Encrypt to generate a free certificate for your frontend, one method is to:
1) Install certbot
2) Open port 80 to your host
3) Point some forward dns to that host (some.forward.dns.com)
4) Run
```
certbot certonly --standalone --preferred-challenges http -d some.forward.dns.com
```
5) Pass the resulting certificate and key to frontend using the -tls-cert and -tls-key options.

## To run production SERVER

Example using server binary built from Makefile:

```
./lightwalletd --tls-cert cert.pem --tls-key key.pem --verus-conf-file VRSC.conf --log-file /logs/server.log --grpc-bind-addr 127.0.0.1:18232
```

## Block cache

Lightwalletd caches all blocks from Sapling activation up to the
most recent block, which takes about an hour the first time you run
lightwalletd. During this syncing, lightwalletd is fully available,
but block fetches are slower until the download completes.

After syncing, lightwalletd will start almost immediately,
because the blocks are cached in local files (by default, within
`/var/lib/lightwalletd/db`; you can specify a different location using
the `--data-dir` command-line option).

Lightwalletd checks the consistency of these files at startup and during
operation as these files may be damaged by, for example, an unclean shutdown.
If the server detects corruption, it will automatically re-downloading blocks
from `verusd` from that height, requiring up to an hour again (no manual
intervention is required). But this should occur rarely.

If lightwalletd detects corruption in these cache files, it will log
a message containing the string `CORRUPTION` and also indicate the
nature of the corruption.

## Darksidewalletd & Testing

Lightwalletd now supports a mode that enables integration testing of itself and
wallets that connect to it. See the [darksidewalletd
docs](docs/darksidewalletd.md) for more information.

# Pull Requests

We welcome pull requests! We like to keep our Go code neatly formatted in a standard way,
which the standard tool [gofmt](https://golang.org/cmd/gofmt/) can do. Also, run golint
prior to checkin and keep things clean.

Our current PR template asks for a design document link from the PR description and a
test plan added as a comment. If no design is needed (i.e. a README update or depenency
version update) then don't check off the box, explain the exception. Ditto the test plan.

 Please consider adding the following to the
file `.git/hooks/pre-commit` in your clone:

```
#!/bin/sh

modified_go_files=$(git diff --cached --name-only -- '*.go')
if test "$modified_go_files"
then
    need_formatting=$(gofmt -l $modified_go_files)
    if test "$need_formatting"
    then
        echo files need formatting (then don't forget to git add):
        echo gofmt -w $need_formatting
        exit 1
    fi
fi
```

You'll also need to make this file executable:

```
$ chmod +x .git/hooks/pre-commit
```

Doing this will prevent commits that break the standard formatting. Simply run the
`gofmt` command as indicated and rerun the `git add` and `git commit` commands.
