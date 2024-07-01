# syntax=docker/dockerfile:1

ARG GO_VERSION="1.21"
ARG RUNNER_IMAGE="gcr.io/distroless/static-debian11"
ARG BUILD_TAGS="netgo,ledger,muslc"

# --------------------------------------------------------
# Builder
# --------------------------------------------------------

FROM golang:${GO_VERSION}-alpine3.18 as builder

ARG GIT_VERSION
ARG GIT_COMMIT
ARG BUILD_TAGS

RUN apk add --no-cache \
    ca-certificates \
    build-base \
    linux-headers

# Download go dependencies
WORKDIR /osmosis
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

# Cosmwasm - Download correct libwasmvm version
RUN ARCH=$(uname -m) && WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm | sed 's/.* //') && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc.$ARCH.a \
    -O /lib/libwasmvm_muslc.a && \
    # verify checksum
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm_muslc.a | grep $(cat /tmp/checksums.txt | grep libwasmvm_muslc.$ARCH | cut -d ' ' -f 1)

# Copy the remaining files
COPY . .

# Build osmosisd binary
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    GOWORK=off go build \
    -mod=readonly \
    -tags ${BUILD_TAGS} \
    -ldflags \
    "-X github.com/cosmos/cosmos-sdk/version.Name="symphony" \
    -X github.com/cosmos/cosmos-sdk/version.AppName="symphonyd" \
    -X github.com/cosmos/cosmos-sdk/version.Version=${GIT_VERSION} \
    -X github.com/cosmos/cosmos-sdk/version.Commit=${GIT_COMMIT} \
    -X github.com/cosmos/cosmos-sdk/version.BuildTags=${BUILD_TAGS} \
    -w -s -linkmode=external -extldflags '-Wl,-z,muldefs -static'" \
    -trimpath \
    -o /osmosis/build/osmosisd \
    /osmosis/cmd/osmosisd/main.go

# --------------------------------------------------------
# Runner
# --------------------------------------------------------

FROM ${RUNNER_IMAGE}

COPY --from=builder /osmosis/build/symphonyd /bin/symphonyd

# Install necessary packages and configure nginx
RUN sudo apt-get update && \
    sudo apt-get install -y nginx && \
    sudo cp ./symphonychain.conf /etc/nginx/sites-available/symphonychain.conf && \
    sudo ln -s /etc/nginx/sites-available/symphonychain.conf /etc/nginx/sites-enabled/ && \
    sudo nginx -t && \
    sudo systemctl reload nginx

# requires user interaction
# RUN sudo apt update && \
#     sudo apt install certbot python3-certbot-nginx && \
#     sudo certbot --nginx -d lcd.testnet.node2.symphonychain.org -d rpc.testnet.node2.symphonychain.org -d lcd.testnet.symphonychain.org -d rpc.testnet.symphonychain.org

# Configure iptables
RUN sudo iptables -A INPUT -p tcp --dport 26654 -s 127.0.0.1 -j ACCEPT && \
    sudo iptables -A INPUT -p tcp --dport 26654 -j DROP && \
    sudo iptables -A INPUT -p tcp --dport 1317 -s 127.0.0.1 -j ACCEPT && \
    sudo iptables -A INPUT -p tcp --dport 1317 -j DROP && \
    sudo iptables -A INPUT -p tcp --dport 9090 -j ACCEPT

ENV HOME /osmosis
WORKDIR $HOME

EXPOSE 26656
EXPOSE 26657
EXPOSE 1317
# Note: uncomment the line below if you need pprof in localosmosis
# We disable it by default in out main Dockerfile for security reasons
# EXPOSE 6060

ENTRYPOINT ["symphonyd"]
