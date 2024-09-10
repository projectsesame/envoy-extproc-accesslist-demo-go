package main

import (
	"flag"
	"log"
	"strings"

	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	ep "github.com/wrossmorrow/envoy-extproc-sdk-go"
)

type ipPool map[string]struct{}

func (s ipPool) add(k string) {
	s[strings.TrimSpace(k)] = struct{}{}
}

func (s ipPool) contains(k string) bool {
	_, ok := s[k]
	return ok
}

func (s ipPool) empty() bool {
	return len(s) == 0
}

func (s ipPool) keys() []string {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return keys

}
func (s ipPool) String() string {
	return strings.Join(s.keys(), ipSep)
}

func newIPPool(ips []string) ipPool {
	pool := make(ipPool)
	for _, ip := range ips {
		pool.add(ip)
	}
	return pool
}

type allowAndBlockRequestProcessor struct {
	opts  *ep.ProcessingOptions
	allow ipPool
	block ipPool
}

func (s *allowAndBlockRequestProcessor) GetName() string {
	return "allow-and-block"
}

func (s *allowAndBlockRequestProcessor) GetOptions() *ep.ProcessingOptions {
	return s.opts
}

const (
	kXFF = "x-forwarded-for"

	ipSep = ","
)

func extractClientIP(headers ep.AllHeaders) string {
	xff := string(headers.RawHeaders[kXFF])
	if xff == "" {
		return ""

	}

	ips := strings.Split(xff, ipSep)
	return strings.TrimSpace(ips[0])

}

func (s *allowAndBlockRequestProcessor) ProcessRequestHeaders(ctx *ep.RequestContext, headers ep.AllHeaders) error {
	cancel := func(code int32) error {
		return ctx.CancelRequest(code, map[string]ep.HeaderValue{}, typev3.StatusCode_name[code])
	}
	ip := extractClientIP(headers)
	if ip == "" {
		log.Printf("no xff")
		return cancel(403)
	}

	// white list
	if !s.allow.empty() && !s.allow.contains(ip) {
		log.Printf("the ip: %s is not in the allow list\n", ip)
		return cancel(403)

	}

	// black list
	if !s.block.empty() && s.block.contains(ip) {
		log.Printf("the ip: %s is in the block list\n", ip)
		return cancel(403)
	}

	return ctx.ContinueRequest()
}

func (s *allowAndBlockRequestProcessor) ProcessRequestBody(ctx *ep.RequestContext, body []byte) error {
	return ctx.ContinueRequest()
}

func (s *allowAndBlockRequestProcessor) ProcessRequestTrailers(ctx *ep.RequestContext, trailers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (s *allowAndBlockRequestProcessor) ProcessResponseHeaders(ctx *ep.RequestContext, headers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (s *allowAndBlockRequestProcessor) ProcessResponseBody(ctx *ep.RequestContext, body []byte) error {
	return ctx.ContinueRequest()
}

func (s *allowAndBlockRequestProcessor) ProcessResponseTrailers(ctx *ep.RequestContext, trailers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (ap *allowAndBlockRequestProcessor) parseCmdArgs(args []string) (allow, block []string, err error) {
	subCmd := flag.NewFlagSet("subcmd", flag.ExitOnError)
	subCmd.Func(kAllowList, "", func(s string) error {
		allow = strings.Split(s, ipSep)
		return nil
	})
	subCmd.Func(kBlockList, "", func(s string) error {
		block = strings.Split(s, ipSep)
		return nil
	})
	err = subCmd.Parse(args)
	return

}

func (s *allowAndBlockRequestProcessor) Init(opts *ep.ProcessingOptions, args []string) error {
	allow, block, err := s.parseCmdArgs(args)
	if err != nil {
		return err
	}

	s.opts = opts
	s.allow = newIPPool(allow)
	s.block = newIPPool(block)

	return nil
}

func (s *allowAndBlockRequestProcessor) Finish() {}
