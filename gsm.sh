#!/bin/sh
#
# This script is an example that shows the order of steps. It will not work
# without modification.

# Directory where git-starteam has been installed.
GS_HOME=/path/to/git-starteam/jars

# Place to store temporary label files created by this script.
LABELS_TMP=$HOME/tmp

# Location of git repo to add merge information to.
GIT_REPO=$HOME/src/MyProject

# Classpath for git-starteam label dumper.
CP="$GS_HOME/jargs.jar:$GS_HOME/starteam110.jar:$GS_HOME/syncronizer.jar"

# Change the arguments to the label dumper program to match your StarTeam
# environment, such as the host, port, project name, and credentials.
java \
    -classpath "$CP" \
    org.sync.LabelDumper \
    -h starteamserver.example.com -P 49201 \
    -p MyProject -U myusername --password mypassword > $LABELS_TMP/labels

# Transform flat labels into CSV matching up merge labels.
gsm-labels -in $LABELS_TMP/labels > $LABELS_TMP/labels.csv

# Rewrite git history to add the merges.
cd $GIT_REPO
git filter-branch -f \
    --parent-filter "gsm-parent-filter -in $LABELS_TMP/labels.csv" \
    --tag-name-filter cat \
    -- --all
