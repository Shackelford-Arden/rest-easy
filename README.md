# rest-easy

This is a little project that I'm playing with to learn more about the Go HTTP stdlib.

## Goals

This project is me looking at writing a small framework for myself and my projects that uses 
as much of the Go standard library as possible while tacking on a few things on top of it that
I find helpful.

These are either features or me just discovering patterns that I like.

- "easy" nested endpoint grouping
- Support for stdlib middlewares
  - One in particular is getting the response status code when access logging
- Stronger typing for handler funcs
- Playing with the "Options" pattern that a lot of Go projects use
- Also looking into graceful shutdown
