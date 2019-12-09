FROM alpine:3.8 AS runtime

# Install restic runtime dependencies and upgrade existing packages.
RUN apk update \
 && apk upgrade \
 && apk add --no-cache \
        ca-certificates \
        fuse \
        openssh \
        jq \
 && update-ca-certificates \
 && rm -rf /var/cache/apk/*

# Install restic and rclone.
COPY etcdctl /usr/local/bin
COPY restic /usr/local/bin
COPY kubectl /usr/local/bin
COPY kubedrutil.sh /usr/local/bin

# ENTRYPOINT ["/usr/local/bin/restic"]
