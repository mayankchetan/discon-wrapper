#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef void (*DISCON_FUNC)(float *avrSWAP, int *aviFAIL, char *accINFILE, char *avcOUTNAME, char *avcMSG);

#ifdef _WIN32
#include <windows.h>
#else
#include <dlfcn.h>
#endif

const char *libPath = "discon-client.dll"; // Path to the DLL on Windows

int main()
{

#ifdef _WIN32
    HMODULE handle = LoadLibrary(libPath);
    if (!handle)
    {
        fprintf(stderr, "Error loading library: %lu\n", GetLastError());
        return EXIT_FAILURE;
    }

    // Load the DISCON function
    DISCON_FUNC DISCON = (DISCON_FUNC)GetProcAddress(handle, "DISCON");
    if (!DISCON)
    {
        fprintf(stderr, "Error loading DISCON function: %lu\n", GetLastError());
        FreeLibrary(handle);
        return EXIT_FAILURE;
    }
#else
    void *handle = dlopen(libPath, RTLD_LAZY);
    if (!handle)
    {
        fprintf(stderr, "Error loading library: %s\n", dlerror());
        return EXIT_FAILURE;
    }

    // Clear any existing errors
    dlerror();

    // Load the DISCON function
    DISCON_FUNC DISCON = (DISCON_FUNC)dlsym(handle, "DISCON");
    char *error = dlerror();
    if (error != NULL)
    {
        fprintf(stderr, "Error loading DISCON function: %s\n", error);
        dlclose(handle);
        return EXIT_FAILURE;
    }
#endif

    // Prepare the arguments for the DISCON function
    const int swap_array_size = 130;                 // Length of the swap array
    const int char_array_size = 32;                  // Length of the character arrays
    float avrSWAP[swap_array_size] = {0};            // Initialize the swap array
    int aviFAIL = 1;                                 // Initialize aviFAIL
    char accINFILE[] = "input.txt";                  // Example input file name
    char avcOUTNAME[char_array_size] = "output.txt"; // Example output file name
    char avcMSG[char_array_size] = "Hello, World!";  // Example message

    // Set total size of swap array
    avrSWAP[128] = (float)swap_array_size; // Set the size of swap array

    // Set the size of accINFILE string including terminator
    avrSWAP[49] = (float)strlen(accINFILE) + 1;

    // Set the size of avcOUTNAME string including terminator
    avrSWAP[50] = (float)strlen(avcOUTNAME) + 1;

    // Set the maximum size of avcOUTNAME string including terminator
    avrSWAP[63] = (float)char_array_size;

    // Set the maximum size of avcMSG string including terminator
    avrSWAP[48] = (float)char_array_size;

    // Call the DISCON function in a loop
    for (int i = 1; i < 10; i++)
    {
        printf("test-app: calling DISCON, iteration %d\n", i);
        DISCON(avrSWAP, &aviFAIL, accINFILE, avcOUTNAME, avcMSG);
        for (int j = 0; j < swap_array_size; j++)
        {
            if (avrSWAP[j] != 0.0)
            {
                printf("test-app: avrSWAP[%d]: %f\n", j, avrSWAP[j]);
            }
        }
        printf("test-app: aviFAIL = %d\n", aviFAIL);
        printf("test-app: accINFILE: %s\n", accINFILE);
        printf("test-app: avcOUTNAME: %s\n", avcOUTNAME);
        printf("test-app: avcMSG: %s\n", avcMSG);
    }

#ifdef _WIN32
    // Close the library
    FreeLibrary(handle);
#else
    // Close the library
    dlclose(handle);
#endif

    return EXIT_SUCCESS;
}