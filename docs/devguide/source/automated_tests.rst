=================
 Automated Tests
=================

It is very important that developers write automated tests at all
levels. The possible levels include:

- controller level.

- Integration tests that verify functionality from the user's point of
  view.

At present, the project has integration tests that are implemented
using the `pytest`_ framework.

Integration Tests
=================

Setup
-----

- Create a virtual environment and install pytest and other required
  modules.

  .. code-block:: bash

    $ python3 -m venv ~/venv/kubedr
    $ export PATH=~/venv/kubedr/bin:$PATH
    $ pip install pytest pytest-dependency kubernetes

- Set up a cluster that is accessible by the ``kubectl`` command.

- Install *KubeDR* and all its dependencies. More instructions will be
  provided later. For now, check "User Guide".

- The tests use `Python Kubernetes Client`_ to interact with the
  cluster. Currently, the  tests load the ``kubeconfig`` set up
  locally. So they work fine if ``kubectl`` works.

Running tests
-------------

Follow instructions above and make sure ``pytest`` is in ``PATH``.

Since the tests work with an existing cluster, they need to be provided
some config data (such as *S3* and *etcd* details). Such data is
passed in a file called ``tests/config/testenv.json``. For an example,
take a look at ``tests/config/sample_testenv.json``.

Here is a sample:

.. code-block:: json

   {
       "backuploc": {
           "endpoint": "http://10.106.44.180:9000",
           "access_key": "minio",
           "secret_key": "minio123"
       },
       "etcd_data": {
           "ca.crt": "/tmp/ca.crt",
           "client.crt": "/tmp/client.crt",
           "client.key": "/tmp/client.key"
       }
   }

If the config data is not provided, the tests will be skipped. If only
"backuploc" is provided, the tests will add the location and verify
that the bucket is properly initialized but skip the backup test.

Once the config data is ready, run the tests as follows:

.. code-block:: bash

  $ cd kcx/tests
  $ ./runtests

Please note that the tests will take noticeable time. This is because
the backup test needs to wait for at least one minute before it can
verify that backup pod is created and backup is done.

Configuring tests
-----------------

The behavior of tests can be controlled by various environment
variables. Currently, this facility is used to configure how tests
check k8s resources created by them.

Waiting for resources to appear
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

The tests need to wait for some resources such as *pods* and
*cronjobs* to appear. This is done by periodically polling to see if
the resource shows up. There are two env variables that control how
many times the polling is done and the interval between each such
attempt.

Here are the relevant env variables.

WAIT_FOR_RES_TO_APPEAR_NUM_ATTEMPTS
    Number of times the resource is checked. Default: 15.

WAIT_FOR_RES_TO_APPEAR_INTERVAL_SECS
    Interval between each poll attempt. Default: 1 second.

Waiting for Pod to be done
~~~~~~~~~~~~~~~~~~~~~~~~~~

In many cases, the tests need to wait for a Pod to be done (say,
backup). The following two env variables control this waiting.

WAIT_FOR_POD_TO_BE_DONE_NUM_ATTEMPTS
    Number of times the Pod status is checked. Default: 5.

WAIT_FOR_POD_TO_BE_DONE_INTERVAL_SECS
    Interval between each poll attempt. Default: 3 seconds.




.. _pytest: https://docs.pytest.org/en/latest/
.. _Python Kubernetes Client: https://github.com/kubernetes-client/python

