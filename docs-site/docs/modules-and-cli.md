---
sidebar_position: 9
---

# Modules and CLI

## Importing Local Files

```GS
import "dir/file.gs"

{
   fnc main {
       hasil = tambah(3, 4)
       print(hasil)
   }
}
```

Files loaded via `import` are the **only exception** to GS's strict rule — they don't need to be wrapped in a task block.

## GS/af — Standard Library Functions

`GS/af` provides 30 built-in utility functions (string, list, map, file I/O, time, and misc operations). It must be imported before use:

```GS
import "GS/af"

{
   fnc main {
       besar = upper("halo")
       print(besar)
      
       angka = sort([5, 2, 8, 1])
       print(angka)
    }
}
```
## GS/fo — Math Functions

`GS/fo` provides advanced math operations (`sqrt`, `pow`, `sin`, `cos`, `abs`, `round`, `max`, `min`, and more):
```GS
import "GS/fo"

{
   fnc main {
       akar = sqrt(16)
       print(akar)
   }
}
```

## Remote Packages

Import a package hosted on GitHub with `GS/username/paket`:
```GS
import "GS/username/mathpkg"

{
   fnc main {
       hasil = Tambah(5, 3)
       print(hasil)
    }
}
```
Packages are stored globally at `~/.GS/username/paket/`. Only exported identifiers (starting with an uppercase letter) are accessible from outside the package.

## Installing Packages

```GS
gsc install username/paket
```

or the shorthand:

```GS
gsi username/paket
```

If a package isn't installed yet, `gsc run`/`gsc build` will install it automatically.

## CLI Reference

| Command | Description |
|---|---|
| `gsc run <file.gs>` | Run a GS source file directly |
| `gsc build <file.gs> -o <output>` | Compile to a standalone binary |
| `gsc install <username/paket>` | Install a package from GitHub |
| `gsi <username/paket>` | Shorthand for `gsc install` |

## That's it!

You've covered the core of GoSong. Explore the [examples](https://github.com/Natarizki/gs-lang/tree/main/examples) in the repository for more.
