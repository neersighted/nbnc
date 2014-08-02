nbnc
====

a simple null (transparent) bnc

nbnc is simply an IRC-aware TCP proxy, or BNC. It uses IRC-style authentication
(`PASS`) for compatibility with existing IRC clients. It is intended for use
cases where you want to disguise your true host or apply a vanity hostname. I
wrote it for personal use with [IRCCloud], a hosted BNC service with a web
interface.

>NAME:
>   nbnc - simple null (transparent) bnc
>
>USAGE:
>   nbnc [global options] command [command options] [arguments...]
>
>VERSION:
>   0.0.1
>
>COMMANDS:
>   help, h      Shows a list of commands or help for one command
>
>GLOBAL OPTIONS:
>   -l, --laddr '0.0.0.0'        Local address to listen on.
>   -L, --lport '1337'           Local port to listen on.
>   -r, --raddr '127.0.0.1'      Remote address to connect to.
>   -R, --rport '6667'           Remote port to connect to.
>   -o, --oaddr                  Outgoing address to connect with.
>   -4                           Force connection to use IPv4.
>   -6                           Force connection to use IPv6.
>   -p, --pass 'opensesame'      Password to authenticate against.
>   --help, -h                   show help
>   --version, -v                print the version

[IRCCloud]: https://irccloud.com
