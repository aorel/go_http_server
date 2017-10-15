## Go http server

### Usage
```bash
go run src/*.go --port=8080 -cpuMax=4 -root=/var/www/html
```

### Options
* -port -- Server port, defaul=80
* -cpuMax -- CPU limit, default=1
* -root -- Root directory, dafault="."


### Tests

#### Functional tests

[Tests repository](https://github.com/init/http-test-suite)

```bash
./httptest.py
```

#### Perfomance tests

```bash
ab -n 100000 -c 100 127.0.0.1:80/httptest/dir2/page.html
```
