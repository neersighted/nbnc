# nbnc

a simple null (transparent) bnc

nbnc is a simple, transparent IRC proxy. nbnc implements none of the features of other bouncers,
instead focusing on simple proxy connections. nbnc does not allow you to reattach, nor does it
implement persistance. nbnc creates a new connection for every connection it receives.

nbnc only takes two arguments, which are self-explanatory:

```
nbnc <config> <log>
```

Configuration is simple, and is done in the toml format:

```
debug = false

listen = ":6667"

cert = "fullchain.pem"
key = "privkey.pem"

[bouncer.espernet]
target = "irc.esper.net:6667"

[bouncer.freenode]
target = "irc.freenode.net:6697"
secure = true

[bouncer.privatenet]
target = "freenetwork.io:6697"
secure = true
noverify = true
password = opensesame

```


To enable SSL listening, both `cert` and `key` must be specified. Encrypted private keys are not
supported.

To connect over SSL, specify `secure`. To skip checking for certificate validity, specify `noverify`.

Log can be any file, or `-` for stdout.

To connect to nbnc, connect to the listen port (enabling/disabling SSL as needed) normally. Configure
a password in the form `name:password`, where `name` is `[bouncer.<name>]`, and password is
`password = <password>` in the config. Password can be omitted if not included in the bouncer config.
