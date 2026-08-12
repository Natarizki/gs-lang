---
sidebar_position: 7
---

# Error Handling

## Raising an Error

Use `error "message"` to raise an error. It automatically propagates up the call stack until caught, or halts the program with a message if it reaches `main`:

```GS
{
   fnc bagi(a, b) {
       if b == 0 {
          error "tidak bisa bagi nol"
        }
        return a / b
     }
  fnc main {
       hasil = bagi(10, 0)
       print(hasil)
    }
}
```
Running this halts the program with an error message, since the error is never caught.

## Catching an Error with `try`

Prefix a call with `try` to catch an error instead of letting the program crash. The result becomes `null` (`_`) if an error occurred:

```GS
{
   fnc bagi(a, b) {
          if b == 0 {
             error "tidak bisa bagi nol"
        }
        return a / b
  }fnc main {
       try hasil = bagi(10, 0)
       if hasil == _ {
            print("gagal bagi, error ketangkep!")
        }

        normal = bagi(10, 2)
        print(normal)
     }
}
```

## Next Steps

Continue to [Tasks](./tasks).
