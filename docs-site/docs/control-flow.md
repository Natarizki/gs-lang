---
sidebar_position: 3
---

# Control Flow

## If / But If / Else

GS uses `if`, `but if` (instead of `else if`), and `else`:
```GS
{
    fnc main {
    umur = 20

    if umur >= 18 {
             print("dewasa")
        } but if umur >= 13 {
             print("remaja")
        } else {
             print("anak-anak")
        }
   }
}
```
## For Loops

GS has a single `for ... in` loop form, commonly paired with `range()`:
```GS
{
    fnc main {
       for i in range(5) {
           print(i)
        }
    }
}
```
You can also iterate directly over a list:
```GS
{
    fnc main {
        angka = [10, 20, 30]
        
        for n in angka {
            print(n)
       }
    }
}
```

## Break and Continue

```GS
{
     fnc main {
         for i in range(10) {
             if i == 5 {
               break
             }
             print(i)
         }
    }
}
```

## Match

`match` compares a value against a list of cases:

```GS
{
    fnc main {
        umur = 18

     match umur {
           18) print("pas 18 tahun")
           20) print("pas 20 tahun")
        }
   }
}
```
## Next Steps

Continue to [Functions](./functions).
