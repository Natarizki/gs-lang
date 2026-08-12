# GoSong (GS)



![GoSong Logo](logo.png)



**GoSong (GS)** is a general-purpose programming language designed to be extremely easy to learn, blending the syntax style of Go and Python, with a unique **task orchestration** feature for declaratively controlling program execution flow.

## Key Features

- **`{}` syntax** — consistent and strict, inspired by Go/C
- **Dynamic typing** — no explicit type declarations needed
- **Task orchestration** — `start`, `then`, `startif`, `elstart` to control the order and conditions of execution between program blocks
- **Automatic error handling** — `error` and `try` with auto-propagation
- **Simple OOP** — struct + methods in Go style (`fnc (receiver type) name { ... }`)
- **String interpolation** — `"hello {name}"`
- **Built-in package manager** — `gsi`/`gsc install` to fetch packages from GitHub
- **Compile to binary** — `gsc build` produces a standalone binary

## Installation
```bash
git clone https://github.com/Natarizki/gs-lang.git
cd gs-lang
go build -o gsc .
```
## Example Program
```GS
{
   fnc main {
   print("hello world!")
  }
}
```
Run it with:
```GS
./gsc run hello.gs
```
Compile to binary:
```GS
./gsc build hello.gs -o hello
./hello
```
## Status

This project is under active development.

## License

GNU General Public License v3.0 — see [LICENSE](LICENSE) for details.
