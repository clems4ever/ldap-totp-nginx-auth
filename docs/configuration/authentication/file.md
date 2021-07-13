---
layout: default
title: File
parent: Authentication backends
grand_parent: Configuration
nav_order: 1
---

# File

**Authelia** supports a file as a users database.


## Configuration

Configuring Authelia to use a file is done by specifying the path to the
file in the configuration file.

```yaml
authentication_backend:
  disable_reset_password: false
  file:
    path: /config/users.yml
    password:
      algorithm: argon2id
      iterations: 1
      salt_length: 16
      parallelism: 8
      memory: 64
```


## Format

The format of the users file is as follows.

```yaml
users:
  john:
    displayname: "John Doe"
    password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
  harry:
    displayname: "Harry Potter"
    password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: harry.potter@authelia.com
    groups: []
  bob:
    displayname: "Bob Dylan"
    password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: bob.dylan@authelia.com
    groups:
      - dev
  james:
    displayname: "James Dean"
    password: "$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM"
    email: james.dean@authelia.com
```

This file should be set with read/write permissions as it could be updated by users
resetting their passwords.


## Options

### path
<div markdown="1">
type: string (path)
{: .label .label-config .label-purple } 
required: yes
{: .label .label-config .label-red }
</div>


### password

#### algorithm
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: argon2id
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Controls the hashing algorithm used for hashing new passwords. Value must be one of `argon2id` or `sha512.


#### iterations
<div markdown="1">
type: integer
{: .label .label-config .label-purple }
required: no
{: .label .label-config .label-green }
</div>

Controls the number of hashing iterations done by the other hashing settings.

When using `argon2id` the minimum is 1, which is also the recommended value.

When using `sha512` the minimum is 1000, and 50000 is the recommended value.


#### salt_length
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: 16
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Controls the length of the random salt added to each password before hashing. It's recommended this value is set to 16,
and there is no documented reason why you'd set it to anything other than this, however the minimum is 8.


#### parallelism
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: 8
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

This setting is specific to `argon2id` and unused with `sha512`. Sets the number of threads used when hashing passwords,
which affects the effective cost of hashing.


#### memory

This setting is specific to `argon2id` and unused with `sha512`. Sets the amount of memory allocated to a single
password hashing action. This memory is released by go after the hashing process completes, however the operating system
may not reclaim it until it needs the memory which may make Authelia appear to be using more memory than it technically
is.


## Passwords

The file contains hashed passwords instead of plain text passwords for security reasons.

You can use Authelia binary or docker image to generate the hash of any password. The
hash-password command has many tunable options, you can view them with the
`authelia hash-password --help` command. For example if you wanted to improve the entropy
you could generate a 16 byte salt and provide it with the `--salt` flag.
Example: `authelia hash-password --salt abcdefghijklhijl`. For argon2id the salt must
always be valid for base64 decoding (characters a through z, A through Z, 0 through 9, and +/).

Passwords passed to `hash-password` should be single quoted if using special characters to prevent parameter substitution.
For instance to generate a hash with the docker image just run:

    $ docker run authelia/authelia:latest authelia hash-password 'yourpassword'
    Password hash: $argon2id$v=19$m=65536$3oc26byQuSkQqksq$zM1QiTvVPrMfV6BVLs2t4gM+af5IN7euO0VB6+Q8ZFs

You may also use the `--config` flag to point to your existing configuration. When used, the values defined in the config will be used instead.

Full CLI Help Documentation:

```
Hash a password to be used in file-based users database. Default algorithm is argon2id.

Usage:
  authelia hash-password [password] [flags]

Flags:
      --config string     Configuration file
  -h, --help              help for hash-password
  -i, --iterations int    set the number of hashing iterations (default 1)
  -k, --key-length int    [argon2id] set the key length param (default 32)
  -m, --memory int        [argon2id] set the amount of memory param (in MB) (default 64)
  -p, --parallelism int   [argon2id] set the parallelism param (default 8)
  -s, --salt string       set the salt string
  -l, --salt-length int   set the auto-generated salt length (default 16)
  -z, --sha512            use sha512 as the algorithm (changes iterations to 50000, change with -i)
```

### Password hash algorithm

The default hash algorithm is Argon2id version 19 with a salt. Argon2id is currently considered
the best hashing algorithm, and in 2015 won the
[Password Hashing Competition](https://en.wikipedia.org/wiki/Password_Hashing_Competition).
It benefits from customizable parameters allowing the cost of computing a hash to scale
into the future which makes it harder to brute-force. Argon2id was implemented due to community
feedback as you can see in this closed [issue](https://github.com/authelia/authelia/issues/577).

For backwards compatibility and user choice support for the SHA512 algorithm is still available.
While it's a reasonable hashing function given high enough iterations, as hardware improves it
has a higher chance of being brute-forced.

Hashes are identifiable as argon2id or SHA512 by their prefix of either `$argon2id$` and `$6$`
respectively,  as described in this [wiki page](https://en.wikipedia.org/wiki/Crypt_(C)).

**Important Note:** When using argon2id Authelia will appear to remain using the memory allocated
to creating the hash. This is due to how [Go](https://golang.org/) allocates memory to the heap when
generating an argon2id hash. Go periodically garbage collects the heap, however this doesn't remove
the memory allocation, it keeps it allocated even though it's technically unused. Under memory
pressure the unused allocated memory will be reclaimed by the operating system, you can test
this on linux with:

    $ stress-ng --vm-bytes $(awk '/MemFree/{printf "%d\n", $2 * 0.9;}' < /proc/meminfo)k --vm-keep -m 1

If this is not desirable we recommend investigating the following options in order of most to least secure:
1. using the [LDAP authentication provider](./ldap.md)
2. adjusting the [memory](#memory) parameter
3. changing the [algorithm](#algorithm)


### Password hash algorithm tuning

All algorithm tuning for Argon2id is supported. The only configuration variables that affect
SHA512 are iterations and salt length. The configuration variables are unique to the file
authentication provider, thus they all exist in a key under the file authentication configuration
key called `password`. We have set what are considered as sane and recommended defaults
to cater for a reasonable system, if you're unsure about which settings to tune, please see the
parameters below, or for a more in depth understanding see the referenced documentation in
[Argon2 links](./file.md#argon2-links).


#### Examples for specific systems

These examples have been tested against a single system to make sure they roughly take
0.5 seconds each. Your results may vary depending on individual specification and
utilization, but they are a good guide to get started. You should however read the
linked documents in [Argon2 links](./file.md#argon2-links).

|    System     |Iterations|Parallelism|Memory |
|:------------: |:--------:|:---------:|:-----:|
|Raspberry Pi 2 |    1     |     8     |    64 |
|Raspberry Pi 3 |    1     |     8     |   128 |
|Raspberry Pi 4 |    1     |     8     |   128 |
|Intel G5 i5 NUC|    1     |     8     |  1024 |


## Argon2 Links

[How to choose the right parameters for Argon2](https://www.twelve21.io/how-to-choose-the-right-parameters-for-argon2/)

[Go Documentation](https://godoc.org/golang.org/x/crypto/argon2)

[IETF Draft](https://tools.ietf.org/id/draft-irtf-cfrg-argon2-09.html)
