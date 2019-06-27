# traceroute
This is a simple Study Case of traceroute in GO, using `icmp` and `IPV4` from the GO experimental packages. The main reason for creating this project is to better understand how TTL works under the hood in TCP package.
So this is a documentation of what I've gathered so far.

### TTL

TTL stands for Time To Live and it's an 8bit field in the IP Header. It's basically a security feature to avoid routing infinite loops between routers. 
When IP package is forwarded from one router to another, TTL decreases value by one. When  TTL reaches zero, the package will be discarded.
  
So in `traceroute`,  TTL is initially set to 1 until it receives an ICMP type "echo reply". To follow the hosts,  we must wait for an ICMP "time exceeded" reply from a gateway. After that we can record the round-trip delay and send another package with TTL, incremented by one.
  
So to implement that, we need to create a basic loop, that will represent the number of hops and increment the TTL in each loop interaction.

Like this:
```
for i := 1; i <= 64; i++ { // predefined hops size
		wm.Body.(*icmp.Echo).Seq = i
		wb, err := wm.Marshal(nil)
		if err != nil {
			log.Fatal(err)
		}
		if err := p.SetTTL(i); err != nil {
			log.Fatal(err)
		}
...
```
_This code is a sample I've found in the `go docs` that I've just adapted to suit what I needed._

Here we are creating a byte map to interact with and read the ICMP messages from the connection.

```
rb := make([]byte, 1500)
...
n, cm, peer, err := p.ReadFrom(rb)
...
```

### How to run?

```
go get github.com/isca0/traceroute
go build
sudo ./traceroute 
```
By default, it will query `redhat.cz` but you can use the flag `-host` to specify a different target to trace.  
_pps: This will open system sockets, so it must run as root._ :wink:

### Author

Igor Brandão  
Czech Challenge  

### References

* https://godoc.org/golang.org/x/net/ipv4
* https://godoc.org/golang.org/x/net/icmp
* https://linux.die.net/man/8/traceroute
