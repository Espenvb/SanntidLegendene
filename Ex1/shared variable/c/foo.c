// Compile with `gcc foo.c -Wall -std=gnu99 -lpthread`, or use the makefile
// The executable will be named `foo` if you use the makefile, or `a.out` if you use gcc directly

#include <pthread.h>
#include <stdio.h>

int i = 0;

pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;


// Note the return type: void*
void *incrementingThreadFunction(void *arg ){
    // TODO: increment i 1_000_000 times
    for(int a = 0; a < 100000; a++){
        pthread_mutex_unlock(&mutex);
        i++;
        pthread_mutex_lock(&mutex);
    }
    
    return NULL;
}

void* decrementingThreadFunction(void *arg){
    // TODO: decrement i 1_000_000 times
    for(int a = 0; a < 1; a++){
        pthread_mutex_unlock(&mutex);
        i--;
        pthread_mutex_lock(&mutex);
    }
    return NULL;
}


int main(){
    // TODO: 
    // start the two functions as their own threads using `pthread_create`
    // Hint: search the web! Maybe try "pthread_create example"?
    pthread_t threadInc;
    pthread_t threadDec;
    
    if(pthread_create(&threadInc, NULL, incrementingThreadFunction, NULL) != 0){
        perror("pthread_create() error");
        return 1;
    }

    if(pthread_create(&threadDec, NULL, decrementingThreadFunction, NULL) != 0){
        perror("pthread_create() error");
        return 1;
    }

    if (pthread_join(threadDec, NULL) != 0) {
        perror("pthread_create() error");
        return 1;
    }

    if (pthread_join(threadInc, NULL) != 0) {
        perror("pthread_create() error");
        return 1;
    }

    // TODO:
    // wait for the two threads to be done before printing the final result
    // Hint: Use `pthread_join`    
    
    printf("The magic number is: %d\n", i);
    return 0;
}
