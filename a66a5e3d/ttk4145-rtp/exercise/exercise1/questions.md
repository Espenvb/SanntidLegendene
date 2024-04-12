Exercise 1 - Theory questions
-----------------------------

### Concepts

What is the difference between *concurrency* and *parallelism*?
> Concurrency is swipping between two threads all the time, while parallelis, is two threads run on two different CPU's at the same time.

What is the difference between a *race condition* and a *data race*? 
> For race condition, the threads are dependent on the timing/order of events. A data race occurs what two different threads access the same 
data race: 2 threads accessing data in arbitrary order
race condition: ordering of events arbitrariy leading to erronous behaviour
(source stackoverflow)
 
*Very* roughly - what does a *scheduler* do, and how does it do it?
> A scheduler decides which thread comes next 


### Engineering

Why would we use multiple threads? What kinds of problems do threads solve?
> *Your answer here*

Some languages support "fibers" (sometimes called "green threads") or "coroutines"? What are they, and why would we rather use them over threads?
> *Your answer here*

Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
> *Your answer here*

What do you think is best - *shared variables* or *message passing*?
> *Your answer here*


