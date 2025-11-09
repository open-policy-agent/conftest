# Conftest documentation

The documentation for Conftest is stored as markdown files in this directory,
and a documentation site generated using [Mkdocs](https://www.mkdocs.org/).

Conftest provides a `Pipfile` for managing the required dependencies. At the top
level of this repository (ie. _not_ in the `docs` directory) run:

```console
pipenv install
```

With the dependencies installed you can run the site locally. Any modifications
to the files in `docs` will be automatically rebuilt.

```console
pipenv run mkdocs serve
INFO    -  Building documentation...
INFO    -  Cleaning site directory
INFO    -  Documentation built in 0.45 seconds
[I 200229 08:22:10 server:296] Serving on http://127.0.0.1:8000
INFO    -  Serving on http://127.0.0.1:8000
[I 200229 08:22:10 handlers:62] Start watching changes
INFO    -  Start watching changes
[I 200229 08:22:10 handlers:64] Start detecting changes
INFO    -  Start detecting changes
```
