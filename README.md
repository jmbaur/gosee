![](./gosee.png)

# GoSee

A minimal, live markdown previewer.

### Usage

```
$ gosee [-host <ip>:<port>] <file>
```

GoSee will watch a markdown file for changes and send updates to the browser
used for viewing. Static assets can be placed in a `static` directory, which
will be served by GoSee.

### Goals:

- [x] Github-flavored markdown styling
- [ ] Support for more filetypes (tex?)
