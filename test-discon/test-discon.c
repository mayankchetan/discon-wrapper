#include <stdio.h>
#include <string.h>

static int num_calls = 0;

void discon(float *avrSWAP, int *aviFAIL, char *accINFILE, char *avcOUTNAME, char *avcMSG)
{

    ++num_calls; // Increment number of calls

    const int swap_length = (int)avrSWAP[128];
    const int infile_length = (int)avrSWAP[49];
    const int outname_length = (int)avrSWAP[50];
    const int msg_length = (int)avrSWAP[48];

    // Set the output parameters (for demonstration purposes)
    *aviFAIL = 0;                                                      // No failure
    snprintf(avcMSG, msg_length, "DISCON called %d times", num_calls); // Set message
    avcMSG[msg_length] = '\0';                                         // Null-terminate the string

    for (int i = 0; i < swap_length; i++)
    {
        if (avrSWAP[i] != 0.0)
        {
            printf("test-discon: avrSWAP[%d]: %f\n", i, avrSWAP[i]);
        }
    }
    printf("test-discon: aviFAIL: %d\n", *aviFAIL);
    printf("test-discon: accINFILE: %.*s\n", infile_length, accINFILE);
    printf("test-discon: avcOUTNAME: %.*s\n", outname_length, avcOUTNAME);
    printf("test-discon: avcMSG: %.*s\n", msg_length, avcMSG);
}