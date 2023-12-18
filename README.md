# Instructions and notes about mcli

## About git submodules

<https://www.vogella.com/tutorials/GitSubmodules/article.html>

to update submodule

    # update submodule in the master branch
    # skip this if you use --recurse-submodules
    # and have the master branch checked out
    export GitSubmodule="plugins/default_plugins/"
    cd ${GitSubmodule}
    git checkout main
    git pull
    git add ${GitSubmodule}

    # commit the change in main repo
    # to use the latest commit in master of the submodule
    cd ..
    git add ${GitSubmodule}
    git commit -m "move submodule ${GitSubmodule} to latest commit in main branch"

    # share your changes
    git push

## Delete a submodule from a repository

Currently Git provides no standard interface to delete a submodule. To remove a submodule called mymodule you need to:

    git submodule deinit -f — ${GitSubmodule}
    rm -rf .git/modules/${GitSubmodule}
    git rm -f ${GitSubmodule}
