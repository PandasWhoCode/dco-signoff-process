# Retroactive DCO Sign-Off

## Retroactive DCO Sign-off

### Background

Several repositories are missing DCO sign-off on historical commits. For commits by members of the Hashgraph organization the `hashgraph/platform-ci` team is authorized to perform a retroactive DCO sign-off.

In order to perform the sign-off the team must do the following:

- Install the `reveal_dco_issues` script in a local virtual environment
- Run `reveal_dco_issues` against a repository
- Configure the retroactive sign-off commit
- Enable pushing empty commits on a repository
- Push the empty commit to the repositories default branch

### Install `reveal_dco_issues` script

Create a virtual environment to operate the script out of

```bash
brew install python@3.12 virtualenv python-setuptools python-packaging python-certifi
mkdir "${HOME}/.virtualenvs"
cd "${HOME}/.virtualenvs"
	virtualenv lftools
source lftools/bin/activate
pip3 install setuptools
pip3 install lftools
```

Navigate to `${HOME}/.virtualenvs/lftools/bin`

```bash
pushd ${HOME}/.virtualenvs/lftools/bin
```

Add the below script to the file `reveal_dco_issues` in this directory 

```bash
#!/usr/bin/env bash

LINE="$(printf '%0.1s' "-"{1..80})"
readonly LINE

LT_LINE="$(printf '%0.1s' "<"{1..80})"
readonly LT_LINE

GT_LINE="$(printf '%0.1s' ">"{1..80})"
readonly GT_LINE

COMMITS_WO_SIGNOFF="$(lftools dco check)"

echo "${LT_LINE}"
echo "Summary of Unique Authors"
echo "${GT_LINE}"
echo
echo "${LINE}"
xargs -IID git log -1 --format='%an <%ae>' ID <<<"${COMMITS_WO_SIGNOFF}" | sort | uniq
echo "${LINE}"
echo
echo

echo "${LT_LINE}"
echo "Summary of Commits"
echo "${GT_LINE}"
echo
echo "${LINE}"
xargs -IID git log -1 --format='%H: %an <%ae>' ID <<<"${COMMITS_WO_SIGNOFF}" | sort | uniq
echo "${LINE}"
echo
echo

echo "${LT_LINE}"
echo "DCO Retroactive Message Format"
echo "${GT_LINE}"
echo
echo "${LINE}"
xargs -IID git log -1 --format='commit %h %s' ID <<<"${COMMITS_WO_SIGNOFF}" | sort | uniq
echo "${LINE}"
echo
echo

echo "${LT_LINE}"
echo "Full Commit Log History"
echo "${GT_LINE}"
echo
for commit in ${COMMITS_WO_SIGNOFF}; do
  echo "${LINE}"
  git log -1 "${commit}" | cat
done
echo "${LINE}"

```

Add executable permissions to the reveal_dco_issues file

```bash
chmod 755 reveal_dco_issues
```

### Running the `reveal_dco_issues` script

Activate the lftools virtual environment if needed

```bash
source ~/.virtualenvs/lftools/bin/activate
```

Navigate to a repository:

```bash
pushd ${REPO_HOME}/hedera-sdk-js
```

Execute the reveal_dco_issues script

```bash
reveal_dco_issues
```

Example output:

```
<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
Summary of Unique Authors
>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

--------------------------------------------------------------------------------
Alexander Gadzhalov <alexander.gadzhalov@limechain.tech>
Andrei Kuzmiankov <andrei@launchbadge.com>
Andrei Kuzmiankov <ender2016@gmail.com>
Austin Bonander <austin@launchbadge.com>
Brendan Graetz <bguiz@users.noreply.github.com>
Cooper Kunz <CooperAKunz@gmail.com>
Cooper Kunz <cooperakunz@gmail.com>
Daniel Akhterov <36461294+danielakhterov@users.noreply.github.com>
Daniel Akhterov <akhterovd@gmail.com>
Daniel Akhterov <daniel@launchbadge.com>
Evan Aspaas <aspaasevan@gmail.com>
Evan Aspaas <evan.aspaas@launchbadge.com>
Giuseppe Bertone <giuseppe.bertone@proton.me>
Jaewook Kim <cmdhema@gmail.com>
....
--------------------------------------------------------------------------------


<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
Summary of Commits
>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

--------------------------------------------------------------------------------
000705e8e806cbbfb2bf1d89ddeedb6f5886a193: Ryan Leckey <ryan@launchbadge.com>
0088413e2553783cfaf35735352c74fc167a0ce6: mehcode <leckey.ryan@gmail.com>
00a34720a703dd103d5e32af69dcebe78307a2a7: Ryan Leckey <ryan@launchbadge.com>
0143a3c821b72fdac4173ed202487f718e0bf3cf: Daniel Akhterov <akhterovd@gmail.com>
01717bfc9ddf5907c0a492e9b5d893b694a767bb: mehcode <leckey.ryan@gmail.com>
01a8e69274b1da7ab18c7d72936d2e2f4723ad46: Cooper Kunz <cooperakunz@gmail.com>
01cdd63d13659476ce0d6d6f611532b50a51ff48: keirabee <keira.black@launchbadge.com>
01ce189d18ec5b7c8d08083b7088e8f6812db0b0: Ryan Leckey <ryan@launchbadge.com>
01e66a302434ed0b83dc0109fa0a1264f9c722a2: Ryan Leckey <ryan@launchbadge.com>
024b0ea184c33d40f4555a65affb78381456ae2b: mehcode <leckey.ryan@gmail.com>
02c52508716a1f161ac6f24b4476568e154dc0d3: Austin Bonander <austin@launchbadge.com>
02cb2b3f23542c8c15b11ba6383a0af9b00d160f: Ryan Leckey <ryan@launchbadge.com>
033c9d94749937ca9389f74fa95eda9d4af58e02: Ryan Leckey <ryan@launchbadge.com>
03ab179fa45755590da2448e57192707ed86a4e8: mehcode <leckey.ryan@gmail.com>
03b47e4df2b5ff3895f495416071e71fa3936099: mehcode <leckey.ryan@gmail.com>
03d53ad82c0505760ea5e3b8cbe996dc26f38e9a: Ryan Leckey <ryan@launchbadge.com>
0421edac8d11fa35bf14cd3caa84af40b6ffcd9a: Daniel Akhterov <akhterovd@gmail.com>
0461479a4e0f82236b598c247f5f8b4deeebd8b4: Ryan Leckey <ryan@launchbadge.com>
...
--------------------------------------------------------------------------------


<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
DCO Retroactive Message Format
>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

--------------------------------------------------------------------------------
commit 000705e8 Update dependencies
commit 0088413e Deploying to gh-pages from master @ 7989356bd9eadfdabd45454874a1cf5eedacf827 🚀
commit 00a34720 See that generated/ ends up in both src/ and lib/
commit 0143a3c8 fix: package.json
commit 01717bfc Deploying to gh-pages from master - 1b8350fd514c36b30a8ff0855a8c7815248f8937 🚀
commit 01a8e692 rename to consensus-pub-sub.js
commit 01cdd63d Fixed some bugs.
commit 01ce189d Fix encoding with ContractFunctionParams with no arguments
commit 01e66a30 Run E2E tests in CI
commit 024b0ea1 Deploying to gh-pages from master @ d666e05c0124e3899d2c75b5890002a72b168d6c 🚀
commit 02c52508 install protobuf on CI
commit 02cb2b3f Re-factoring call params / func result
commit 033c9d94 Create node.yml
commit 03ab179f Deploying to gh-pages from master @ d97ef4a0a15758a4e46fe2ffbc72ee27554d9f77 🚀
commit 03b47e4d Deploying to gh-pages from master - de1173277c38bc861c73ae35131f9b2614d5f5a0 🚀
commit 03d53ad8 add dpdm to lints to check for circular imports
commit 0421edac fix: import proto directly when needed
commit 0461479a Update README.md
commit 047910a3 Deprecating existing setMaxGasAllowance method and adding setMaxGasAl… (#1144)
commit 04791c8b chore: update @hashgraph/proto
commit 04d7b147 Prepare CHANGELOG for v1.1.5
commit 04e0db10 Rename setPorxyAccount to setProxyAccountId
commit 04f2c4ba proto: add src/ to npm files
commit 05098c23 feat: re-add Key test from master branch
commit 053c473a Fix/get cost query (#1496)
commit 05d18879 fix: queries and transactions for node test
commit 05d3923d chore: update dependencies
commit 06d2b9b1 Update dependencies
commit 06d5d08c fix: Client and TopicMessageIntegrationTest
commit 06e7375b release(cryptography): v1.0.17
commit 06fbeb1b chore(wip): attempt to fix proto generation
commit 074a4780 fix `Keys.deriveSeed()` in the browser
commit 07977328 test: set an explicit node ID before freeze
commit 08e626ba chore: add some more documentation to the SDK (#1139)
...
--------------------------------------------------------------------------------


<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
Full Commit Log History
>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

--------------------------------------------------------------------------------
commit bf6d103bbe5bfad3c138ecf1f8ae5c4853df8adf
Author: Brendan Graetz <bguiz@users.noreply.github.com>
Date:   Thu Apr 20 20:14:27 2023 +0800

    fix: spelling of `HARDENED` exported constant (#1561)
....
--------------------------------------------------------------------------------

```

### Configure the retroactive sign-off commit

Fetch the latest commits

```bash
flow-*pushd ${REPO_HOME}/repo_to_check
git checkout main # or whatever the default branch is
git fetch origin
```

Execute the `reveal_dco_issues` script

```bash
reveal_dco_issues > ~/dco_issues.txt
```

Example output:

```
<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
DCO Retroactive Message Format
>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
commit 000705e8 Update dependencies
commit 0088413e Deploying to gh-pages from master @ 7989356bd9eadfdabd45454874a1cf5eedacf827 🚀
commit 00a34720 See that generated/ ends up in both src/ and lib/
commit 0143a3c8 fix: package.json
commit 01717bfc Deploying to gh-pages from master - 1b8350fd514c36b30a8ff0855a8c7815248f8937 🚀
commit 01a8e692 rename to consensus-pub-sub.js
commit 01cdd63d Fixed some bugs.
commit 01ce189d Fix encoding with ContractFunctionParams with no arguments
commit 01e66a30 Run E2E tests in CI
commit 024b0ea1 Deploying to gh-pages from master @ d666e05c0124e3899d2c75b5890002a72b168d6c 🚀
commit 02c52508 install protobuf on CI
commit 02cb2b3f Re-factoring call params / func result
commit 033c9d94 Create node.yml
....
--------------------------------------------------------------------------------
```

Note that only the commits listed not the header/footer of the section need to be copied

Execute the git commit command to create an empty git commit
```bash
git commit --allow-empty --signoff --gpg-sign
```

Fill in the commit message with the output from the `reveal_dco_issues` script, ensuring to include the DCO sign-off message at the end of the commit message
```bash
### This first line is the commit header that allows github to know that we're doing retroactive signoff ###
I, [Commit Author First Name] [Commit Author Last Name], retroactively sign off on these commits:

### The below lines are the DCO Retroactive Message Format from the reveal_dco_issues script output. ###
commit 000705e8 Update dependencies
commit 0088413e Deploying to gh-pages from master @ 7989356bd9eadfdabd45454874a1cf5eedacf827 🚀
commit 00a34720 See that generated/ ends up in both src/ and lib/
commit 0143a3c8 fix: package.json
commit 01717bfc Deploying to gh-pages from master - 1b8350fd514c36b30a8ff0855a8c7815248f8937 🚀
commit 01a8e692 rename to consensus-pub-sub.js
commit 01cdd63d Fixed some bugs.
commit 01ce189d Fix encoding with ContractFunctionParams with no arguments
commit 01e66a30 Run E2E tests in CI
commit 024b0ea1 Deploying to gh-pages from master @ d666e05c0124e3899d2c75b5890002a72b168d6c 🚀
commit 02c52508 install protobuf on CI
commit 02cb2b3f Re-factoring call params / func result
commit 033c9d94 Create node.yml

### All lines below here are automatically generated from the git commit command ###

Signed-off-by: [Commit Author First Name] [Commit Author Last Name] <user.email@domain.com>

# Please Enter the commit message for your changes. Lines starting
# with '#' will be ignored, and an empty message aborts the commit.

# On branch main
# Your branch is up to date with 'origin/main'.
```

Note the lines in the commit message above that begin and end with ### are not typically part of the commit message and do not need to be included

## Enable pushing an empty commit to the repository (if necessary)

Navigate to the repository settings: `https://github.com/<org>/<repository>/settings/rules`

Add `platform-ci` to the bypass list for the following rules

- [Branch] Layer 1a: Basic Limits - Force Push
- [Branch] Layer 2: PR Requirements
- [Branch] Layer 3: Basic Status Check Requirements
- [Branch] Layer 3a: Basic Status Check Requirements - Code Compiles
- [Branch] Layer 5: PR Formatting Status Check Requirements
- Additional rules as needed

### Push the commit to the repository

Since the empty commit is already staged and ready to go we just need to push up to the repository

```bash
git push -u origin main --no-verify
```

### Cleanup

Delete the file `~/dco_issues.txt`

```bash
rm ~/dco_issues.txt
```

- Reset GitHub Repository Ruleset Bypasses

  Navigate to the repository settings: `https://github.com/<org>/<repository>/settings/rules`

  Remove `platform-ci` from the bypass list for the following rules

  - [Branch] Layer 1a: Basic Limits - Force Push
  - [Branch] Layer 2: PR Requirements
  - [Branch] Layer 3: Basic Status Check Requirements
  - [Branch] Layer 3a: Basic Status Check Requirements - Code Compiles
  - [Branch] Layer 5: PR Formatting Status Check Requirements
  - Additional rules as needed
