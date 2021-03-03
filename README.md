# udpsigner
​
Boneh–Lynn–Shacham multi-party threshold signatures

This is a simple framework to run MPC tests.
It uses UDP discovery to find other nodes, based on the project:
https://github.com/schollz/peerdiscovery 

It also uses 
https://github.com/manifoldco/promptui 
to provide a simplistic stateful CLI

Everything is "under construction" atm, but it should work.
The simplest way to test it is to use a cli for generating Shamir Secret Shares of an BN256 BLS signatures using   
https://github.com/san-lab/secretsplitcli 
(for docker, you probably want to run the containers mapping the current directory with sharefiles, something like:  
 "docker run -it --rm -v $(pwd):/data udpsigner",
 and inside the docker import from "/data/youkeysharefile.json")             

TESTING  
Use docker to build the image.
Then run a few instances of the udpsigner in separate docker containers.
They should discover one-another automatically. 
Generate some SSS key shares using secretsplitcli, with any number of shares and threshold (within reason ;) ). 
Then import respective keyfiles with shares into different nodes and issue a "job request" from ine of the nodes.
The other nodes can then "aprove" the job request (and contribute to the MPC), or reject it.
If enough nodes agree to collaborate, a valid signature (or a vaild Public Key) will be generated.

As of beginnig of March 2021, there is only a couple of MPC algorithms: to calculate a BLS public key and a BLS message signature based on the (threshold) key shares owned by the nodes.


DEPENDENCIES/BUILD  
Getting the code: 

    git clone https://github.com/san-lab/secretsplitcli
    git clone https://github.com/san-lab/udpsigner.git 

Explicit dependencies:

    go get golang.org/x/crypto/sha3
    go get github.com/schollz/peerdiscovery
    go get github.com/google/uuid
    go get go.dedis.ch/kyber/pairing
    go get github.com/manifoldco/promptui
    go get golang.org/x/term

