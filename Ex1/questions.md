Exercise 1 - Theory questions
-----------------------------

### Concepts

What is the difference between *concurrency* and *parallelism*?
> *Your answer here*


What is the difference between a *race condition* and a *data race*? 
> *Your answer here* 
 
*Very* roughly - what does a *scheduler* do, and how does it do it?
> *Your answer here* 


### Engineering

Why would we use multiple threads? What kinds of problems do threads solve?
> *Your answer here*

Some languages support "fibers" (sometimes called "green threads") or "coroutines"? What are they, and why would we rather use them over threads?
> *Your answer here*

Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
> ytelsen kan bli bedre. istedefor å kjøre ting i serie, kan det gjøres mer parallelt. Problemet er at du må behandle forskjellige threads, og passe på ting som for eksemepel race conditions. 

What do you think is best - *shared variables* or *message passing*?
> shared variables er lettere å implementere, men man kan få synkronisasjons problemer.
> unngår en del synkronisasjons problemer ved at messages behandles og handles på om det er mulig.
> kan være vanskeligere å implementere?


