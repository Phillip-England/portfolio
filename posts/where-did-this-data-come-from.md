---
title: "Where Did This Data Come From?"
subtitle: "Making use of stack traces in TypeScript with Bun."
date: 2026-08-15
description: ""
tags: []
draft: false
---

## Where Did This Object Come From?

When building systems, it is important to know the origin of the data you are dealing with. In Bun, we have direct access to the stack trace. For example:

```ts
let t: { stack?: string } = {};
Error.captureStackTrace(t, undefined);
console.log(t);
```

This will produce something like:

```bash
{
  originalLine: 2,
  originalColumn: 24,
  stack: "Error\n    at /Users/phillipengland/Projects/edu/index.ts:3:7",
}
```

But, take note, we are using `Error.captureStackTrace(t, undefined);`. What happens if we do not use `undefined`?

...I had to go do a bit of research to figure this one out, let me walk you through what I found.

## What is Currently on the Stack?

At first, I ended up with something like this:

```ts
function exampleOne() {
    function foo() {
        console.log('ahhh!');
    }

    foo();

    let t: { stack?: string } = {};
    Error.captureStackTrace(t);
    console.log(t.stack);
}

exampleOne();
```

Which produces the following output on my system:

```bash
ahhh!
Error
    at exampleOne (/Users/phillipengland/Projects/edu/index.ts:17:11)
    at /Users/phillipengland/Projects/edu/index.ts:21
```

Here is sort of how I verbalize this to myself:

```bash
# foo prints this
ahhh!

# we make a call to the static method 'captureStackTrace' on type 'Error'
Error
    # 'captureStackTrace' was called at this location while 'exampleOne' was running
    at exampleOne (/Users/phillipengland/Projects/edu/index.ts:17:11)

    # this is the location where 'exampleOne' was called
    at /Users/phillipengland/Projects/edu/index.ts:21
```

One thing I observe is that we do gain access to the location of `Error.captureStackTrace(t)`, but the stack frame is associated with the function currently executing, which in this case is `exampleOne`. We also get the rest of the stack leading up to that call. This might not be ideal, let's take a look at another example.

## Capturing the Exact Location of Error.captureStackTrace

Consider the following example:

```ts
function exampleTwo(t: { stack?: string }) {
    console.log(t.stack);
}

let t: { stack?: string } = {};
Error.captureStackTrace(t);
exampleTwo(t);
```

It produces the following output on my system:

```bash
Error
    at /Users/phillipengland/Projects/edu/index.ts:18:7
```

In this case, I observe we gain access to the direct location where `Error.captureStackTrace(t)` was called.

Likewise, if we continue to pipe the captured stack downward, it preserves its location:

```ts
function foo(t: { stack?: string }) {
    console.log(t.stack);
}

function exampleThree(t: { stack?: string }) {
    foo(t);
}

let t: { stack?: string } = {};
Error.captureStackTrace(t);
exampleThree(t);
```

## Creating a Function to Capture Location

What if we wanted a function which allowed us to capture the current location? That would be a useful primitive in many different applications.

Here is what I came up with at first:

```ts
function exampleFour(): { stack?: string } {
    let t: { stack?: string } = {};
    Error.captureStackTrace(t);
    return t;
}

let loc = exampleFour();
console.log(loc.stack);
```

Which produces the following output on my system:

```bash
Error
    at exampleFour (/Users/phillipengland/Projects/edu/index.ts:13:11)
    at /Users/phillipengland/Projects/edu/index.ts:17:11
```

But, notice how this is cluttered and contains more than just the location I intend to grab? Well, `Error.captureStackTrace` provides a solution.

The second argument allows us to tell `captureStackTrace` which function should be omitted from the resulting stack trace, along with the frames above it. We can modify the function like so:

```ts
function exampleFive(): { stack?: string } {
    let t: { stack?: string } = {};
    Error.captureStackTrace(t, exampleFive);
    return t;
}

let loc = exampleFive();
console.log(loc.stack);
```

Which produces:

```bash
Error
    at /Users/phillipengland/Projects/edu/index.ts:16:11
```

Now the `exampleFive` frame itself is gone, leaving us with the location where `exampleFive` was called.

This provides a very clear path to creating a class dedicated to capturing the precise location of an object's initialization.

## Putting it All Together

Using these concepts, I eventually ended up with something like this:

```ts
class Loc {
  constructor(
    public readonly file: string,
    public readonly line: number,
    public readonly column: number,
  ) {}

  static get(): Loc | null {
    const target: { stack?: string } = {};

    Error.captureStackTrace(target, Loc.get);

    if (!target.stack) return null;

    return Loc.parse(target.stack);
  }

  private static parse(stack: string): Loc | null {
    for (const line of stack.split("\n")) {
      const match = line
        .trim()
        .match(/^at (?:.+ \()?(.+):(\d+):(\d+)\)?$/);

      if (!match) continue;

      return new Loc(
        match[1] as string,
        Number(match[2]),
        Number(match[3]),
      );
    }

    return null;
  }
}

console.log(Loc.get());
```

Which produced the following on my system:

```bash
Loc {
  file: "/Users/phillipengland/Projects/edu/index.ts",
  line: 46,
  column: 17,
}
```

Which is exactly what we want. A simple primitive in Bun which can be composed into other objects to determine their place of origin within a given program, like so:

```ts
class Foo {
    loc: Loc | null;

    constructor(loc: Loc | null) {
        this.loc = loc;
    }
}

let foo = new Foo(Loc.get());
```

Now `foo.loc` points us back to the location where `Loc.get()` was called, giving the object a simple way to carry information about where it originated.