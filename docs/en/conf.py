# Configuration file for the Sphinx documentation builder.
#
# This file only contains a selection of the most common options. For a full
# list see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html

# -- Path setup --------------------------------------------------------------

# If extensions (or modules to document with autodoc) are in another directory,
# add these directories to sys.path here. If the directory is relative to the
# documentation root, use os.path.abspath to make it absolute, like shown here.
#
# import os
# import sys
# sys.path.insert(0, os.path.abspath('.'))

# -- Project information -----------------------------------------------------
from jupyter_sphinx_theme import *

init_theme()


project = ''
copyright = '2021, Elena Aksenova, Nikita Lazarev, David Badalyan, Oleg Borisenko'
author = 'Elena Aksenova, Nikita Lazarev, David Badalyan, Oleg Borisenko'

# The full version, including alpha/beta/rc tags
release = 'v1.0.0'


# -- General configuration ---------------------------------------------------

# Add any Sphinx extension module names here, as strings. They can be
# extensions coming with Sphinx (named 'sphinx.ext.*') or your custom
# ones.
extensions = [
	'docxbuilder',
	"sphinx.ext.intersphinx",
    "sphinx.ext.autodoc",
    "sphinx.ext.mathjax",
    "sphinx.ext.viewcode",
]

# Add any paths that contain templates here, relative to this directory.
templates_path = ['_templates']
html_static_path = ["_static"]

# List of patterns, relative to source directory, that match files and
# directories to ignore when looking for source files.
# This pattern also affects html_static_path and html_extra_path.
exclude_patterns = ['_build', 'Thumbs.db', '.DS_Store']


# -- Options for HTML output -------------------------------------------------

# The theme to use for HTML and HTML Help pages.  See the documentation for
# a list of builtin themes.
#
# html_theme = 'alabaster'

# Add any paths that contain custom static files (such as style sheets) here,
# relative to this directory. They are copied after the builtin static files,
# so a file named "default.css" will overwrite the builtin "default.css".
# html_static_path = ['_static']

locale_dirs = ['./locale/']   # po files will be created in this directory
gettext_compact = False     # optional: avoid file concatenation in sub directories.

def setup(app):
    app.add_css_file('my-styles.css')

html_logo = "_static/logo.png"

html_theme_options = {
    "navbar_site_name": "",
    "navbar_links": [("Russian Version", "../ru/index")]
}

# navbar_links =[("Russian Version", "../ru/build/html/index")]
# import kotti_docs_theme
# html_theme = "kotti_docs_theme"
# html_theme_path = [kotti_docs_theme.get_theme_dir()]