---
layout: default
title: Server
parent: Configuration
nav_order: 9
---

# Server

The server section configures and tunes the http server module Authelia uses.

## Configuration

```yaml
server:
  host: 0.0.0.0
  port: 9091
  tls_key: ""
  tls_cert: ""
  read_buffer_size: 4096
  write_buffer_size: 4096
  path: ""
  enable_pprof: false
  enable_expvars: false
```

## Options

### host
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: 0.0.0.0
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Defines the address to listen on. See also [port](#port). Should typically be `0.0.0.0` or `127.0.0.1`, the former for
containerized environments and the later for daemonized environments like init.d and systemd.

Note: If utilising an IPv6 literal address it must be enclosed by square brackets and quoted:

```yaml
host: "[fd00:1111:2222:3333::1]"
```

### port
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: 9091
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Defines the port to listen on. See also [host](#host).

### tls_key
<div markdown="1">
type: string (path)
{: .label .label-config .label-purple } 
default: ""
{: .label .label-config .label-blue }
required: situational
{: .label .label-config .label-yellow }
</div>

The path to the private key for TLS connections. Must be in DER base64/PEM format.

Authelia's typically listens for plain unencrypted connections. This is by design as most environments allow to
security on lower areas of the OSI model. However it required, if you specify both of this option and the 
[tls_cert](#tls_cert) options, Authelia will listen for TLS connections.

### tls_cert
<div markdown="1">
type: string (path)
{: .label .label-config .label-purple } 
default: ""
{: .label .label-config .label-blue }
required: situational
{: .label .label-config .label-yellow }
</div>

The path to the public certificate for TLS connections. Must be in DER base64/PEM format.

Authelia's typically listens for plain unencrypted connections. This is by design as most environments allow to
security on lower areas of the OSI model. However it required, if you specify both of this option and the
[tls_key](#tls_key) options, Authelia will listen for TLS connections.

### read_buffer_size
<div markdown="1">
type: integer 
{: .label .label-config .label-purple } 
default: 4096
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Configures the maximum request size. The default of 4096 is generally sufficient for most use cases.

### write_buffer_size
<div markdown="1">
type: integer 
{: .label .label-config .label-purple } 
default: 4096
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Configures the maximum response size. The default of 4096 is generally sufficient for most use cases.

### path
<div markdown="1">
type: string 
{: .label .label-config .label-purple } 
default: ""
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Authelia by default is served from the root `/` location, either via its own domain or subdomain.

Modifying this setting will allow you to serve Authelia out from a specified base path. Please note
that currently only a single level path is supported meaning slashes are not allowed, and only
alphanumeric characters are supported.

Example: https://auth.example.com/, https://example.com/
```yaml
server:
  path: ""
```

Example: https://auth.example.com/authelia/, https://example.com/authelia/
```yaml
server:
  path: authelia
```

### enable_pprof
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Enables the go pprof endpoints.

### enable_expvars
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Enables the go expvars endpoints.


## Additional Notes

### Buffer Sizes

The read and write buffer sizes generally should be the same. This is because when Authelia verifies
if the user is authorized to visit a URL, it also sends back nearly the same size response as the request. However
you're able to tune these individually depending on your needs.
