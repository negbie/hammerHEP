![hammerHEPimg](https://user-images.githubusercontent.com/20154956/27484126-5eba9f42-5828-11e7-9ac5-ceda711253df.png)

### Install:

Get it from the releases:
https://github.com/negbie/hammerHEP/releases

Or:
```bash
go install github.com/negbie/hammerHEP
```


### Usage of ./hammerHEP:

```bash
  -address string
    	Destination Address (default "localhost")
  -port string
    	Destination Port (default "9060")
  -protocol string
    	Possible protocols are HEP,IPFIX (default "HEP")
  -rate int
    	Packets per second (default 16)
  -transport string
    	Possible transports are UDP,TCP,TLS (default "TLS")
     
################################################################

./hammerHEP -rate 100
```
