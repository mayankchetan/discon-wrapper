#include <stdio.h>
#include <stdlib.h>

#ifdef _WIN32
#include <windows.h>
#else
#include <dlfcn.h>
#endif

#define NUM_HANDLES 8192

typedef void (*discon_func)(float *, int *, char *, char *, char *);

static void *library_handles[NUM_HANDLES] = {NULL};
static void *function_handles[NUM_HANDLES] = {NULL};

void discon(int connID, float *avrSWAP, int *aviFAIL, char *accINFILE, char *avcOUTNAME, char *avcMSG)
{
    discon_func discon = (discon_func)function_handles[connID];
    discon(avrSWAP, aviFAIL, accINFILE, avcOUTNAME, avcMSG);
}

int load_shared_library(int connID, const char *library_path, const char *function_name)
{

#ifdef _WIN32
    library_handles[connID] = LoadLibrary(library_paths[connID]);
    if (!library_handles[connID])
    {
        fprintf(stderr, "Failed to load library: %s\n", library_paths[connID]);
        fprintf(stderr, "Error: %lu\n", GetLastError());
        return 1;
    }

    function_handles[connID] = GetProcAddress((HMODULE)library_handles[connID], function_name);
    if (!function_handles[connID])
    {
        fprintf(stderr, "Failed to get function: %s\n", function_name);
        FreeLibrary((HMODULE)library_handles[connID]);
        return 2;
    }
#else
    library_handles[connID] = dlopen(library_path, RTLD_LAZY);
    if (!library_handles[connID])
    {
        fprintf(stderr, "Failed to load library: %s\nError: %s\n", library_path, dlerror());
        return 1;
    }

    function_handles[connID] = dlsym(library_handles[connID], function_name);
    char *error = dlerror();
    if (error != NULL)
    {
        fprintf(stderr, "Failed to get function: %s\nError: %s\n", function_name, error);
        dlclose(library_handles[connID]);
        return 2;
    }
#endif

    return 0;
}

void unload_shared_library(int connID)
{
#ifdef _WIN32
    if (library_handles[connID])
    {
        FreeLibrary((HMODULE)library_handles[connID]);
        library_handles[connID] = NULL;
    }
#else
    if (library_handles[connID])
    {
        dlclose(library_handles[connID]);
        library_handles[connID] = NULL;
    }
#endif
    function_handles[connID] = NULL;
}