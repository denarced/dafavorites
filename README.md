# dafavorites
Fetch user's favorite deviations from deviantart.com

# Usage
To fetch the source and build it:

    go get github.com/denarced/dafavorites ; \
        go get golang.org/x/net/html && \
        go install github.com/denarced/dafavorites/./...
  
It'll download the source code and build the binary. The running `dafavorites david` will fetch favorites for user _david_. The end result will be the deviations in a temporary directory and information on them in file _deviantFetch.json_. In the temporary directory each deviation is stored in its own sub directory in order to preserve the original filename. The sub directory names are UUIDs. It tries to also download the sometimes larger image available on the website via "Download" button. If the image is bigger than the smaller image linked to in the downloaded RSS it is kept. Both are.

# Future
This little tool was created solely for my own use. I use it to backup my favorite deviations because often enough the authors decide to remove their creations from Deviant Art. It's good enough for now so I have no plans to further develop it.

# Design
I needed some practice on parallel execution and I used this project. To be clear: the level of parallel execution is way too high. Not necessary at all. I added it merely to get some practice.
