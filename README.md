git-starteam-merges
===================

Adds merge information to git repositories created by [git-starteam] [1].

Applicable to projects which use the following StarTeam merge policy:

  1. Merge source branches are labeled before the merge.
  2. The merge is based on that source label.
  3. The target is labeled immediately after the merge.
  4. The merge labels are named so that they can be automatically paired.

See `gsm.sh` for an example of how to use these programs together.

Installation
------------

First, [install Go] [2] if you don't already have version 1 or later. Then
run these two commands to install the git-starteam-merges programs:

    $ go get github.com/patrick-higgins/git-starteam-merges/gsm-labels
    $ go get github.com/patrick-higgins/git-starteam-merges/gsm-add-merges

Dumping Labels
--------------

These programs operate on a dump of labels created by the
`org.sync.LabelDumper` program from [git-starteam] [1]. It is assumed that
you have already used git-starteam to import your StarTeam project
into git, so you should be familiar with how to build git-starteam
from source and have all the jars needed by it.

gsm-labels
----------

The `gsm-labels` program requires knowledge of your merge labeling
conventions to be able to match up source and target merge labels.

The merge label conventions it supports are hard-coded and likely not
applicable to your codebase. Change the `FindMerges` method and the
`v1.go`, `v2.go`, and `v3.go` files as needed to support your own
conventions. In some cases, deleting `v2.go` and/or `v3.go` or
creating additional files such as `v4.go` and `v5.go` may be
appropriate depending on how many different labeling conventions exist
in your codebase.

gsm-add-merges
-----------------

The `gsm-add-merges` program is a filter for `git fast-export` that can
be piped to `git fast-import --force`.

It uses the CSV file produced by `gsm-labels` to create git merge commits
for each StarTeam merge.


  [1]: https://github.com/planestraveler/git-starteam                    "git-starteam"
  [2]: http://golang.org/doc/install                                     "install Go"
