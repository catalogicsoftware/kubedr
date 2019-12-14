## Overview

Kubernetes stores all its objects in *etcd* so backing up data in
*etcd* is crucial for DR purposes. This project implements a tool that
backs up *etcd* data and certificates to any S3 bucket. It follows the
*operator* pattern that is popular in Kubernetes world.

An operator is basically a combination of *custom resources (CRs)*
coupled with *controllers* that manage the CRs. There would be one
controller for each CR. In addition to controllers, operators can also
contain *webhooks* that can be used to validate the data in resources
as well as to set defaults when some fields are not provided in the
input. Our operator uses webhooks for both these purposes.

For data transfer to S3, we currently use a tool called
[restic](https://restic.net) but we will be able to change the specific
backup tool in a backwards compatible manner.
