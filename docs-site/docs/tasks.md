---
sidebar_position: 8
---

# Tasks

Task orchestration is GS's most distinctive feature. Every top-level `{ }` block in your entry file is a **task**, automatically assigned a sequential ID (`id::001`, `id::002`, `id::003`, ...).

## Single Task Programs

If your file has only one task, it runs automatically — no orchestration needed:

```GS
{
   fnc main {
       print("hello world!")
   }
}
```

## Multiple Tasks: `start` and `then`

When you have multiple tasks, use `start` to pick the entry point, and `then` to run tasks sequentially afterward:

```GS
{
   fnc main {
       print("task A")
   }
}

{
   fnc main {
      print("task B")
   }
}

{
   fnc main {
       print("task C")
   }
}

start = id::001
then = id::002 | id::003
```
Output:
```
task A
task B
task C
```

`start`/`then` are only allowed **outside** any task block, at the top level of the entry file.

## Conditional Task Control: `startif` and `elstart`

Inside a task, use `startif`/`elstart` to conditionally trigger another task:

```
{
   fnc main {
       x = 10
       startif x > 5 {
          start = id::003
       } elstart {
          start = id::002
     }
  }
}

{
   fnc main {
       print("condition was false")
    }
}

{
   fnc main {
       print("condition was true")
   }
}

start = id::001
```

`startif`/`elstart` are only allowed **inside** a task block — outside a task, use `start`/`then` instead.

## Tasks That Aren't Referenced

If a task isn't referenced by `start`/`then` and isn't invoked by another task's `startif`/`elstart`, it becomes dead code and never runs.

## Next Steps

Continue to [Modules and CLI](./modules-and-cli).
