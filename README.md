# Chooser

A tool to compare which go packages will need be to rebuilt as a result of changes between two git diffs

# Running

    make
    chooser -dir github.com/package/here/ -from HEAD -to HEAD~1

# TODO

- Currently Naive approach using git,go list. Try change to go/build package
