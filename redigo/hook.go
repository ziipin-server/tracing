package redigo

import (
	"context"
	"fmt"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type Conn interface {
	redis.Conn

	// WithContext returns a shallow copy of the connection with
	// its context changed to ctx.
	//
	// To report commands as spans, ctx must contain a transaction or span.
	WithContext(ctx context.Context) Conn
}

type contextConn struct {
	redis.Conn
	ctx context.Context
}

func Wrap(conn redis.Conn) Conn {
	return &contextConn{conn, context.Background()}
}

func (c *contextConn) WithContext(ctx context.Context) Conn {
	c.ctx = ctx
	return c
}

func (c *contextConn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	if span := opentracing.SpanFromContext(c.ctx); span == nil {
		return c.Conn.Do(commandName, args...)
	}

	spanName := strings.ToUpper(commandName)
	span, _ := opentracing.StartSpanFromContext(c.ctx, spanName)
	ext.DBType.Set(span, "redis")
	ext.DBStatement.Set(span, fmt.Sprintf("%v", args))
	defer span.Finish()

	return c.Conn.Do(commandName, args...)
}
