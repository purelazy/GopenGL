# How to create a new (this) repository on the command line

echo "# GopenGL" >> README.md
git init
git add README.md
git commit -m "first commit"
git remote add origin https://github.com/purelazy/GopenGL.git
git push -u origin master