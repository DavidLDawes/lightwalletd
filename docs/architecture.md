# Definitions

A **light wallet** is not a full participant in the network of VerusCoin peers. It can send and receive payments, but does not store or validate a copy of the blockchain.

A **compact transaction** is a representation of a Zcash Sapling transaction that contains only the information necessary to detect that a given Sapling payment output is for you and to spend a note.

A **compact block** is a collection of compact transactions along with certain metadata (such as the block header) from their source block.

# Architecture

```
+----------+
|  zcashd  |                       +----------+    +-------+
+----+-----+              +------->+ frontend +--->+       |
     |                    |        +----------+    |  L    +<----Client
     | raw blocks    +----+----+                   |  O B  |
     v               |         |                   |  A A  |
+----+-----+         |         |   +----------+    |  D L  +<---Client
| ingester +-------->+ storage +-->+ frontend +--->+    A  |
+----------+ compact |         |   +----------+    |    N  +<-------Client
              blocks |         |                   |    C  |
                     +----+----+                   |    E  +<----Client
                          |        +----------+    |    R  |
                          +------->+ frontend +--->+       +<------Client
                                   +----------+    +-------+
```

## Ingester

The ingester is the component responsible for transforming raw Zcash block data into a compact block. See BlockIngestor in common/common.go, it saves blocks in the file system in the files db-main-blocks and db-main-lengths. It also caches blocks in memory as it sees them, making subsequent accesses faster.

The ingester is a modular component. Anything that can retrieve the necessary data and put it into storage can fulfill this role. Currently, the only ingester available communicates to verusd through RPCs and parses that raw block data. 

**How do I run it?**

⚠️ This section literally describes how to execute the binaries from source code. This is suitable only for testing, not production deployment. See section Production for cleaner instructions.

⚠️ Bringing up a fresh compact block database can take several hours of uninterrupted runtime.

First, install [Go >= 1.11](https://golang.org/dl/#stable). Older versions of Go may work but are not actively supported at this time. Note that the version of Go packaged by Debian stable (or anything prior to Buster) is far too old to work.

Now clone this repo and start the ingester. The first run will start slow as Go builds the sqlite C interface:

```
$ git clone https://github.com/asherda/lightwalletd
$ cd lightwalletd
$ swig -go  -intgosize 64 -c++ -cgo -gccgo -Wall -v parser/verushash/verushash.i
$ make
$ ./server --conf-file ~/.komodo/VRSC/VRSC.conf --log-file /logs/server.log --bind-addr 127.0.0.1:18232
```

To see the other command line options, run `./server --help`.

## Frontend

The frontend is the component that talks to clients. 

It exposes an API that allows a client to query for current blockheight, request ranges of compact block data, request specific transaction details, and send new Zcash transactions.

The API is specified in [Protocol Buffers](https://developers.google.com/protocol-buffers/) and implemented using [gRPC](https://grpc.io). You can find the exact details in [these files](https://github.com/davidldawes/lightwalletd/tree/master/walletrpc).

**How do I run it?**

⚠️ This section literally describes how to execute the binaries from source code. This is suitable only for testing, not production deployment. See section Production for cleaner instructions.

First, install [Go >= 1.11](https://golang.org/dl/#stable). Older versions of Go may work but are not actively supported at this time. Note that the version of Go packaged by Debian stable (or anything prior to Buster) is far too old to work.

Now clone this repo and start the frontend. The first run will start slow as Go builds the sqlite C interface:

```
$ git clone https://github.com/zcash/lightwalletd
$ cd lightwalletd
$ go run cmd/server/main.go --db-path <path to the same sqlite db> --bind-addr 0.0.0.0:9067
```

To see the other command line options, run `go run cmd/server/main.go --help`.

**What should I watch out for?**

x509 Certificates! This software relies on the confidentiality and integrity of a modern TLS connection between incoming clients and the front-end. Without an x509 certificate that incoming clients accurately authenticate, the security properties of this software are lost.

Otherwise, not much! This is a very simple piece of software. Make sure you point it at the same storage as the ingester. See the "Production" section for some caveats.

Support for users sending transactions will require the ability to make JSON-RPC calls to a zcashd instance. By default the frontend tries to pull RPC credentials from your zcashd.conf file, but you can specify other credentials via command line flag. In the future, it should be possible to do this with environment variables [(#2)](https://github.com/zcash/lightwalletd/issues/2).

## Storage

The storage provider is the component that caches compact blocks and their metadata for the frontend to retrieve and serve to clients.

The ingestor uses a couple of db* files and an in memory cache. The first time it is run (with no db* files present yet) it will start from the mid 2000 range and get all of the blocks up to the current top, then loop forever checking for new blocks every 40 seconds. Each time a block is requested it is cached if it wasn't already; if it wasn't cached then the block is requested from verusd directly (and cached).

**How do I run it?**

Commands above work fine.

If you want to see what it is up to then tail the log file. In the example above it's set to /logs//server.log, so use:
```
tal -f /logs/server.log
```
If you want to retest the loading, then stop it (hit a <Ctrl>C in it's terminal window, where you ran ./server), delete the db* files in /lightwalletd, and restart server. It will start fetching the blocks from verusd again, logging every 100 to the log file (see the tail command above).

**What should I watch out for?**

Startups are slow; provisioning with db* files harvested from up to date servers would speed things up. Actually using shared persistent sotrage - a DB, or redis or memcached, or perhaps eventually when there are huge numbers of chains Cassandra - would provide be better.

## Production

⚠️ This is informational documentation about a piece of alpha software. It has not yet undergone audits or been subject to rigorous testing. It lacks some affordances necessary for production-level reliability. We do not recommend using it to handle customer funds at this time (March 2019).

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

**Dependencies**

The first-order dependencies of this code are:

- Go (>= 1.11 suggested; older versions are currently unsupported)
- libsqlite3-dev (used by our sqlite interface library; optional with another datastore)

**Containers**

This software was designed to be container-friendly! We highly recommend that you package and deploy the software in this manner. We've created an example Docker environment that is likewise new and minimally tested, but it's functional.

**What's missing?**

lightwalletd currently lacks several things that you'll want in production. Caveats include:

- There are no monitoring / metrics endpoints yet. You're on your own to notice if it goes down or check on its performance.
- Logging coverage is patchy and inconsistent. However, what exists emits structured JSON compatible with various collectors.
- Logging may capture identifiable user data. It hasn't received any privacy analysis yet and makes no attempt at sanitization.
- The only storage provider we've implemented is sqlite. sqlite is [likely not appropriate](https://sqlite.org/whentouse.html) for the number of concurrent requests we expect to handle. Because sqlite uses a global write lock, the code limits the number of open database connections to *one* and currently makes no distinction between read-only (frontend) and read/write (ingester) connections. It will probably begin to exhibit lock contention at low user counts, and should be improved or replaced with your own data store in production.
- [Load-balancing with gRPC](https://grpc.io/blog/loadbalancing) may not work quite like you're used to. A full explanation is beyond the scope of this document, but we recommend looking into [Envoy](https://www.envoyproxy.io/), [nginx](https://nginx.com), or [haproxy](https://www.haproxy.org) depending on your existing infrastructure.
