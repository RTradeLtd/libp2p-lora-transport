#include <pthread.h>
#include <stdlib.h>
#include <stdio.h>
#include <stdbool.h>
#include "../../include/croutine/croutine.h"
#include <time.h>
#ifdef _WIN32 // conditionally import sleep
#include <Windows.h>
#else
#include <unistd.h>
#endif

#define MAX_THREADS 4

void lock_croutine_fut(struct croutine_future *fut) {
    pthread_mutex_lock(&fut->mutex);
}
void unlock_croutine_fut(struct croutine_future *fut) {
    pthread_mutex_unlock(&fut->mutex);
}

bool has_fired_croutine_fut(struct croutine_future *fut) {
    lock_croutine_fut(fut);
    bool fired = (*fut).fired;
    unlock_croutine_fut(fut);
    return fired;
}

int get_exit_code_croutine_fut(struct croutine_future *fut) {
    lock_croutine_fut(fut);
    int exitCode = (*fut).exitCode;
    unlock_croutine_fut(fut);
    return exitCode;
}

void print_croutine_fut(struct croutine_future *fut) {
    lock_croutine_fut(fut);
    printf(
        "id: %d\treturnValueAddr: %s\texitCode: %d\tready: %d\tfired: %d\n",
        (*fut).id, (*fut).returnValue, (*fut).exitCode, (*fut).ready, (*fut).fired
    );
    unlock_croutine_fut(fut);
}
/* 
fire_croutine_fut: is used to execute an arbitrary function, storing the result in the "future"
This is my "hacky" attempt at a hybrid goroutine+channel
*/
void fire_croutine_fut(
    pthread_t *thread, 
    pthread_attr_t *attr,
    void *arg, 
    void *(execute_func)(void *),
    struct croutine_future *fut
) {
    int response = pthread_create(thread, attr, execute_func, arg);
    lock_croutine_fut(fut);
    fut->exitCode = response;
    fut->returnValue = (char*)&response;
    fut->fired = true;
    unlock_croutine_fut(fut);
}

struct croutine_future *new_croutine_future() {
    char *returnValue = malloc(sizeof(char));
    int exitCode;
    pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;
    exitCode = 1;
    struct croutine_future *fut = malloc(sizeof(croutine_future));
    (*fut).returnValue = returnValue;
    (*fut).exitCode = exitCode;
    (*fut).mutex = mutex;
    (*fut).ready = true;
    (*fut).returnValue = "hello world";
    return fut;
}

