==============
 Local Builds
==============

Operator
=================
The *KubeDR* operator resource can be built locally through utilizing the `Makefile` at the root of the repository, simply run `make build`, and *KubeDR*
will start both its local docker build and operator customization. The resulting output artifact will be the `kubedr.yaml` file under `kubedr/`.

Documentation
=================
Since *KubeDR* uses Sphinx to build its docs, first setup a python virtual environment:

.. code-block:: bash
  $ python3 -m venv ~/venv/sphinx
  $ export PATH=~/venv/sphinx/bin:$PATH
  $ pip install sphinx sphinx_rtd_theme

To build the *userguide*:

.. code-block:: bash
  $ cd docs/devguide
  $ make html

To build the *devguide*: 

.. code-block:: bash
  $ cd docs/userguide
  $ make html
