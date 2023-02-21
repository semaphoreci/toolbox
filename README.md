# Toolbox

Toolset used in Semaphore 2.0 jobs.

## Installation

``` bash
# Install binaries
bash ~/.toolbox/install-toolbox

# Source functions into current session
source ~/.toolbox/toolbox

# Add toolbox to your bash_profile to activate it in every SSH session
echo 'source ~/.toolbox/toolbox' >> ~/.bash_profile
```

## How to add new tools

The process for adding new cli tools to the toolbox differs a bit depending on how they are written. If a new tool is written in
bash, like `sem-install`, `sem-service`, or `retry`, the process is pretty simple:
 
  - bash script will be copied into the toolbox during the `release` step inside the Semaphore CI/CD pipeline. (This happens 
  inside release/create.sh, function `hosted::create_initial_content`). The release process will be triggered only when a new release is
  created inside the GitHub repo.

  - In order for the new version of toolbox to be used on Semaphore's hosted agents, toolbox version should be updated [here](https://github.com/renderedtext/s2-platform/blob/master/Makefile). (Needles to say,
people outside of Semaphore can not perform this step)

  - Later, when the toolbox is being installed on agent machines before a job starts its execution, `install-toolbox` is called.
  In that script, binaries that have been packaged during the release process mentioned above are just linked, like here:

``` bash
install_cmd ln -sf ~/.toolbox/sem-context $INSTALL_PATH/sem-context
install_cmd chmod +x $INSTALL_PATH/sem-context
if [[ $? -eq 0 ]];then
  echo "sem-context installed"
else
  echo "toolbox_install_error{module='sem-context'} 1" >> /tmp/toolbox_metrics
fi
```

### What if your cli tool is written as a separate project in `Go`, for example, like `sem-context` or `cache-cli`

In that case, a few extra steps are neccessary:

  - You must compile those tools within the Semaphore CI/CD pipeline, and upload those binaries to artifacts, so that they would be available when
  we reach the release step within the pipeline. This is an example portion of `semaphore.yaml` for `sem-context`:
  
``` yaml
        - name: Build sem-context CLI
          commands:
            - sem-version go 1.17
            - "export GOPATH=~/go"
            - "export PATH=/home/semaphore/go/bin:$PATH"
            - checkout
            - cd sem-context
            - make build.linux
            - make build.darwin
            - artifact push workflow bin/linux/sem-context -d bin/linux/sem-context
            - artifact push workflow bin/darwin/sem-context -d bin/darwin/sem-context
```
  - After that, `release.yaml` should be changed to pull those binaries before the rest of the release process begins:

``` yaml
        - name: Create and Upload assets
          commands:
            - checkout
            - artifact pull workflow bin/linux/sem-context -d sem-context/bin/linux/sem-context
            - artifact pull workflow bin/darwin/sem-context -d sem-context/bin/darwin/sem-context
            - bash release/create.sh -a
```
  - `create.sh` script should also be altered to remove the source code for your cli tool from the final toolbox. This is done by listing the folder containing the source code inside the `exclusions`, within the `hosted::create_initial_content` function. Also, binaries that have been pulled from
artifacts at the beginning of the release process should be copied into the toolbox. This can be done in the `hosted::pack` function of the `create.sh` script:

``` bash
cp {directory_where_you_pulled_the_binaries} /tmp/Linux/toolbox/{binary_name}
```
  
