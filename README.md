# go_homepage [![Build Status](https://travis-ci.com/s12chung/go_homepage.svg?branch=master)](https://travis-ci.com/s12chung/go_homepage)

A static site generator for https://stevenchung.ca written with [`gostatic`](https://github.com/s12chung/gostatic). `gostatic` is the generic code extracted from this project.

It has:
- A homepage of blog post listings
- Blog posts written in Markdown
- An atom feed of blog posts
- Reading page full of Goodreads reviews
- About page written in Markdown

Goodreads reviews are retrieved via API and cached locally. See [`gostatic`](https://github.com/s12chung/gostatic) for usage.
