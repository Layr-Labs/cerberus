# Remote Signer Implementation of cerberus-api
This is a remote signer which supports BLS signatures on the BN254 curve.


### Installation
#### Quick start
```bash
$ git clone https://github.com/Layr-Labs/cerberus.git
$ cd cerberus
$ make start
```

#### Manual
```bash
git clone https://github.com/Layr-Labs/cerberus.git
cd cerberus
go build -o bin/cerberus cmd/cerberus/main.go
./bin/cerberus 
```

### Usage options
| Options        | Description                                 | Default         |
|----------------|---------------------------------------------|-----------------|
| keystore-dir   | Directory to store encrypted keystore files | ./data/keystore |
| grpc-port      | gRPC port for starting signer server        | 50051           |
| log-format     | format of the logs (text, json)             | text            |
| log-level      | debug, info, warn, error                    | info            |
| metrics-port   | port to expose prometheus metrics           | 9091            |
| tls-ca-cert    | certificate to enable TLS connection        |                 |
| tls-server-key | server key to enable TLS connection         |                 |
| help           | show help                                   |                 |
| version        | show version                                |                 |


### Monitoring
The signer exposes prometheus metrics on the `/metrics` endpoint. You can scrape these metrics using a prometheus server.
There is a grafana dashboard available in the `monitoring` directory. You can import this dashboard into your grafana server to monitor the signer.

### Configuring Server-side TLS (optional)

Server-side TLS support is provided to encrypt traffic between the client and server. This can be enabled by starting the service with `tls-ca-cert` and `tls-server-key` parameters set:

#### Generating TLS certificates

For local testing purposes, the following commands can be used to generate a server certificate and key.

Create a file named `openssl.cnf` with the following content:

```
[ req ]
default_bits       = 2048
default_md         = sha256
default_keyfile    = server.key
prompt             = no
encrypt_key        = no

distinguished_name = req_distinguished_name
x509_extensions    = v3_req

[ req_distinguished_name ]
C            = US
ST           = California
L            = San Francisco
O            = My Company
OU           = My Division
CN           = localhost

[ v3_req ]
subjectAltName = @alt_names

[ alt_names ]
DNS.1 = localhost
```

```bash
# Generate the private key
openssl genpkey -algorithm RSA -out server.key

# Generate the certificate signing request (CSR)
openssl req -new -key server.key -out server.csr -config openssl.cnf

# Generate the self-signed certificate with SAN
openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt -extensions v3_req -extfile openssl.cnf

```

server.crt and server.key files can then be used to start the server with TLS support.

#### Starting the server with TLS support

```
cerberus -tls-ca-cert server.crt -tls-server-key server.key
```

The server can then be queried over a secure connection using a gRPC client that supports TLS. For example, using `grpcurl`:

```
grpcurl -cacert server.crt -d '{"password": "test"}' -import-path . -proto proto/keymanager.proto localhost:50051 keymanager.v1.KeyManager/GenerateKeyPair
```

#### Connecting a GO client with the server using TLS

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/Layr-Labs/cerberus-api/pkg/api/v1"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
)

func main() {
    creds, err := credentials.NewClientTLSFromFile("server.crt", "")
    if err != nil {
        log.Fatalf("could not load tls cert: %s", err)
    }

    conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
    if err != nil {
        log.Fatalf("did not connect: %v", err)
    }
    defer conn.Close()

    c := v1.NewSignerClient(conn)

    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    req := &v1.SignGenericRequest{
        PublicKey: "0xabcd",
        Password:  "p@$$w0rd",
        Data:      []byte{0x01, 0x02, 0x03},
    }
    resp, err := c.SignGeneric(ctx, req)
    if err != nil {
        log.Fatalf("could not sign: %v", err)
    }
    fmt.Printf("Signature: %v\n", resp.Signature)
}
```

### Migrating keys from eigenlayer-cli to cerberus
If you created your keys using the eigenlayer-cli,
you won't be able to directly copy the encrypted json file as this keystore uses ERC2335 format (eigenlayer-cli will add support for this soon).

You can migrate them to cerberus using the following steps:
1. Export your keys from eigenlayer-cli
    ```bash
    eigenlayer keys export --key-type bls <key-name>
    ```
2. Copy the private key from the output.
3. Import the key into cerberus
    ```bash
    grpcurl -plaintext -d '{"privateKey": "<pk>", "password": "p@$$w0rd"}' <ip>:<port> keymanager.v1.KeyManager/ImportKey
    ```

## Security Bugs
Please report security vulnerabilities to security@eigenlabs.org. Do NOT report security bugs via Github Issues.
