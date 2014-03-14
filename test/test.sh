#!/bin/bash

cd $(dirname $0)

# Import the test repos.
# repo1 is the reference repo with merges.
# repo2 is the repo without merges that should be reconstructed.
for repo in repo1 repo2 ; do
    rm -rf $repo.git
    git init --bare $repo.git
    ( cd $repo.git && git fast-import < ../$repo.export )
done

# Convert the tags in repo2 to annotated tags.
(
    cd repo2.git
    git for-each-ref refs/tags |
    while read obj objtype ref ; do
        tag=${ref#refs/tags/}
        git tag -d "$tag"
        git tag -a -f -m "$tag" "$tag" "$obj"
    done
)

# Build gsm-add-merges
(
    cd ../gsm-add-merges
    go build
)

# Add the merges to repo2.
(
    cd repo2.git
    git fast-export --no-data --all | \
        ../../gsm-add-merges/gsm-add-merges -in ../repo2-merges.csv | \
        git fast-import --force
)

# Check that the number of revisions are equal
c1=$(GIT_DIR=repo1.git git rev-list --all | wc -l)
c2=$(GIT_DIR=repo2.git git rev-list --all | wc -l)
test $c1 -eq $c2 || {
    echo >&2 "ERROR: incorrect number of revisions: got $c2, want $c1"
    exit 1
}

echo "All tests passed"
