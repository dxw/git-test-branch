# git-test-branch

Runs tests on each commit on a branch, concurrently.

## Installation

Requirements:

- flock
  - In Debian/Ubuntu the package is named `util-linux`
  - On macOS it appears that there's a brew package called `flock`

At this early stage I've been doing this:

```
% export PATH=${PATH}:/workbench/src/github.com/dxw/git-test-branch/bin/
```

## Usage

If you're on a branch with multiple commits, you might like to know that every commit is passing before it gets merged. You can do that with `git rebase -i origin/main -x 'composer install && vendor/bin/peridot spec` or you could run this:

```
git test-branch origin/main..@ 'composer install && vendor/bin/peridot spec'
```

This will run several tests concurrently (5 at the moment), and then print the pass/fail status of those tests.

Here's a trivial example where I've cut out some output of `git test-branch` (which may be removed later on):

```
% git test-branch origin/main..@ 'grep cidr composer.json'
[...]
951fea5 [PASS] Normalise addresses before storage
07b769d [PASS] composer require dxw/cidr
0971cb9 [FAIL] Add a passing test for IPv6
2f8f33d [FAIL] Do not expect current_time() to be called
4e095cf [FAIL] DRY specifying IP address
33cfb19 [FAIL] DRY class instantiation
04de2d8 [FAIL] Add psalm tests
bae8c5c [FAIL] composer require dxw/result
```
