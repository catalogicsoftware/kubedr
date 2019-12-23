==============
 Introduction
==============

The *Kubedr* project implements data protection for metadata of
Kubernetes stored in *etcd*. In addition, certificates can be backed
up as well but that is optional.

Documentation
=============

This guide is built using `sphinx`_ and uses `Read the Docs`_
theme.

Installation
------------

  .. code-block:: bash

    $ python3 -m venv ~/venv/sphinx
    $ export PATH=~/venv/sphinx/bin:$PATH
    $ pip install sphinx sphinx_rtd_theme

    # For local builds, this helps in continuous build and refresh.
    $ pip install sphinx-autobuild

Building
--------

  .. code-block:: bash

    $ cd docs/devguide
    $ make html

This will generate HTML files in the directory ``html``. If you are
making changes locally and would like to automatically build and
refresh the generated files, use the following build command:

  .. code-block:: bash

    $ cd docs/devguide
    $ sphinx-autobuild source build/html


.. _sphinx: http://www.sphinx-doc.org/en/master/index.html
.. _Read the Docs: https://github.com/readthedocs/sphinx_rtd_theme


