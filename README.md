![hammerHEPimg](https://user-images.githubusercontent.com/20154956/27484126-5eba9f42-5828-11e7-9ac5-ceda711253df.png)

### Install:

Get it from the releases:
https://github.com/negbie/hammerHEP/releases

Or:
```bash
go get github.com/negbie/hammerHEP
```


### Usage of ./hammerHEP:

```bash
  -addr string
        Address to send packets (default "localhost")
  -port string
        Port to send packets (default "9060")
  -prot string
        Supported protocols are hep,ipfix (default "hep")
  -rate int
        How many packets per second to send (default 1)
        
################################################################

./hammerHEP -rate 1000
./hammerHEP -rate 1000 -prot ipfix -port 4739
```
