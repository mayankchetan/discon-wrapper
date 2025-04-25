#include <stdio.h>
#include <stdlib.h>

#ifdef _WIN32
#include <windows.h>
#else
#include <dlfcn.h>
#endif

typedef void (*discon_func)(float*, int*, char*, char*, char*);

typedef struct {
    void* library_handle;
    void* function_handle;
} LibraryContext;

void discon_with_context(LibraryContext* context, float* avrSWAP, int* aviFAIL, char* accINFILE, char* avcOUTNAME, char* avcMSG) {
    discon_func discon = (discon_func)context->function_handle;
    discon(avrSWAP, aviFAIL, accINFILE, avcOUTNAME, avcMSG);
}

LibraryContext* create_library_context() {
    LibraryContext* context = (LibraryContext*)malloc(sizeof(LibraryContext));
    if (context) {
        context->library_handle = NULL;
        context->function_handle = NULL;
    }
    return context;
}

int load_shared_library_with_context(LibraryContext* context, const char* library_path, const char* function_name) {
    if (!context) {
        return 3; // Invalid context
    }

#ifdef _WIN32
    context->library_handle = LoadLibrary(library_path);
    if (!context->library_handle) {
        fprintf(stderr, "Failed to load library: %s\n", library_path);
        return 1;
    }

    context->function_handle = GetProcAddress((HMODULE)context->library_handle, function_name);
    if (!context->function_handle) {
        fprintf(stderr, "Failed to get function: %s\n", function_name);
        FreeLibrary((HMODULE)context->library_handle);
        context->library_handle = NULL;
        return 2;
    }
#else
    context->library_handle = dlopen(library_path, RTLD_LAZY);
    if (!context->library_handle) {
        fprintf(stderr, "Failed to load library: %s\nError: %s\n", library_path, dlerror());
        return 1;
    }

    context->function_handle = dlsym(context->library_handle, function_name);
    char* error = dlerror();
    if (error != NULL) {
        fprintf(stderr, "Failed to get function: %s\nError: %s\n", function_name, error);
        dlclose(context->library_handle);
        context->library_handle = NULL;
        return 2;
    }
#endif

    return 0;
}

void unload_shared_library_with_context(LibraryContext* context) {
    if (!context) {
        return;
    }

#ifdef _WIN32
    if (context->library_handle) {
        FreeLibrary((HMODULE)context->library_handle);
        context->library_handle = NULL;
    }
#else
    if (context->library_handle) {
        dlclose(context->library_handle);
        context->library_handle = NULL;
    }
#endif
    context->function_handle = NULL;
}

void free_library_context(LibraryContext* context) {
    if (context) {
        unload_shared_library_with_context(context);
        free(context);
    }
}

// Keep the old functions for backward compatibility but implement them using the new context-based functions
static LibraryContext* global_context = NULL;

void discon(float* avrSWAP, int* aviFAIL, char* accINFILE, char* avcOUTNAME, char* avcMSG) {
    if (global_context) {
        discon_with_context(global_context, avrSWAP, aviFAIL, accINFILE, avcOUTNAME, avcMSG);
    }
}

int load_shared_library(const char* library_path, const char* function_name) {
    if (!global_context) {
        global_context = create_library_context();
    }
    return load_shared_library_with_context(global_context, library_path, function_name);
}

void unload_shared_library() {
    if (global_context) {
        unload_shared_library_with_context(global_context);
    }
}