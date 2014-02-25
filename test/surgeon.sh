#!/bin/bash

rm -rf repo2-added.git
reposurgeon <<EOF
read <repo2.export
<b.Merge-from-a-2>..<a.Merge-to-b-2> merge
<b.Merge-from-a-1>..<a.Merge-to-b-1> merge
<a.Merge-from-c-1>..<c.Merge-to-a-1> merge
write >repo2-added.export
EOF

git init --bare repo2-added.git
GIT_DIR=repo2-added.git git fast-import < repo2-added.export
