---
sidebar_position: 5
---

# Structs and OOP

## Defining a Struct

```GS
{
   orang {
     nama
     umur
   }
}
```

## Creating an Instance

An instance reuses the **same name** as the struct type — the struct definition and the instance literal share the identifier:
```GS
{
    orang {
      nama
      umur
   }
   
   fnc main {
        orang {
          nama = "Budi"
          umur = 20
       }
       budi = orang
       print(budi.nama)
   }
}
```

## Methods

Methods use a Go-style receiver syntax: `fnc (receiver type) name { ... }`:
```GS
{
   orang {
     nama
     umur
   }

   fnc (o orang) sapa {
        print("halo, aku {o.nama}")
     }

   fnc main {
        orang {
          nama = "Budi"
          umur = 20
       }
       budi = orang
       budi.sapa()
   }
}
```

## Next Steps

Continue to [Lists and Maps](./lists-and-maps).
