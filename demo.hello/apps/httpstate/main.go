package main

import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"path"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

// refer: https://github.com/davecheney/httpstat

const (
	httpsTemplate = `
         DNS Lookup   TCP Connection   TLS Handshake   Server Processing   Content Transfer
[%s  |     %s  |    %s  |        %s  |       %s  ]
            |                |               |                   |                  |
   namelookup:%s      |               |                   |                  |
                       connect:%s     |                   |                  |
                                   pretransfer:%s         |                  |
                                                     starttransfer:%s        |
                                                                             total:%s
`

	httpTemplate = `
         DNS Lookup   TCP Connection   Server Processing   Content Transfer
[ %s  |     %s  |        %s  |       %s  ]
             |                |                   |                  |
    namelookup:%s      |                   |                  |
                        connect:%s         |                  |
                                      starttransfer:%s        |
                                                              total:%s
`
)

var (
	// command line flags
	httpMethod      string
	postBody        string
	followRedirects bool
	onlyHeader      bool
	insecure        bool
	httpHeaders     headers
	saveOutput      bool
	outputFile      string
	showVersion     bool
	clientCertFile  string
	fourOnly        bool
	sixOnly         bool

	// number of redirects followed
	redirectsFollowed int
	// for -v flag, updated during the release process with -ldflags=-X=main.version=...
	version = "devel"
)

const maxRedirects = 10

func init() {
	flag.StringVar(&httpMethod, "X", "GET", "HTTP method to use")
	flag.StringVar(&postBody, "d", "", "the body of a POST or PUT request; from file use @filename")
	flag.BoolVar(&followRedirects, "L", false, "follow 30x redirects")
	flag.BoolVar(&onlyHeader, "I", false, "don't read body of request")
	flag.BoolVar(&insecure, "k", false, "allow insecure SSL connections")
	flag.Var(&httpHeaders, "H", "set HTTP header; repeatable: -H 'Accept: ...' -H 'Range: ...'")
	flag.BoolVar(&saveOutput, "O", false, "save body as remote filename")
	flag.StringVar(&outputFile, "o", "", "output file for body")
	flag.BoolVar(&showVersion, "v", false, "print version number")
	flag.StringVar(&clientCertFile, "E", "", "client cert file for tls config")
	flag.BoolVar(&fourOnly, "4", false, "resolve IPv4 addresses only")
	flag.BoolVar(&sixOnly, "6", false, "resolve IPv6 addresses only")
}

/*
http headers
*/

type headers []string

func (h headers) String() string {
	tmpSlice := make([]string, 0, len(h))
	for _, v := range h {
		tmpSlice = append(tmpSlice, "-H "+v)
	}
	return strings.Join(tmpSlice, ",")
}

func (h headers) Set(v string) error {
	h = append(h, v)
	return nil
}

func (h headers) Len() int { return len(h) }

func (h headers) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h headers) Less(i, j int) bool {
	a, b := h[i], h[j]
	if a == "Server" {
		return true
	}
	if b == "Server" {
		return false
	}

	lessHeader := func(header string) bool {
		switch header {
		case "Connection",
			"Keep-Alive",
			"Proxy-Authenticate",
			"Proxy-Authorization",
			"TE",
			"Trailers",
			"Transfer-Encoding",
			"Upgrade":
			return false
		default:
			return true
		}
	}
	x, y := lessHeader(a), lessHeader(b)
	if x == y {
		return a < b
	}
	return x
}

/*
main
*/

func main() {
	flag.Parse()

	if showVersion {
		fmt.Printf("%s %s (runtime: %s)\n", os.Args[0], version, runtime.Version())
		os.Exit(0)
	}

	if fourOnly && sixOnly {
		fmt.Fprintf(os.Stderr, "%s: Only one of -4 and -6 may be specified\n", os.Args[0])
		os.Exit(-1)
	}

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		os.Exit(2)
	}

	if (httpMethod == "POST" || httpMethod == "PUT") && postBody == "" {
		log.Fatal("must supply post body using -d when POST or PUT is used")
	}

	if onlyHeader {
		httpMethod = "HEAD"
	}

	uri := args[0]
	url := parseURL(uri)
	visit(url)
}

func parseURL(uri string) *url.URL {
	if !strings.Contains(uri, "://") && !strings.HasPrefix(uri, "//") {
		uri = "//" + uri
	}

	retURL, err := url.Parse(uri)
	if err != nil {
		log.Fatalf("could not parse url %q: %v", uri, err)
	}

	if retURL.Scheme == "" {
		retURL.Scheme = "http"
		if !strings.HasSuffix(retURL.Host, ":80") {
			retURL.Scheme += "s"
		}
	}
	return retURL
}

// visit visits a url and times the interaction.
// If the response is a 30x, visit follows the redirect.
func visit(url *url.URL) {
	// dns lookup:     t1 - t0
	// tcp connect:    t2 - t1
	// tls handshake:  t6 - t5
	// server process: t4 - t3
	var t0, t1, t2, t3, t4, t5, t6 time.Time

	// request trace
	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { t0 = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { t1 = time.Now() },
		ConnectStart: func(_, _ string) {
			if t1.IsZero() {
				t1 = time.Now() // connecting to IP
			}
		},
		ConnectDone: func(net, addr string, err error) {
			if err != nil {
				log.Fatalf("unable to connect to host %v: %v", addr, err)
			}
			t2 = time.Now()
			printf("\n%s%s\n", color.GreenString("Connected to "), color.CyanString(addr))
		},
		GotConn:              func(_ httptrace.GotConnInfo) { t3 = time.Now() },
		GotFirstResponseByte: func() { t4 = time.Now() },
		TLSHandshakeStart:    func() { t5 = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { t6 = time.Now() },
	}

	req := newRequest(httpMethod, url, postBody)
	req = req.WithContext(httptrace.WithClientTrace(context.Background(), trace))

	// http transport
	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
		ForceAttemptHTTP2:     true,
	}
	switch {
	case fourOnly:
		tr.DialContext = dialContext("tcp4")
	case sixOnly:
		tr.DialContext = dialContext("tcp6")
	}

	if url.Scheme == "https" {
		host, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			host = req.Host
		}
		tr.TLSClientConfig = &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: insecure,
			Certificates:       readClientCert(clientCertFile),
		}
	}

	// http client
	client := &http.Client{
		Transport: tr,
		// always refuse to follow redirects, visit does that manually if required.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to read response: %v", err)
	}

	bodyMsg := readResponseBody(req, resp)
	resp.Body.Close()

	// after read body
	// content transfer: t7 - t4
	t7 := time.Now()

	// output
	if t0.IsZero() {
		// we skipped DNS
		t0 = t1
	}

	// status line
	printf("\n%s%s%s\n", color.GreenString("HTTP"), grayscale(14)("/"),
		color.CyanString("%d.%d %s", resp.ProtoMajor, resp.ProtoMinor, resp.Status))

	// headers
	names := make([]string, 0, len(resp.Header))
	for val := range resp.Header {
		names = append(names, val)
	}
	sort.Sort(headers(names))
	for _, name := range names {
		printf("%s %s\n", grayscale(14)(name+":"), color.CyanString(strings.Join(resp.Header[name], ",")))
	}

	if bodyMsg != "" {
		printf("\n%s\n", bodyMsg)
	}

	fmta := func(d time.Duration) string {
		return color.CyanString("%7dms", int(d/time.Millisecond))
	}
	fmtb := func(d time.Duration) string {
		return color.CyanString("%-9s", strconv.Itoa(int(d/time.Millisecond))+"ms")
	}
	colorize := func(s string) string {
		lines := strings.Split(s, "\n")
		lines[0] = grayscale(16)(lines[0])
		return strings.Join(lines, "\n")
	}

	// times
	fmt.Println()
	switch url.Scheme {
	case "https":
		printf(colorize(httpsTemplate),
			fmta(t1.Sub(t0)), // dns lookup
			fmta(t2.Sub(t1)), // tcp connection
			fmta(t6.Sub(t5)), // tls handshake
			fmta(t4.Sub(t3)), // server processing
			fmta(t7.Sub(t4)), // content transfer
			fmtb(t1.Sub(t0)), // namelookup
			fmtb(t2.Sub(t0)), // connect
			fmtb(t3.Sub(t0)), // pretransfer
			fmtb(t4.Sub(t0)), // starttransfer
			fmtb(t7.Sub(t0)), // total
		)
	case "http":
		printf(colorize(httpTemplate),
			fmta(t1.Sub(t0)), // dns lookup
			fmta(t3.Sub(t1)), // tcp connection
			fmta(t4.Sub(t3)), // server processing
			fmta(t7.Sub(t4)), // content transfer
			fmtb(t1.Sub(t0)), // namelookup
			fmtb(t3.Sub(t0)), // connect
			fmtb(t4.Sub(t0)), // starttransfer
			fmtb(t7.Sub(t0)), // total
		)
	}

	// redirect
	if followRedirects && isRedirect(resp) {
		loc, err := resp.Location()
		if err != nil {
			if err == http.ErrNoLocation {
				// 30x but no Location to follow, give up.
				return
			}
			log.Fatalf("unable to follow redirect: %v", err)
		}

		redirectsFollowed++
		if redirectsFollowed > maxRedirects {
			log.Fatalf("maximum number of redirects (%d) followed", maxRedirects)
		}
		visit(loc)
	}
}

func newRequest(method string, url *url.URL, body string) *http.Request {
	req, err := http.NewRequest(method, url.String(), createBody(body))
	if err != nil {
		log.Fatalf("unable to create request: %v", err)
	}

	for _, h := range httpHeaders {
		k, v := headerKeyValue(h)
		if strings.EqualFold(k, "host") {
			req.Host = v
			continue
		}
		req.Header.Add(k, v)
	}
	return req
}

func createBody(body string) io.Reader {
	if strings.HasPrefix(body, "@") {
		filename := body[1:]
		f, err := os.Open(filename)
		if err != nil {
			log.Fatalf("failed to open data file %s: %v", filename, err)
		}
		return f
	}
	return strings.NewReader(body)
}

func headerKeyValue(h string) (string, string) {
	idx := strings.Index(h, ":")
	if idx == -1 {
		log.Fatalf("Header '%s' has invalid format, missing ':'", h)
	}
	return strings.TrimRight(h[:idx], " "), strings.TrimLeft(h[idx:], " :")
}

func dialContext(network string) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: false,
		}).DialContext(ctx, network, addr)
	}
}

// readClientCert: helper function to read client certificate from pem formatted file.
func readClientCert(filename string) []tls.Certificate {
	if filename == "" {
		return nil
	}

	var (
		pkeyPem []byte
		certPem []byte
	)

	readBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to read client certificate file: %v", err)
	}
	for {
		block, rest := pem.Decode(readBytes)
		if block == nil {
			break
		}
		readBytes = rest

		if strings.HasPrefix(block.Type, "PRIVATE KEY") {
			pkeyPem = pem.EncodeToMemory(block)
		}
		if strings.HasSuffix(block.Type, "CERTIFICATE") {
			certPem = pem.EncodeToMemory(block)
		}
	}

	cert, err := tls.X509KeyPair(certPem, pkeyPem)
	if err != nil {
		log.Fatalf("unable to load client cert and key pair: %v", err)
	}
	return []tls.Certificate{cert}
}

func readResponseBody(req *http.Request, resp *http.Response) string {
	if isRedirect(resp) || req.Method == http.MethodHead {
		return ""
	}

	writer := ioutil.Discard
	msg := color.CyanString("Body discarded")

	if saveOutput || outputFile != "" {
		filename := outputFile
		if saveOutput {
			if filename = getFilenameFromHeaders(resp.Header); filename == "" {
				filename = path.Base(req.URL.RequestURI())
			}
			if filename == "/" {
				log.Fatalf("No remote filename; specify output filename with -o to save response body")
			}
		}

		f, err := os.Create(filename)
		if err != nil {
			log.Fatalf("unable to create file %s: %v", filename, err)
		}
		defer f.Close()
		writer = f
		msg = color.CyanString("Body read")
	}

	if _, err := io.Copy(writer, resp.Body); err != nil && writer != ioutil.Discard {
		log.Fatalf("failed to read response body: %v", err)
	}
	return msg
}

func isRedirect(resp *http.Response) bool {
	return resp.StatusCode > 299 && resp.StatusCode < 400
}

func getFilenameFromHeaders(headers http.Header) string {
	if hdr := headers.Get("Content-Disposition"); hdr != "" {
		mt, params, err := mime.ParseMediaType(hdr)
		if err == nil && mt == "attachment" {
			if filename := params["filename"]; filename != "" {
				return filename
			}
		}
	}
	return ""
}

/*
print
*/

func printf(format string, args ...interface{}) {
	fmt.Fprintf(color.Output, format, args...)
}

func grayscale(code color.Attribute) func(string, ...interface{}) string {
	return color.New(code + 232).SprintfFunc()
}
