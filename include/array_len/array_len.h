/* returns the number of elements in an array */
#define array_len(x) ( (size_t)sizeof(x) / sizeof(x[0]) )
/* returns the size of the array in bytes */
#define array_size(x) sizeof(x)