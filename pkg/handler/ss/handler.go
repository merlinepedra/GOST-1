package ss

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/go-gost/gosocks5"
	"github.com/go-gost/gost/pkg/bypass"
	"github.com/go-gost/gost/pkg/chain"
	"github.com/go-gost/gost/pkg/common/util/ss"
	"github.com/go-gost/gost/pkg/handler"
	"github.com/go-gost/gost/pkg/logger"
	md "github.com/go-gost/gost/pkg/metadata"
	"github.com/go-gost/gost/pkg/registry"
)

func init() {
	registry.RegisterHandler("ss", NewHandler)
}

type ssHandler struct {
	bypass bypass.Bypass
	router *chain.Router
	logger logger.Logger
	md     metadata
}

func NewHandler(opts ...handler.Option) handler.Handler {
	options := &handler.Options{}
	for _, opt := range opts {
		opt(options)
	}

	return &ssHandler{
		bypass: options.Bypass,
		router: (&chain.Router{}).
			WithLogger(options.Logger).
			WithResolver(options.Resolver),
		logger: options.Logger,
	}
}

func (h *ssHandler) Init(md md.Metadata) (err error) {
	if err := h.parseMetadata(md); err != nil {
		return err
	}

	h.router.WithRetry(h.md.retryCount)

	return nil
}

// implements chain.Chainable interface
func (h *ssHandler) WithChain(chain *chain.Chain) {
	h.router.WithChain(chain)
}

func (h *ssHandler) Handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	start := time.Now()
	h.logger = h.logger.WithFields(map[string]interface{}{
		"remote": conn.RemoteAddr().String(),
		"local":  conn.LocalAddr().String(),
	})

	h.logger.Infof("%s <> %s", conn.RemoteAddr(), conn.LocalAddr())
	defer func() {
		h.logger.WithFields(map[string]interface{}{
			"duration": time.Since(start),
		}).Infof("%s >< %s", conn.RemoteAddr(), conn.LocalAddr())
	}()

	if h.md.cipher != nil {
		conn = ss.ShadowConn(h.md.cipher.StreamConn(conn), nil)
	}

	if h.md.readTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(h.md.readTimeout))
	}

	addr := &gosocks5.Addr{}
	if _, err := addr.ReadFrom(conn); err != nil {
		h.logger.Error(err)
		io.Copy(ioutil.Discard, conn)
		return
	}

	h.logger = h.logger.WithFields(map[string]interface{}{
		"dst": addr.String(),
	})

	h.logger.Infof("%s >> %s", conn.RemoteAddr(), addr)

	if h.bypass != nil && h.bypass.Contains(addr.String()) {
		h.logger.Info("bypass: ", addr.String())
		return
	}

	cc, err := h.router.Dial(ctx, "tcp", addr.String())
	if err != nil {
		return
	}
	defer cc.Close()

	t := time.Now()
	h.logger.Infof("%s <-> %s", conn.RemoteAddr(), addr)
	handler.Transport(conn, cc)
	h.logger.
		WithFields(map[string]interface{}{
			"duration": time.Since(t),
		}).
		Infof("%s >-< %s", conn.RemoteAddr(), addr)
}
