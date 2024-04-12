// Compile with `gcc foo.c -Wall -std=gnu99 -lpthread`, or use the makefile
// The executable will be named `foo` if you use the makefile, or `a.out` if you use gcc directly

#include <pthread.h>
#include <stdio.h>

int i = 0;

// Note the return type: void*
void* incrementingThreadFunction(){
    for (int j = 0; j < 1000000; j++)
    {
        i++;
    }
    
    return NULL;
}

void* decrementingThreadFunction(){
    for (int k = 0; k < 1000000; k++)
    {
        i--;
    }
    
    return NULL;
}


int main(){
    pthread_t thread_increasing;
    pthread_t thread_decreasing;
    pthread_create(&thread_increasing,NULL,incrementingThreadFunction,NULL);
    pthread_create(&thread_decreasing,NULL,decrementingThreadFunction,NULL);  
    
    pthread_join(thread_increasing,NULL);
    pthread_join(thread_decreasing,NULL);    
    
    printf("The magic number is: %d\n", i);
    return 0;
}


//The result is not equal to zero because the threads change between each other. 