# Templator
ShellCode Template, generate custom shellcode loader trought template selection

<p align="center">
  <img src="screenshots/logo.png" alt="Templator Logo" width="500"/>
</p>

# Interface
<img width="1283" height="894" alt="image" src="https://github.com/user-attachments/assets/844f1059-6e1d-41b0-970b-d76406d3d45c" />
<img width="1262" height="509" alt="image" src="https://github.com/user-attachments/assets/a6c9f9b3-7c07-4bf8-9a0e-3d3c2bf92e73" />
<img width="607" height="295" alt="image" src="https://github.com/user-attachments/assets/15fb901e-a68f-4d58-a8ac-9ebbc14f935f" />


# Template management

## Template Variable 
- position to insert shellcode : `{{shell_code}}`

- position to insert process name : `{{process_name}}`

- position to insert process path : `{{process_path}}`

- position to insert encryption key : `{{key}}`

- position to insert encryption iv: `{{iv}}`

- position to insert encrypted shellcode : `{{encrypted_shellcode}}`

## Template referencement 
You need to give a template that correspond exactly to the of the template.json name
![image](screenshots/template-name.png)
![image](screenshots/ref-template-name.png)

## metadata of template
        "supports_obfuscation": true / false
        "support_encryption":true / false
        "needs_process": true / false
        "needs_process_path": true / false
