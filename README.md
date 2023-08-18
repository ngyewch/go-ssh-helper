# go-ssh-helper

## Supported ssh_config keywords

| Keyword                          | Supported |
|----------------------------------|-----------|
| Host                             | Yes       |
| Match                            | No        |
| AddKeysToAgent                   | No        |
| AddressFamily                    | Yes       |
| BatchMode                        |           |
| BindAddress                      |           |
| BindInterface                    |           |
| CanonicalDomains                 |           |
| CanonicalizeFallbackLocal        |           |
| CanonicalizeHostname             |           |
| CanonicalizeMaxDots              |           |
| CanonicalizePermittedCNAMEs      |           |
| CASignatureAlgorithms            |           |
| CertificateFile                  |           |
| CheckHostIP                      |           |
| Ciphers                          |           |
| ClearAllForwardings              |           |
| Compression                      |           |
| ConnectionAttempts               |           |
| ConnectTimeout                   | Yes       |
| ControlMaster                    |           |
| ControlPath                      |           |
| ControlPersist                   |           |
| DynamicForward                   |           |
| EnableEscapeCommandline          |           |
| EnableSSHKeysign                 |           |
| EscapeChar                       |           |
| ExitOnForwardFailure             |           |
| FingerprintHash                  |           |
| ForkAfterAuthentication          |           |
| ForwardAgent                     |           |
| ForwardX11                       |           |
| ForwardX11Timeout                |           |
| ForwardX11Trusted                |           |
| GatewayPorts                     |           |
| GlobalKnownHostsFile             |           |
| GSSAPIAuthentication             |           |
| GSSAPIDelegateCredentials        |           |
| HashKnownHosts                   |           |
| HostbasedAcceptedAlgorithms      |           |
| HostbasedAuthentication          |           |
| HostKeyAlgorithms                |           |
| HostKeyAlias                     |           |
| Hostname                         | Yes       |
| IdentitiesOnly                   |           |
| IdentityAgent                    |           |
| IdentityFile                     | Yes       |
| IgnoreUnknown                    |           |
| Include                          | Yes       |
| IPQoS                            |           |
| KbdInteractiveAuthentication     |           |
| KbdInteractiveDevices            |           |
| KexAlgorithms                    |           |
| KnownHostsCommand                |           |
| LocalCommand                     |           |
| LocalForward                     |           |
| LogLevel                         | No        |
| LogVerbose                       | No        |
| MACs                             |           |
| NoHostAuthenticationForLocalhost |           |
| NumberOfPasswordPrompts          |           |
| PasswordAuthentication           |           |
| PermitLocalCommand               |           |
| PermitRemoteOpen                 |           |
| PKCS11Provider                   | No        |
| Port                             | Yes       |
| PreferredAuthentications         |           |
| ProxyCommand                     |           |
| ProxyJump                        | Yes       |
| ProxyUseFdpass                   |           |
| PubkeyAcceptedAlgorithms         |           |
| PubkeyAuthentication             |           |
| RekeyLimit                       |           |
| RemoteCommand                    |           |
| RemoteForward                    |           |
| RequestTTY                       |           |
| RequiredRSASize                  |           |
| RevokedHostKeys                  |           |
| SecurityKeyProvider              | No        |
| SendEnv                          |           |
| ServerAliveCountMax              |           |
| ServerAliveInterval              |           |
| SessionType                      |           |
| SetEnv                           |           |
| StdinNull                        |           |
| StreamLocalBindMask              |           |
| StreamLocalBindUnlink            |           |
| StrictHostKeyChecking            |           |
| SyslogFacility                   |           |
| TCPKeepAlive                     |           |
| Tag                              |           |
| Tunnel                           |           |
| TunnelDevice                     |           |
| UpdateHostKeys                   |           |
| User                             | Yes       |
| UserKnownHostsFile               |           |
| VerifyHostKeyDNS                 |           |
| VisualHostKey                    |           |
| XAuthLocation                    |           |
