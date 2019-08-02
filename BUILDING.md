
# Building the common-host-utils

The cleanest way to build in this package is to do the following:

1. Create an empty directory for your go workspace
   * ```➜  mkdir godemo &&  cd godemo```
2. Set your GOPATH to this directory
   * ```➜  export GOPATH=`pwd` ```
3. Fork the repo if not already done. `https://help.github.com/en/articles/fork-a-repo`
4. Get the repository using git or 'go get' (examples below)
   * Use git to obtain the repository
     * ```➜  git clone https://github.com/<username>/common-host-utils.git src/github.com/hpe-storage/common-host-utils```
   * Use 'go get' to obtain chapid
     * ```➜  go get -d github.com/hpe-storage/common-host-utils/cmd/chapid```
5. Change your working directory to the root of the repository
   * ```➜  cd src/github.com/hpe-storage/common-host-utils```
6. Build the entire repository to make sure everything compiles and tests
   * ```➜  make all```
