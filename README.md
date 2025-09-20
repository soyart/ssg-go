# ssg (static site generator)

ssg is a Markdown static site generator.

ssg generates a website from directory tree,
with Markdown files being converted and assembled into HTMLs.

## Usage

```sh
ssg <src> <dst> <title> <url>
```

- Files or directories whose names start with `.` are ignored.

  Files listed in `${src}/.ssgignore` are also ignored in a fashion similar
  to `.gitignore`. To see how `.ssgignore` works in ssg-go, see
  [the test `TestSsgignore`](./ssg_test.go).

- Files with extensions other than `.md` and `html` will simply be copied
  into mirrored `${dst}`.

  If we have `foo.html` and `foo.md`, the HTML file wins.

- ssg reads Markdown files under `${src}`, converts each to HTML,
  and prepends and appends the resulting HTML with `_header.html`
  and `_footer.html` respectively.

  The assembled output file is then mirrored into `${dst}`
  with `.html` extension.

- In the end, ssg generates metadata such as `${dst}/sitemap.xml` with data
  from the CLI parameter and the output tree, and `${dst}/.files` to remember
  what files it had processed.

- HTML tags `<head><title>` is extracted from the first Markdown h1 (default),
  or a default value provided at the command-line (the 3rd argument).

  > With ssg-go, the titles can also be extracted from special line starting
  > with `:ssg-title` tag. This line will be removed from the output.

# ssg-go

ssg-go is a drop-in replacement and library for replacing
[romanzolotarev's implementation of ssg](https://romanzolotarev.com/bin/ssg).

ssg-go provides an executable and a Go library. [soyweb](https://github.com/soyart/soyweb)
is another ssg-go sister project which extends ssg-go via `Option`.

## Differences between ssg and ssg-go

### ssg-go ignores `.files`

In the original ssg, filenames listed in `.files` are
ignored and not re-generated. Unlike the original ssg, ssg-go ignores `${dst}/.files`
simply because it adds needless complexity.

By ignoring `.files`, we can be sure that the output directory is generated in a
functional fashion, i.e. we'll always get the same output with the same source material.

To do caching from previous run, an option to store file hashes in `${dst}/.files.sha256`
seems attractive. But upon closer inspection, it seems problems will arise when people
use ssg-go with other wrappers that read other files or do substitutions.

### ssg-go concurrent writers

ssg-go has built-in concurrent output writers.

Environment variable `SSG_WRITERS` sets the number of concurrent writers in ssg-go,
i.e. at any point in time, at most `SSG_WRITERS` number of threads are writing output
files.

The default value for concurrent writer is 20. If the supplied value is illegal,
ssg-go falls back to 20 concurrent writers.

> To write outputs sequentially, set the write concurrency value to 1:
>
> ```shell
> SSG_WRITERS=1 ssg mySrc myDst myTitle myUrl
> ```

### ssg-go custom title tag for `_header.html`

ssg-go also parses `_header.go` for title replacement placeholder.
Currently, ssg-go recognizes 2 placeholders:

- `{{from-h1}}`

  This will prompt ssg-go to use the first Markdown line starting with `#` value as head title.
  For example, if this is your Markdown:

  ```markdown
  ## This is H2

  # This is H1

  :ssg-title This is also an H1

  This is paragraph
  ```

  then `This is H1` will be used as the page's title.

- `{{from-tag}}`

  Like with `{{from-h1}}`, but finds the first line starting with `:ssg-title` instead,
  i.e. `This is also an H1` from the example above will be used as the page's title.

  > Note: `{{from-tag}}` directive will make ssg look for pattern `:ssg-title YourTitle\n\n`,
  > so users must always append an empty line after the title tag line.

For example, consider the following header/footer templates and a Markdown page:

```html
<!-- _header.html -->

<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>{{from-tag}}</title>
</head>
<body>
```

```html
<!-- _footer.html -->

</body>
</html>
 ```

```markdown
<!-- some/path/foo.md -->

Mar 24 2024

:ssg-title Real Header

# Some Header 2

Some para
```

This is the generated HTML equivalent, in `${dst}/some/path/foo.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Real Header</title>
</head>
<body>
<p>Mar 24 2024</p>
<h1>Some Header 2</p>
<p>Some para</p>
</body>
</html>
```

Note how `{{from-tag}}` in `_header.html` will cause ssg-go to use `Real Header`
as the document head title.

On the other hand, the `{{from-h1}}` will cause ssg-go to use `Some Header 2`
as the document head title.

### Cascading header and footer templates

ssg-go cascades `_header.html` and `_footer.html` down the directory tree

If your tree looks like this:

```
├── _header.html
├── blog
│   ├── 2023
│   │   ├── _header.html
│   │   ├── bar.md
│   │   ├── baz
│   │   │   └── index.md
│   │   └── foo.md
│   ├── _header.html
│   └── index.md
└── index.md  
```

Then:

- `/index.md` will use `/_header.html`

- `/blog/index.md` will use `/blog/_header.html`

- `/blog/2023/baz/index.md` will use `/blog/2023/_header.html`

## Extending and consuming ssg-go

### ssg-go walk

Given `src` and `dst` paths, ssg-go walks `src` and performs the
following operations for each source file:

- If path is ignored

  ssg-go continues to the next input.

- If path is unignored directory

  ssg-go collects templates from `_header.html` and `_footer.html`

- If path is an unignored file

  ssg-go reads the data and send it to all of the `Pipeline`s.

  The output from the last `Pipeline` is used as input to ssg-go *core*,
  which handles the Markdown conversion and assembly of HTML outputs.

  ```
  raw_data -> [pipelines] -> [core] -> output
  ```

  Pipelines can use these 2 well known errors control ssg-go walk control flow:

  - `ErrBreakPipelines`

    ssg-go stops going through pipelines and immedietly advances to core

  - `ErrSkipCore`

    Like with `ErrBreakPipelines`, but ssg-go also skips core.

### ssg-go core

For an input file, ssg-go performs these actions:

- If path is a file

  ssg-go calls `Hook` on the file to modify the data.
  We can use minifiers here.

- If path has non-`.md` extension

  ssg-go will not convert it to HTML,
  and it will simply mirror the file to `$dst`:

  ```
  core_input -> [hook] -> output
  ```

- If path has `.md` extension

  ssg-go assembles and adds the HTML output to the outputs.
  After the assembly, `HookGenerate` is called on the data.

  ```
  core_input -> [hook] -> [covert-to-html] -> [hookGenerate] -> output
  ```

### Options

Go programmers can extend ssg-go via the [`Option` type](./options.go).

[soyweb](../soyweb/) also extends ssg via `Option`,
and provides extra functionality such as index generator and minifiers.

Extending via options can be done in 2 rough categories:

- Hooks

  Hooks modify file content in-place in memory (without renaming input files).
  In short, it maps the input bytes 1-1 to new input bytes.

  A good usecase for hooks would be minifiers or content filter.

- Pipelines

  Pipelines have full control of the walk, and can arbitarily generate new
  ssg-go outputs and change input properties such as filenames and modes.

  In soyweb, pipelines are used to implement the [index generator](../soyweb/index.go).

#### `Hook` option

`Hook` is a Go function used to modify data after it is read,
preserving the filename. `Hook` is only called on the raw inputs
but not on the generated HTMLs.

It is enabled with `WithHook(hook)`

#### `HookGenerate` option

`HookGenerate` is a Go function called on every generated HTML.
For example, soyweb uses this option to implement output minifier.

It is enabled with `WithHookGenerate(hook)`

#### `Pipeline` option

`Pipeline` is a Go function called on a file during directory walk.
To reduce complexity, ignored files and ssg headers/footers are not sent
to `Pipeline`. This preserves the core functionality of the original ssg.

Pipelines can be chained together with `WithPipelines(p1, p2, p3)`

`WithPipelines` accepts 2 types of spread parameters:

- `Pipeline`

  An alias to `func(path string, data []byte, d fs.DirEntry) (string, []byte, fs.DirEntry, error)`
  We can call these pipelines *pure* pipelines.

  An example would be this `Pipeline`, that reads atime from filesystems
  and prepends the date string to Markdowns:

  ```go
  func pipelineUpdatedAt(path string, data []byte, d fs.DirEntry) (string, []byte, fs.DirEntry, error) {
  	if d.IsDir() {
  		return path, data, d, nil
  	}
  	if filepath.Ext(path) != ".md" {
  		return path, data, d, nil
  	}
  	info, err := d.Info()
  	if err != nil {
  		return "", nil, nil, err
  	}
  	buf := bytes.NewBufferString(
  		fmt.Sprintf("%s", info.ModTime().Format(time.DateOnly)),
  	)

  	buf.ReadFrom(bytes.NewBuffer(data))
  	return path, buf.Bytes(), d, nil
  }
  ```

- `func(s *Ssg) Pipeline`

  Some pipelines are not pure and might need something from `Ssg`, so
  we allow these pipeline constructors as arguments to `WithPipelines`.

  An example for this type of pipelines would be the [index generator](../soyweb/index.go),
  which needs to know which files are ignored in addition to `$src` and `$dst`.

### Streaming and caching builds

To minimize runtime memory usage, ssg-go builds and writes concurrently.
There're 2 main ssg threads: one is for building the outputs,
and the other is the write thread.

The build thread *sequentially* reads, builds and sends outputs
to the write thread via a buffered Go channel.

Bufffering allows the builder thread to continue to build and send outputs
to the writer until the buffer is full.

This helps reduce back pressure, and keeps memory usage low.
The buffer size is, by default, 2x of the number of writers.

This means that, at any point in time during a generation of any number of files
with 20 writers, ssg-go will at most only hold 40 output files
in memory (in the buffered channel).

If you are importing ssg-go to your code and you don't want this
streaming behavior, you can use the exposed function `Build`, `WriteOutSlice`,
and `GenerateMetadata`:

```go
files, dist, err := ssg.Build(src, dst, title, url)
if err != nil {
  panic(err)
}

err = ssg.WriteOutSlice(dist)
if err != nil {
  panic(err)
}

err = GenerateMetadata(src, dst, urk, files, dist, time.Now())
if err != nil {
  panic(err)
}
```
