# gRPC vs. REST: Performance Simplified

> From <https://medium.com/@bimeshde/grpc-vs-rest-performance-simplified-fd35d01bbd4>

“Breaking down the monolith”. These were words I heard several times over the course of my previous internships. Companies everywhere are realizing the benefits of building a microservice-based architecture. From lower costs to better performance to less downtime, microservices provide countless benefits relative to their preceding monolithic design. Now with all of these microservices talking to each other thousands of times each second, communication between them needs to be fast and reliable. The traditional method of doing this is JSON-backed HTTP/1.1 REST communication. However, alternatives such as gRPC provide significant benefits in performance, cost, and convenience.

## Problem

When classes inside a monolithic service communicate with each other, they do so through well-defined interfaces. These interfaces come with language-native objects to use to pass into and accept from them. Most errors in format and usage would be caught by the compiler and no new objects have to be created by consumers. Any conversion that happens between objects through converters and populators is done at a binary level, and not into a human-readable format.

Compare this to a microservice-based design. Whenever we are trying to consume a new service, we need to build our own objects using their API documentation, making sure the field types and names match up exactly. Then, we need to convert our data into this new object. Next, we need to convert this object into JSON using some converter. Finally, we would perform this entire process again in reverse when accepting responses from the API. This whole process causes two major problems: poor performance and slow development.

## Comparison

Considering the problems with the status quo, JSON-backed REST over HTTP/1.1, I’ll be comparing it to a solution that I argue is much better suited for the microservice paradigm we find ourselves in today: gRPC. To compare there effectiveness, I have three major constraints:

- Language-neutral: we want the flexibility to use the best technologies for the job
- Easy to use: development speed is essential
- Fast: every extra millisecond ends up losing customers and costing thousands of dollars in the long run

------

## Language and platform support

### REST and JSON

REST has support from nearly every type of environment. From backend applications to mobile to web, REST and HTTP/1.1 just work.

For JSON, libraries exist for nearly every language in existence and it’s the default content type assumed for many REST-based services. And at worst, you could construct JSON using strings of text since JSON really is just plain text formatted in a specific way.

### gRPC and Protocol Buffers

From gRPC’s official website [1], it’s supported by most popular languages and platforms: C++, Java (including Android), Python, Go, Ruby, C#, Objective-C (including iOS), JavaScript (Node.js and browsers) and more. However, support for many of these platforms is new and in turn arguably not mature enough for production use. For example, we had many issues using grpc-web for browser gRPC support, some due to lack of features and others due to lack of documentation.

For Protocol Buffers as well, libraries for many of the supported languages aren’t as well developed as the libraries for C++ and Java. This was very apparent when looking for documentation or trying to create a code generation plugin using any of the less popular languages: the functionality we wanted was either buried in responses to GitHub issues or not implemented at all.

### Outcomes

While we were eventually able to build everything we wanted with to do gRPC and Protocol Buffers in the languages we were working with, JSON definitely has much better support and documentation in most of these languages. That’s why we decided whenever starting a project in a new language, we need to confirm that gRPC support existed to the extent we needed.

------

## Connection — HTTP/2 vs. HTTP/1.1

Note: HTTP/2 is required by gRPC but also usable with REST. This isn’t really a fair comparison since HTTP/2 was built to address many of the pain points of HTTP/1.1. Here are some of the key problems with HTTP/1.1, along with their solutions in HTTP/2:

### Concurrent Requests

Concurrent requests in HTTP/1.1 aren’t supported [2]. However, since this is essential to modern applications, several workarounds are used by HTTP/1.1 to create this functionality. *Pipelining* works by a client sending multiple requests to a server before receiving a response [3]. This allows the server to process all requests in parallel and send the responses back when they are done. However, a limitation of this is that the responses still have to be sent back in the same order as the requests came in. This leads to something called *head-of-line blocking*, which means later requests are stalled by waiting for requests that were sent before them to complete [2]. Some other examples of workarounds used by web consumers include *spriting* (putting all of your images into one image) and *concatenation* (combining JavaScript files), both of which reduce the number of HTTP requests that a client has to make [2].

In HTTP/2 however, none of these workarounds are needed and are actually counterproductive in many cases. HTTP/2 natively supports *request multiplexing* [5], which allows for an unbounded amount of requests to be made and responded to concurrently and asynchronously. This is achieved by allowing multiple simultaneously open *streams* of data on a single TCP connection. Then, when *frames* of data are sent over this connection, they contain a *stream identifier.* This identifier is sent and used by both the client and server to identify which stream each frame is for [4].

### Request Framing

Requests and responses HTTP/1.1 are entirely in plaintext. Everything is delimited by newline characters, including where the headers and payload end. This, along with optional whitespace characters and varying termination patterns based on request and response type lead to confusing implementation, and in turn many parsing and security errors [2].

HTTP/2 is different since the headers and the payload are separated into their own frames. Each frame starts with a nine-byte header that specifies the frame length, type, stream, and some flags [3]. The separation of the headers and payload allow for better header compression using a new algorithm called *HPACK*, which works by using various compression methods (Static Dictionary, Dynamic Dictionary, and Huffman Encoding) that are specific to headers, yielding more than two times better compression than gzip performed by TLS with HTTP/1.1 [7]. For example, Static Dictionary compresses the 61 most common headers down to only one byte!

### Benchmarks

I created a simple Go server that supports HTTP/2 and HTTP/1.1 with an endpoint supporting GET requests. The endpoint had to be exposed via HTTPS since HTTP/2 is only supported over TLS.

`main.go`

```golang
package main

import (
	"log"
	"encoding/json"
	"net/http"
	"github.com/Bimde/grpc-vs-rest/pb"
)

func handle(w http.ResponseWriter, _ *http.Request) {
	random := pb.Random{RandomString: "a_random_string", RandomInt: 1984}
	bytes, err := json.Marshal(&random)

	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func main() {
	server := &http.Server{Addr: "bimde:8080", Handler: http.HandlerFunc(handle)}
	log.Fatal(server.ListenAndServeTLS("server.crt", "server.key"))
}
```

Then, I wrote a client-side method that consumed the endpoint.

`http_test.go`

```golang
package main

import (
	"log"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

var client http.Client

func init() {
	client = http.Client{}
}

func get(path string, output interface{}) error {
    req, err := http.NewRequest("GET", path, nil)
    if err != nil {
        log.Println("error creating request ", err)
        return err
    }
	
    res, err := client.Do(req)
    if err != nil {
        log.Println("error executing request ", err)
        return err
    }

    bytes, err := ioutil.ReadAll(res.Body)
    if err != nil {
        log.Println("error reading response body ", err)
        return err
    }

    err = json.Unmarshal(bytes, output)
    if err != nil {
        log.Println("error unmarshalling response ", err)
        return err
    }

    return nil
}
```

Finally, I created benchmarks using Go’s built-in benchmarking tool using HTTP/1.1 and HTTP/2 transports. The general idea is to test how quickly a particular transport could execute a specific number of requests (this number is chosen by the benchmarking tool).

Note that the custom local certificate pool was required because of the certificate was created locally and not issued by a trusted certificate authority. The code below (which I took from an online tutorial [6]):

`tls.go`

```golang
// This code was taken from https://posener.github.io/http2/
func createTLSConfigWithCustomCert() *tls.Config {
	// Create a pool with the server certificate since it is not signed
	// by a known CA
	caCert, err := ioutil.ReadFile("server/server.crt")
	if err != nil {
		log.Fatalf("Reading server certificate: %s", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create TLS configuration with the certificate of the server
	return &tls.Config{
		RootCAs: caCertPool,
	}
}
```

Then, for the HTTP/2 test, I was able to just spin up new goroutines (similar, but more lightweight compared to new threads) for each call and run thousands of requests in parallel.

`http2_test.go`

```golang
func BenchmarkHTTP2Get(b *testing.B) {
	client.Transport = &http2.Transport{
		TLSClientConfig: createTLSConfigWithCustomCert(),
	}

	var wg sync.WaitGroup
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			get("https://bimde:8080", &pb.Random{})
			wg.Done()
		}()
	}
	wg.Wait()
}
```

This resulted in an average of about 350 ms per request when running 10000 requests at once. This is quite slow, so we’ll address that later on.

Trying the same thing with HTTP/1.1, however, yielded this error:

```text
Get https://bimde:8080: dial tcp 127.0.1.1:8080: socket: too many open files
```

HTTP/1.1 just didn’t support that many connections at once (since HTTP/1.1 needs multiple TCP connections for concurrent requests).

So, I implemented a Job/Worker pattern [9] to control how many concurrent requests were being executed. This works by having a queue that the test adds jobs to (I’m using a channel in Go), and workers, who consume jobs from this queue as quickly as they can. The number of concurrent requests is dependent on the number of goroutines created, which is defined by the `noWorkers` variable below.

`job_worker_helpers.go`

```golang
type Request struct {
	Path string
	Random *pb.Random
}

const stopRequestPath = "STOP"

func startWorkers(requestQueue *chan Request, noWorkers int) func() {
	var wg sync.WaitGroup
	wg.Add(noWorkers)
	for i := 0; i < noWorkers; i++ {
		startWorker(requestQueue, &wg)
	}
	// Returns a function that stops as many workers as were just started
	return func() {
		stopRequest := Request{Path: stopRequestPath}
		for i := 0; i < noWorkers; i++ {
			*requestQueue <- stopRequest
		}
		wg.Wait()
	}
}

func startWorker(requestQueue *chan Request, wg *sync.WaitGroup) {
	go func() {
		for {
			request := <- *requestQueue
			if (request.Path == stopRequestPath) {
				wg.Done()
				return
			}
			get(request.Path, request.Random)
		}
	}()
}
```

Here’s the HTTP/1.1 test:

`http1_workers_test.go`

```golang
const noWorkers = 2

func BenchmarkHTTP11Get(b *testing.B) {
	client.Transport = &http.Transport{
		TLSClientConfig: createTLSConfigWithCustomCert(),
	}
	requestQueue := make(chan Request)
	defer startWorkers(&requestQueue, noWorkers)()
	b.ResetTimer() // don't count worker initialization time
	for i := 0; i < b.N; i++ {
		requestQueue <- Request{Path: "https://bimde:8080", Random: &pb.Random{}}
	}
}
```

Here’s an updated HTTP/2 test:

`http2_workers_test.go`

```golang
func BenchmarkHTTP2GetWithWokers(b *testing.B) {
	client.Transport = &http2.Transport{
		TLSClientConfig: createTLSConfigWithCustomCert(),
	}
	requestQueue := make(chan Request)
	defer startWorkers(&requestQueue, noWorkers)()
	b.ResetTimer() // don't count worker initialization time
	for i := 0; i < b.N; i++ {
		requestQueue <- Request{Path: "https://bimde:8080", Random: &pb.Random{}}
	}
}
```

Using this pattern I was finally able to get reasonable results for both HTTP/1.1 and HTTP/2.

![Image for post](https://miro.medium.com/max/1200/1*RwTT_iWkGhTY3S78l7ixsw.png)

If you notice, the runtime per request for HTTP/1.1 starts out better than HTTP/2 using a single goroutine (and in turn one request at a time over a single TCP connection). However, as the processing demands start to increase and the number of simultaneous workers increases, HTTP/1.1 quickly starts to fall apart. Using just 4 simultaneous connections brings HTTP/1.1 to its knees.

HTTP/2, on the other hand, just keeps on scaling. Even at 32 simultaneous streams, the runtime/request just keeps on going down. Below is another chart, this time testing the limits of HTTP/2.

![Image for post](https://miro.medium.com/max/1200/1*nOqWv1NYDEUkR-clTvfu7g.png)

As you can see, HTTP/2 only really starts to fall apart at over 500 concurrent streams over a single TCP connection. That’s a ridiculous improvement over the 4 connections of HTTP/1.1.

### Adoption

While almost every device browser in use right now supports HTTP/1.1, only \~70% of clients support HTTP/2. This would mean we’d need to support both protocols to support all clients. Ideally, all of our services could support HTTP/2 and fallback onto HTTP/1.1 for pre-existing services not yet upgraded.

### Outcomes

The key benefit of HTTP/1.1 is a wider adoption by the general public. Due to the massive performance advantage at scale, HTTP/2 is a no-brainer for internal communication, at the very least. This narrows down our decision to either REST with HTTP/2 or gRPC (which only supports HTTP/2).

------

## Ease of Use

### Writing Code

As discussed before, REST APIs are a very general specification that’s accessible from anywhere. This generally makes actually making these REST requests more verbose than they need to be. This is especially true considering the requirement to convert language-based objects to JSON and back from JSON to language-based objects in order to make a REST request. Here’s an example of a minimal Go function that makes a POST request using a struct as input and another struct for output using the built-in HTTP and JSON libraries:

`http_client_post_request.go`

```golang
func Post(path string, input interface{}, output interface{}) error {
  data, err := json.Marshal(input)
  if err != nil {
    log.Println("error marshalling input ", err)
    return err
  }
  
  body := bytes.NewBuffer(data)
  req, err := http.NewRequest("POST", path, body)
  if err != nil {
    log.Println("error creating request ", err)
    return err
  }
  
  var client http.Client
  res, err := client.Do(req)
  if err != nil {
    log.Println("error executing request ", err)
    return err
  }
  
  bytes, err := ioutil.ReadAll(res.Body)
  if err != nil {
    log.Println("error reading response body ", err)
    return err
  }
  
  err = json.Unmarshal(bytes, output)
  if err != nil {
    log.Println("error unmarshalling response ", err)
    return err
  }
  
  return nil
}
```

Word count: 103

Here’s trying to achieve the same thing using gRPC and Protocol Buffers:

`random.proto`

```golang
syntax = "proto3";

package pb;

message Random {
  string randomString = 1;
  int32 randomInt = 2;
}

service RandomService {
  rpc DoSomething (Random) returns (Random) {}
}
```

`random.go`

```golang
func random(c *context.Context, input *pb.Random) (*pb.Random, error) {
  conn, err := grpc.Dial("sever_address:port")
  if err != nil {
    log.Fatalf("Dial failed: %v", err)
  }
  
  client := pb.NewRandomServiceClient(conn)
  return client.DoSomething(c)
}
```

Word count: 53

As you can see, consuming gRPC endpoints is definitely less code than consuming REST endpoints (especially since you only need to perform the dial once). However, upon closer inspection of the code, you could see that much of the added complexity to the REST request comes from serializing the input Go structs into JSON data and then back to Go structs for the output. In fact, it’s 50% of the word count!

Further, in REST, since the client isn’t provided with any language-native objects for the API, they usually end up just creating these objects themselves. Using gRPC and Protocol Buffers, where language-native objects are provided for clients, many errors related to dealing with the API are caught by the compiler [9], which is significantly more convenient than looking at error codes of a REST API.

Since the object creation isn’t even part of the word count difference above, consuming gRPC endpoints ends up being significantly simpler and faster to implement compared to REST.

### Debugging and Support

Since JSON objects are in plaintext format, they can be created by hand, which makes them very easy to work with as a developer. And if you encounter a problem, you could visually inspect the JSON objects in your code and figure out what’s wrong. You could even just edit the JSON objects yourself to add or remove properties. This is particularly useful when consuming a new API you haven’t worked with before.

Protocol Buffers over a gRPC request make it much harder to directly see what data is being passed over the wire, since it’s just encoded into binary. However, since the data being sent over is an exact representation of a native language object, we could just use pre-existing language-specific debugging tools to see the state of the objects before a gRPC request is sent. This allows for fairly easy debugging as well.

It definitely helps to be able to see the data that’s being passed over the network using JSON. However, Protocol Buffers’ strong integration into languages provides an almost as easy way to figure out what’s going on with a request.

### Outcomes

While gRPC has a larger learning curve, less support, and is harder to debug directly, its improvements in developer efficiency (especially on the client side), presents a strong advantage. We decided that provided support and documentation for Protocol Buffers in a particular language is strong, we should be able to overcome the debugging problems using language-based debugging tools, and in turn, benefit from the faster development time of gRPC.

------

## Overall Performance

We’ve already compared HTTP/1.1 and HTTP/2. So now let’s take the best form of REST (REST over HTTP/2) and pit it against everything gRPC has to offer.

Conveniently, at this point, we’ve already written all the client code we need. We’ll be comparing the performance of the simple POST request and its gRPC equivalent from the ‘Ease of Use’ section above.

### Server Implementations

`main.go`

```golang
package main

import (
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	
	"github.com/Bimde/grpc-vs-rest/pb"
)

type server struct{}

func main() {
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterRandomServiceServer(s, &server{})
	log.Println("Starting gRPC server")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (s *server) DoSomething(_ context.Context, random *pb.Random) (*pb.Random, error) {
	random.RandomString = "[Updated] " + random.RandomString;
	return random, nil
}
```

`main.go`

```golang
package main

import (
    "log"
    "encoding/json"
    "net/http"
    "github.com/Bimde/grpc-vs-rest/pb"
)

func handle(w http.ResponseWriter, req *http.Request) {
    decoder := json.NewDecoder(req.Body)
    var random pb.Random
    if err := decoder.Decode(&random); err != nil {
        panic(err)
    }
    random.RandomString = "[Updated] " + random.RandomString

    bytes, err := json.Marshal(&random)
    if err != nil {
	panic(err)
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(bytes)
}

func main() {
    server := &http.Server{Addr: "bimde:8080", Handler: http.HandlerFunc(handle)}
    log.Fatal(server.ListenAndServeTLS("server.crt", "server.key"))
}
```

Both servers are fairly simple, implementing the ADT required by the client. Both servers are running locally over HTTP/2.

### Benchmarks

Since we already have a Job/Worker implementation from the HTTP/1.1 vs. HTTP/2 benchmarks, we could reuse that code.

Now, all we need are individual benchmarks. First, the REST benchmark:

`rest_benchmark.go`

```golang
func BenchmarkHTTP2GetWithWokers(b *testing.B) {
	client.Transport = &http2.Transport{
		TLSClientConfig: createTLSConfigWithCustomCert(),
	}
	requestQueue := make(chan Request)
	defer startWorkers(&requestQueue, noWorkers, startPostWorker)()
	b.ResetTimer() // don't count worker initialization time
  
	for i := 0; i < b.N; i++ {
		requestQueue <- Request{
			Path: "https://bimde:8080", 
			Random: &pb.Random{
				RandomInt: 2019, 
				RandomString: "a_string",
			},
		}
	}
}

func startPostWorker(requestQueue *chan Request, wg *sync.WaitGroup) {
	go func() {
		for {
			request := <- *requestQueue
			if (request.Path == stopRequestPath) {
				wg.Done()
				return
			}
			post(request.Path, request.Random, request.Random)
		}
	}()
}
```

Next, the gRPC benchmark:

`grpc_benchmark.go`

```golang
func BenchmarkGRPCWithWokers(b *testing.B) {
	conn, err := grpc.Dial("bimde:9090", grpc.WithInsecure())
	if err != nil {
		  log.Fatalf("Dial failed: %v", err)
	}
	client := pb.NewRandomServiceClient(conn)
	requestQueue := make(chan Request)
	defer startWorkers(&requestQueue, noWorkers, getStartGRPCWorkerFunction(client))()
	b.ResetTimer() // don't count worker initialization time

	for i := 0; i < b.N; i++ {
		requestQueue <- Request{
			Path: "http://localhost:9090", 
			Random: &pb.Random{
				RandomInt: 2019, 
				RandomString: "a_string",
			},
		}
	}
}

func getStartGRPCWorkerFunction(client pb.RandomServiceClient) func (*chan Request, *sync.WaitGroup) {
	return func(requestQueue *chan Request, wg *sync.WaitGroup) {
		go func() {
			for {
				request := <- *requestQueue
				if (request.Path == stopRequestPath) {
					wg.Done()
					return
				}
				client.DoSomething(context.TODO(), request.Random)
			}
		}()
	}
}
```

Notice that the `getStartGRPCWorkerFunction` function returns a closure with the a `RandomServiceClient` in it. This is what allows us to dial the gRPC server only once, i.e. perform only a single TCP handshake for the entirely of a test.

### Results

![Image for post](https://miro.medium.com/max/1200/1*crxDIvpyETiqyJpUOibQZQ.png)

While REST over HTTP/2 scales about as well as gRPC does, in terms of pure performance gRPC brings a reduction in processing time of 50–75% throughout the entire workload range.

------

## Outcomes

Let’s go back to our original criteria:

- Language-neutral
- Fast
- Easy to use

In terms of language support, JSON-backed REST is the clear winner. gRPC’s language support has improved drastically over the last couple of years, however, and it’s arguably sufficient for most use cases.

Our performance comparisons eliminate HTTP/1.1 from all use cases but supporting legacy clients through a front-end API service. Between gRPC and REST over HTTP/2, the performance difference is still significant. Anytime that request performance is a key issue, gRPC seems to be the correct choice.

In terms of ease of use, developers need to write less code to do the same thing in gRPC compared to REST. Debugging is *different*, but not necessarily any harder. It’s more a problem of developers getting used to a new paradigm.

From our findings, we can see that gRPC is a much better solution for internal service to service communication. It has better performance, improves development speed, and is sufficiently language-neutral. We can conclude that we should default to building gRPC services unless REST is needed to support external clients, or to support a language/platform gRPC isn’t built for yet.

### Impacts

- Reduced latency for customers; a better user experience
- Lower processing time for requests; lower costs
- Improved developer efficiency; lower costs for companies and more new features developed

## Source code

<https://github.com/Bimde/grpc-vs-rest>

## Sources

- [1] <https://grpc.io/docs/>
- [2] <https://daniel.haxx.se/http2/>
- [3] <https://hpbn.co/http2/>
- [4] <https://httpwg.org/specs/rfc7540.html>
- [5] <https://developers.google.com/web/fundamentals/performance/http2/#request_and_response_multiplexing>
- [6] <https://posener.github.io/http2/>
- [7] <http://www.rfc-editor.org/rfc/pdfrfc/rfc7541.txt.pdf>
- [8] <https://improbable.io/blog/grpc-web-moving-past-restjson-towards-type-safe-web-apis>
- [9] <http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/>

