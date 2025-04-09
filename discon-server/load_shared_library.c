#include <stdio.h>
#include <stdlib.h>

#ifdef _WIN32
#include <windows.h>
#else
#include <dlfcn.h>
#endif

typedef void (*discon_func)(float*, int*, char*, char*, char*);

static void* library_handle = NULL;
static void* function_handle = NULL;

void discon(float* avrSWAP, int* aviFAIL, char* accINFILE, char* avcOUTNAME, char* avcMSG) {
    discon_func discon = (discon_func)function_handle;
    discon(avrSWAP, aviFAIL, accINFILE, avcOUTNAME, avcMSG);
}


int load_shared_library(const char* library_path, const char* function_name) {

#ifdef _WIN32
    library_handle = LoadLibrary(library_path);
    if (!library_handle) {
        fprintf(stderr, "Failed to load library: %s\n", library_path);
        return 1;
    }

    function_handle = GetProcAddress((HMODULE)library_handle, function_name);
    if (!function_handle) {
        fprintf(stderr, "Failed to get function: %s\n", function_name);
        FreeLibrary((HMODULE)library_handle);
        return 2;
    }
#else
    library_handle = dlopen(library_path, RTLD_LAZY);
    if (!library_handle) {
        fprintf(stderr, "Failed to load library: %s\nError: %s\n", library_path, dlerror());
        return 1;
    }

    function_handle = dlsym(library_handle, function_name);
    char* error = dlerror();
    if (error != NULL) {
        fprintf(stderr, "Failed to get function: %s\nError: %s\n", function_name, error);
        dlclose(library_handle);
        return 2;
    }
#endif

    return 0;
}

void unload_shared_library() {
#ifdef _WIN32
    if (library_handle) {
        FreeLibrary((HMODULE)library_handle);
        library_handle = NULL;
    }
#else
    if (library_handle) {
        dlclose(library_handle);
        library_handle = NULL;
    }
#endif
    function_handle = NULL;
}