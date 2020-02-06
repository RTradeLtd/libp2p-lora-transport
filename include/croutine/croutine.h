#include <pthread.h>
#include <stdlib.h>
#include <stdio.h>
#include <stdbool.h>

// STRUCTS

typedef struct go_thread_data {
    int tid;
    char *data;
} croutine_data;

typedef struct croutine_future{
    char *returnValue;
    int exitCode, id;
    bool ready, fired;
    pthread_mutex_t mutex;
}  croutine_future;

// FUNCTIONS

struct croutine_future *new_croutine_future();

void fire_croutine_fut(
    pthread_t *thread, 
    pthread_attr_t *attr,
    void *arg, 
    void *(execute_func)(void *),
    struct croutine_future *fut
);
void lock_croutine_fut(struct croutine_future *fut);
void unlock_croutine_fut(struct croutine_future *fut);
bool has_fired_croutine_fut(struct croutine_future *fut);
int get_exit_code_croutine_fut(struct croutine_future *fut);
// returns a string representation of the future
char *string_croutine_fut(struct croutine_future *fut);
