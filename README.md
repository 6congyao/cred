# cred

A distributed daemon based on etcdv3 for IAM

## Installation

On MacOS you can install or upgrade to the latest released version with Homebrew:
```sh
$ brew install dep
$ brew upgrade dep
```

On other platforms you can use the `install.sh` script:

```sh
$ curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
```

Then clone the repository:
```sh
$ git clone <repository>
```

Run `dep` and enjoy:
```sh
$ dep ensure
```

Prepare the envs :
```sh
export CRED_PORT=<port> (Default: 9011)
export CRED_META_URL=<meta_url>
export CRED_STS_URL=<sts_url>
export CRED_SYNC_TTL=<sync_ttl> (Default: 3600)
export CRED_LOG_LEVEL=<debug/info/warning/error> (Default: error)
```

Run cred
```sh
$ credsvc
```

Run cred with mock mode, neither http server nor sts involved
```sh
$ credsvc -mock
```

## Features:

1. Cluster with self-registration
2. Watcher mutex based on distributed lock
3. Http API for operation and maintenance
4. Mock mode support skipping the interactive with sts for easily testing
