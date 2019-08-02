
# Building the common-host-utils

The cleanest way to build in this package is to do the following:

1. Create an empty directory for your go workspace
   * ```➜  mkdir godemo &&  cd godemo```
2. Set your GOPATH to this directory
   * ```➜  export GOPATH=`pwd` ```
3. Get the repository using git or 'go get' (examples below)
   * Use git to obtain the repository
     * ```➜  git clone https://github.com/<username>/common-host-utils.git src/github.com/hpe-storage/common-host-utils```
   * Use 'go get' to obtain chapid
     * ```➜  go get -d github.com/hpe-storage/common-host-utils/cmd/chapid```
4. Change your working directory to the root of the repository
   * ```➜  cd src/github.com/hpe-storage/common-host-utils```
5. The tests are configured to run on linux/64, so set your GO OS
   * ```➜  export GOOS=darwin```
6. Build then entire repository to make sure everything compiles and tests
   * ```➜  make all```
