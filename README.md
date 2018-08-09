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
$ CRED_PORT=<port> (Default: 9011)
$ CRED_META_URL=<meta_url>
$ CRED_STS_URL=<sts_url>
$ export CRED_PORT
$ export CRED_META_URL
$ export CRED_STS_URL
```

## Todo:

Integration with sts

Distributed lock