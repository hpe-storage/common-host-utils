# Utilities

Common host utilities

## Build Instructions

The cleanest way to build in this package is to do the following:

1. Create an empty directory for your go workspace
   * ```➜  mkdir godemo &&  cd godemo```
1. Set your GOPATH to this directory
   * ```➜  export GOPATH=`pwd` ```
1. Get the repository using git or 'go get' (examples below)
   * Use git to obtain the repository
     * ```➜  git clone https://github.com/<username>/common-host-utils.git src/github.com/hpe-storage/common-host-utils```
   * Use 'go get' to obtain chapid
     * ```➜  go get -d github.com/hpe-storage/common-host-utils/cmd/chapid```
1. Change your working directory to the root of the repository
   * ```➜  cd src/github.com/hpe-storage/common-host-utils```
1. The tests are configured to run on linux/64, so set your GO OS
   * ```➜  export GOOS=darwin```
1. Build then entire repository to make sure everything compiles and tests
   * ```➜  make all```