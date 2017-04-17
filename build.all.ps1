$gocmd='github.com/deviceio/cli/cmd/deviceio-cli'

$env:GOOS='windows';$env:GOARCH='amd64';go install std; go build -o $env:GOPATH\bin\deviceio-cli.windows.amd64.exe $gocmd 
$env:GOOS='windows';$env:GOARCH='386';go install std; go build -o $env:GOPATH\bin\deviceio-cli.windows.386.exe $gocmd 

$env:GOOS='linux';$env:GOARCH='amd64';go install std; go build -o $env:GOPATH\bin\deviceio-cli.linux.amd64 $gocmd
$env:GOOS='linux';$env:GOARCH='386';go install std; go build -o $env:GOPATH\bin\deviceio-cli.linux.386 $gocmd

$env:GOOS='darwin';$env:GOARCH='amd64';go install std; go build -o $env:GOPATH\bin\deviceio-cli.darwin.amd64 $gocmd
$env:GOOS='darwin';$env:GOARCH='386';go install std; go build -o $env:GOPATH\bin\deviceio-cli.darwin.386 $gocmd