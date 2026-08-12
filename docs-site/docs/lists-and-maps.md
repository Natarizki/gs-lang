---
sidebar_position: 6
---

# Lists and Maps

## List Literals

```GS
{
   fnc main {
       angka = [1, 2, 3, 4, 5]
       print(angka[0])
       print(len(angka))
    }
}
```

## Slicing

```GS
{
   fnc main {
       angka = [1, 2, 3, 4, 5]
       potongan = angka[1:3]
       print(potongan)
    }
}
```

## Nested Lists

```GS
{
   fnc main {
       matrix = [[1, 2], [3, 4], [5, 6]]
       print(matrix[1][1])
    }
}
```

## Map Literals

```GS
{
   fnc main {
       data = {"nama": "Budi", "kota": "Jakarta"}
       print(data["nama"])
    }
}
```

## Nested Maps

```GS
{
   fnc main {
       data = {
         "user": {"nama": "Budi", "umur": 20},
         "aktif": true
      }
       print(data["user"]["nama"])
    }
}
```

## Next Steps

Continue to [Error Handling](./error-handling).
