Exercise 1 - Theory questions
-----------------------------

### Concepts

What is the difference between *concurrency* and *parallelism*?
> *Your answer here*

Concurrency: Multiplke things, for exmaple threads, run at the same time on the same CPU. Can be done by mutlthreading.
Parallelism: Run a program/task/operation on mutliple CPUs/systems 


What is the difference between a *race condition* and a *data race*? 
> *Your answer here* 
Race condition: Happens when threads use a shared recource. The outcome depends on the timing of the threads and the order in which it does things.

Data race: Very similar to race condition, two or more threads accses a shared recourse and alter it in some way(write operation). Data race then occurs when there is no proper sync between the threads to prevent errors. 


 
*Very* roughly - what does a *scheduler* do, and how does it do it?
> *Your answer here* 
A scheduler 


### Engineering

Why would we use multiple threads? What kinds of problems do threads solve?
> *Your answer here*
Threads are good since they can solve problems faster and is can run different taskes at the same time. This is usefuel since we can devide problems intro different threads and run all of them at the same time and stuff.

Some languages support "fibers" (sometimes called "green threads") or "coroutines"? What are they, and why would we rather use them over threads?
> *Your answer here*

Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
> *Your answer here*

ytelsen kan bli bedre. istedefor å kjøre ting i serie, kan det gjøres mer parallelt. Problemet er at du må behandle forskjellige threads, og passe på ting som for eksemepel race conditions.

What do you think is best - *shared variables* or *message passing*?
> *Your answer here*

shared variables er lettere å implementere, men man kan få synkronisasjons problemer. unngår en del synkronisasjons problemer ved at messages behandles og handles på om det er mulig. kan være vanskeligere å implementere?

