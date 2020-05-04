# KBF

Kubernetes yml port forwarding

# Use

Usage:
  kbf connect [flags]

 Flags:
   -f, --file string   forward file path (default $PWD/forward.yml)
   -h, --help          help for connect

## Forward file structure

```yaml
services:
  - name: service-01
    namespace: default
    port: 8080
    targetPort: 880
  - name: service-02
    namespace: default
    port: 8081
    targetPort: 8080
```

 # Build

```bash
git clone https://github.com/eduardobarbosa/kbf.git
cd kbf
go build
./kbf --help
```

# Install using Homebrew

```bash
brew tap eduardobarbosa/core
brew install eduardobarbosa/core/kbf
```