---
sidebar_position: 4
---

# Functions

## Declaring Functions

Functions are declared with `fnc`:
```GS
{
   fnc tambah(a, b) {
        return a + b
     }
   fnc main {
       hasil = tambah(3, 4)
       print(hasil)
   }
}
```

## Multiple Return Values

```GS
{
     fnc bagi(a, b) {
         return a / b, a % b
      }

     fnc main {
         hasil, sisa = bagi(10, 3)
         print(hasil)
         print(sisa)
    }
}
```

## Functions as Values

Functions are first-class values — you can store them in variables and pass them as arguments:
```GS
{
    fnc kali2(x) {
        return x * 2
     }

    fnc main {
        f = kali2
        print(f(5))
     }
}
```

## Next Steps

Continue to [Structs and OOP](./structs-and-oop).
