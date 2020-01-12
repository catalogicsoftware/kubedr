# KubeDR Docs

The documentation for *KubeDR* is divided into two guides

- [User Guide](https://www.catalogicsoftware.com/)

- [Developer Guide](https://www.catalogicsoftware.com/)

We use [Sphinx](http://www.sphinx-doc.org/en/master/) to format and
build the documentation. The guides use
[Read the Docs](https://github.com/readthedocs/sphinx_rtd_theme)
theme.

## Installation

Here is one way to install Sphinx.

```bash
$ python3 -m venv ~/venv/sphinx
$ export PATH=~/venv/sphinx/bin:$PATH
$ pip install sphinx sphinx_rtd_theme

# For local builds, this helps in continuous build and refresh.
$ pip install sphinx-autobuild
```

## Build

```bash
$ cd docs/devguide
$ make html
```

This will generate HTML files in the directory ``html``. If you are
making changes locally and would like to automatically build and
refresh the generated files, use the following build command:

```bash
$ cd docs/devguide
$ sphinx-autobuild source build/html
```

## Guidelines

- The format for the documentation is
  [reStructuredText](http://www.sphinx-doc.org/en/master/usage/restructuredtext/index.html).

- The source for docs should be readable in text form so please keep
  lines short (80 chars). This will also help in checking diffs.
  
- Before checkin in or submitting a PR, please build locally and
  confirm that there are no errors or warnings from Sphinx.
