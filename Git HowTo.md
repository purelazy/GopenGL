# How to create a new (this) repository on the command line

echo "# GopenGL" >> README.md
git init
git add README.md

# Record changes to the repository

git commit -m "first commit"

# Adding a remote repository, e.g. github

The git remote add command takes two arguments:

A remote name, for example, origin
A remote URL. E.g.

git remote add origin https://github.com/purelazy/GopenGL.git

# git push: upload local repository to a remote repository.

git push -u origin master