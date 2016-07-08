# dafavorites
Fetch user's favorite deviations from deviantart.com

# Usage
To build and fetch the favorites of user _david_:

    go get github.com/denarced/dafavorites && \
        ./dafavorites david
  
It'll

1. download the source code
2. build the binary
3. execute the binary

The end result will be the deviations in a temporary directory and information on them in file _deviantFetch.json_. In the temporary directory each deviation is stored in its own sub directory in order to preserve the original filename. The sub directory names are UUIDs.

# Future
This little tool was created solely for my own use. I use it to backup my favorite deviations because often enough the authors decide to remove their creations from Deviant Art. It's good enough for now so I have no plans to further develop it.
