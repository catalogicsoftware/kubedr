[![Go Report Card](https://goreportcard.com/badge/github.com/catalogicsoftware/kubedr)](https://goreportcard.com/report/github.com/catalogicsoftware/kubedr)
[![Discuss at kubedr-discuss@googlegroups.com](https://img.shields.io/badge/discuss-kubedr--discuss%40googlegroups.com-blue)](https://groups.google.com/d/forum/kubedr-discuss)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Open Source Love svg2](https://badges.frapsoft.com/os/v2/open-source.svg?v=103)](https://github.com/ellerbrock/open-source-badges/)

# Kubernetes Cluster Backup

Welcome to KubeDR!

![catalogic Logo](logos/logo-2.5-horiz-small.png)

## Overview

Kubernetes stores all the cluster data (such as resource specs) in
[etcd](https://etcd.io/). The *KubeDR* project implements data
protection for this data. In addition, certificates can be backed up
as well but that is optional.

**The project is currently in Alpha state and hence is not meant for
production use.**

## Rationale

There are projects and products that backup application data (stored
in *Persistent Volumes*) but there is no project that provides same
first class backup support for the very important Kubernetes cluster
data stored in *etcd*.

For sure, there are recipes on how to take *etcd* snapshot but the
*KubeDR* project from
[Catalogic Software](https://www.catalogicsoftware.com/)
aims to provide complete end to end data protection for Kubernetes
data stored in *etcd*. In addition, this project will backup
certificates as well so if a master needs to be rebuilt, all the data
is available.

## Features

Here is a list of high level features provided by *KubeDR*. For more
details, please see
[User Guide](https://www.catalogicsoftware.com/clab-docs/kubedr/userguide).

- Backup cluster data in *etcd* to any S3 compatible storage.
- Backup certificates
- Pause and resume backups
- Clean up older snapshots based on a retention setting.
- Restore *etcd* snapshot
- Restore certificates

## Roadmap

The following list shows many items that are planned for
*KubeDR*. Some of them are improvements while others are new
features.

- Improve monitoring/reporting.
- Support *Helm* installs.
- Implement referential integrity semantics.
- Improve restore capabilities.
- Support file system as a target for backups.

## Documentation

We use [Sphinx](http://www.sphinx-doc.org/en/master/) for docs. Source
for the documentation is in "docs" directory. For built documentation,
see below:

- [User Guide](https://catalogicsoftware.com/clab-docs/kubedr/userguide)
- [Developer Guide](https://catalogicsoftware.com/clab-docs/kubedr/devguide)

## Feedback

We would love to hear feedback from our users. Please feel free to open
issues for bugs as well as for any feature requests.

For any questions and discussions, please join us over at our Google Group:
[kubedr-discuss](https://groups.google.com/d/forum/kubedr-discuss).

Please note that the project is in Alpha so there may be many
corner cases where things may not work as expected. We are actively
working on fixing any bugs and on adding new features.