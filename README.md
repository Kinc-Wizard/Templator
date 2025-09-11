# Templator

![GitHub Stars](https://img.shields.io/github/stars/Kinc-Wizard/Templator?style=social)
![GitHub Forks](https://img.shields.io/github/forks/Kinc-Wizard/Templator?style=social)
![GitHub Watchers](https://img.shields.io/github/watchers/Kinc-Wizard/Templator?style=social)

ShellCode Template, generate custom shellcode loader trought template selection

<p align="center">
  <img src="screenshots/logo.png" alt="Templator Logo" width="500"/>
</p>

This program is initially designed to store my/your custom loaders and make them easier to modify so you can quickly inject the shellcode of your preferred C2.

In a second phase, the goal is to allow you to generate a custom loader from scratch by simply referencing placeholders, using the selected programming language and API level.

For now, only Windows is targeted. As a result, there is no native distinction between Linux and Windows when referencing templates.

# Interface
<img width="1283" height="894" alt="image" src="https://github.com/user-attachments/assets/844f1059-6e1d-41b0-970b-d76406d3d45c" />
<img width="1262" height="509" alt="image" src="https://github.com/user-attachments/assets/a6c9f9b3-7c07-4bf8-9a0e-3d3c2bf92e73" />
<img width="607" height="295" alt="image" src="https://github.com/user-attachments/assets/15fb901e-a68f-4d58-a8ac-9ebbc14f935f" />

# Install
Project work with golang 1.21
```
git clone https://github.com/Kinc-Wizard/Templator.git
cd Templator
go mod tidy
chmod +x ./install_dependencies.sh
```

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

## Features v0.6
- [x] Add compilation command line inside template.json : 
- [x] Update read me : download and installation , goal and purpose of templator
- [ ] Add New template ( cpp - native - spawn process injection)
- [x] Use logo to make loading page during compilation
- [x] Decrase verbosity web terminal
- [x] Unitary test in new folder add to GitIgnore, unitary need to test to compile all template in all archtecture available.
- [ ] Post demo video in the readme
- [ ] Increase choice methode for encryption template
- [x] update Read me with, view, fork and clone incrémenation notion
- [x] GitIgnore on output
- [ ] Column custom commentaty by operator
- [ ] Column hash for IOCs 

 
## Features v0.7
### Generator function :
- [ ] Add référence library :
    ==> Native.h
    ==> Win32.h
- [ ] New web page dedicated to template génération
      - [ ] Select langage
      - [ ] Select api level
      - [ ] Visualize in interconnected map the possible systemcall : pre-phrase --> possibility step by step (1 OpenProcress, Exec, Download) --> find a way to map the possibility
      - [ ] Can save the actual template
      - [ ] Can debbug (try too compile) actual template
      - [ ] Expérimenten full place holder template génération {{win32.openprocress, native.ntopenprocess}}
      - [ ] Add New place holder {{langage.c}} in function of language