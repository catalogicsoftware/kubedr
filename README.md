# Kubernetes Cluster Backup

Welcome to KubeDR!

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
*KubeDR* project from Catalogic Software aims to provide complete end
to end data protection for Kubernetes data stored in *etcd*. In
addition, this project will backup certificates as well so if a master
needs to be rebuilt, all the data is available.

## Features

Here is a list of high level features provided by *KubeDR*. For more
details, please see
[User Guide](https://www.catalogicsoftware.com/).

- Backup cluster data in *etcd* to any S3 compatible storage.
- Backup certificates
- Pause and resume backups
- Clean up older snapshots based on a retention setting.
- Restore *etcd* snapshot
- Restore certificates

## Future

The following list shows many items that are planned for
*KubeDR*. Some of them are improvements while others are new
features. 

- Improve monitoring/reporting.

- Implement referential integrity semantics.

- Improve restore capabilities. 

- Support file system as a target for backups.

## Documentation

we use [Sphinx](http://www.sphinx-doc.org/en/master/) for docs. Source
for the documentation is in "docs" directory. For built documentation,
see below:

- [User Guide](https://www.catalogicsoftware.com/).

- [Developer Guide](https://www.catalogicsoftware.com/).

## Feedback

We would love to hear feedback from users. Please feel free to open
issues for bugs as well as for any feature requests. 

Please note that the project is in Alpha state so there may be many
corner cases where things may not work as expected. We are actively
working on fixing any bugs and on adding new features.
