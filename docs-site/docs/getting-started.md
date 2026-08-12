---
sidebar_position: 2
---

# Getting Started

## Why the outer `{ }`?

GS uses `{ }` as its **only** block style — no indentation-based blocks. The outer `{ }` at the top of your entry file is a **task block**, GS's unit of program execution (more on this in [Tasks](./tasks)). If you forget it, GS gives you a parse error and the program won't run.

## Variables

Variables in GS are dynamically typed — no need to declare a type:
```GS
{
   fnc main {
      nama = "Budi"
      umur = 20
     aktif = true
   }
}
```
## String Interpolation

Embed expressions directly inside strings using `{}`:
```GS
{
  fnc main {
      nama = "Budi"
      umur = 20
      
      print("hello, {nama}! you are {umur} years old")
    
  }
}
```
## Comments

GS supports both `//` and `#` for comments:
```GS
// this is a comment
```

## this is also a comment

Multi-line comments use `/* ... */`.

## Compiling to a Binary

Instead of interpreting the file each time, you can compile it into a standalone binary:
```bash
./gsc build hello.gs -o hello
./hello
```
The resulting binary runs without needing `gsc` or the source file anymore.

## Next Steps

Continue to [Control Flow](./control-flow).
