# Giki

----

Giki is a simple wiki which supports:

- [MathJax] for LaTeX formatting
- [multimarkdown] for markdown formatting
- git for page versioning
- [prettify] for syntax highlighting

It started as a translation from Racket of [Uiki](https://github.com/mattmight/uiki).

## Dependencies

- git
- multimarkdown

## Build and run

```
go get -u github.com/tidyoux/giki
cd giki
make
cd build/bin
./giki
```
