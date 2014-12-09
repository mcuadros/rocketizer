rocketizer [![Build Status](https://travis-ci.org/mcuadros/rocketizer.png?branch=master)](https://travis-ci.org/mcuadros/rocketizer)
==============================

Painless Dockerfile transformation to [Rocket](https://github.com/coreos/rocket) containers.

Creates ACI Rocket containers using as source a [Dockerfile](https://docs.docker.com/reference/builder/), the `ENTRYPOINT`, `CMD`, `VOLUME`, `ENV` and `EXPOSE` expressions are used to create the manifest and the paths from `COPY` and `ADD` are compress into the [ACI container](https://github.com/coreos/rocket/tree/master/app-container) in the `rootfs` directory. Other expressions are ignored.

Based on the hype of this days around `containers` has been carry to me build this small tool. Maybe is now is useless but maybe someday or in some point could become usefull for someone.

> Warning: Since the `FROM` expresion is ignored, the generated containers are useless when the binaries are not copy with a `ADD` or `COPY` expressions and statically  compiled


Installation
------------

```
wget https://github.com/mcuadros/rocketizer/releases/download/v0.1.1/rocketizer_v0.1.1_linux_amd64.tar.gz
tar -xvzf rocketizer_v0.1.1_linux_amd64.tar.gz
cp rocketizer_v0.1.1_linux_amd64/rocketizer /usr/local/bin/
```

browse the [`releases`](https://github.com/mcuadros/rocketizer/releases) section to see other archs


Usage
-----

Based on the following Dockerfile:

```docker
FROM golang

# Copy outyet command inside the container.
ADD outyet /go/bin/outyet

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/outyet

# Document that the service listens on port 8080.
EXPOSE 8080
```

The following command generate a ACI container from the previous Dockfile:

```sh
./rocketizer convert --name outyet --version 1.0.0
Building outyet<1.0.0>
Parsing Dockerfile Dockerfile
Compressing files... OK
New ACI created outyet-v1.0.0-linux-amd64.aci
```

Create a `outyet-v1.0.0-linux-amd64.aci` container with the following contents:
```
rootfs/go/bin/outyet
app
```

The `app` manifest contains:

```json
{
  "acVersion": "1.0.0",
  "acKind": "AppManifest",
  "name": "outyet",
  "version": "1.0.0",
  "os": "linux",
  "arch": "amd64",
  "exec": [
    "\/go\/bin\/outyet"
  ],
  "eventHandlers": null,
  "user": "",
  "group": "",
  "environment": null,
  "mountPoints": null,
  "ports": [
    {
      "name": "8080",
      "protocol": "tcp",
      "port": 8080,
      "socketActivated": false
    }
  ],
  "isolators": null,
  "annotations": null
}
```

License
-------

MIT, see [LICENSE](LICENSE)
