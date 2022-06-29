# Goonit

A Golang unit test module

## What Is Goonit

Goonit is a Golang module that helps you write short human-readable unit tests
for Go code.

## Why I Made This

I started writing my own code to support Go unit tests because I couldn't make
sense of [Ginkgo](https://github.com/onsi/ginkgo).  That was more due to my own
cognitive challenges than any defect in the project.  Lots of developers I
respect use it, as do many other successful Go projects.  I don't think there is
anything wrong with Ginkgo.

It just wasn't what I wanted.

I wanted to be able to look at a single failing unit test and know right away
what it tests, what input it needs, and what result it expects without ever
having seen that test before.  I could not look at a set of tests built with
Ginkgo and understand what a failing test actually meant or what I had to fix.

I wanted to keep my tests in separate functions with descriptive names.  When a
unit test function's name describes the test, the developer has a decent chance
of knowing what might be causing that test to fail without the need to review
the test code, even if they didn't write that test themselves.

I wanted to write reusable functions that set up the inputs for a single test at
a time.  A unit test that covers a single path through the tested code is much 
easier to read and understand, and setup function names can describe the inputs
the test needs to reproduce the tested conditions.  They can also describe how 
the test differs from its siblings.

I wanted the test's call to the function under test to be as obvious and visible
as possible.  That makes it easier to see what function the test was made to 
exercise, the parameters it expects, and the results it returns.

I wanted the expected results to be clear and easily associated with
the input the test created to produce them.

I wanted to write unit tests that look like little poems.  Ginkgo tests just 
didn't look like that to me.  To me, they looked like spreadsheets or tables,
which are tedious to read.

So I started writing the tests I wanted.  As I wrote those tests, I'd break out
the reusable parts common to multiple tests, and over time I'd reuse those bits
in different projects.  I reused some parts often enough that it occurred to me
to package them into a portable Go module.

And here we are.

## Why It's Named Goonit

The name starts with **Go** and ends with **Unit**.

It includes **Goo** which feels like fun.

It includes **Goon** which makes me think of Golden Age comic book henchmen.

Plus it's fun to say.

Goonit.

Say it with me.

Goonit.

OK, you seem convinced.

## Why I Prefer Short Readable Single-Use Unit Tests

Unit tests that are short and readable are more likely to be maintained over a
project's lifetime, and less likely to end up being deleted or disabled because
someone - maybe me - is under a tight deadline to fix a bug or ship a feature 
and fixing a failing unit tests seems too difficult or time-consuming.


