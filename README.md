# EX-Manager
Ex(ploit)-Manager helps you and your CTF team manage all your ex(ploits)s.

## What it does
Ex-Manager defines a *simple* and *correct* protocol that all your exploits must respect.
Doing this helps you and your team to be more productive,
not having to think about **iterate over targets** and **submit** the hundreds of flags you
are collecting.
Come on, go meet/write your latest ex(ploit).

## You cannot force me to use your protocol
You are right. FULL STOP.  
**Fork** this repo and **pull** all your requests **inside**.

## Protocol
Ex(ploit)-Manager doesn't force you to use a specific programming language,
it expects only an executable file, it can be anything, from *Bash*, *Python*,
*Go* or even *C*.

The executable will be called with a single target
and it may print to stdout the flags.
```
(EX-MANAGER)$ ./your-new-ex.py -target 10.0.0.1
CTF{Lbh sbhaq zr}
XXX{SXMgaXQgc29tZXRoaW5nIGFib3V0IGZ1cnJ5Pz8=}
CTF{ViBmdWJoeXEgZmdiYw==}
```

## Submition
At Ex(ploit)-Manager we take *submission* seriously, flags are collected andù
then a professional submitter, or one provided by you,
will send all the flags to the server.

If you have wrote a custom submitter and you want to share it, feel free
to open a pull request.
