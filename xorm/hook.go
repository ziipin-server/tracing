package xorm

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/ziipin-server/zpmysql"
	"go.elastic.co/apm/module/apmsql"
)

type TracerXormHook struct{}

func (h *TracerXormHook) BeforeProcess(c *zpmysql.HookContext) (context.Context, error) {
	if c.Ctx == nil {
		return nil, nil
	}
	span := opentracing.SpanFromContext(c.Ctx)
	if span == nil {
		return c.Ctx, nil
	}
	_, ctx := opentracing.StartSpanFromContext(c.Ctx, apmsql.QuerySignature(c.SQL))
	return ctx, nil
}

func (h *TracerXormHook) AfterProcess(c *zpmysql.HookContext) error {
	if c.Ctx == nil {
		return nil
	}
	span := opentracing.SpanFromContext(c.Ctx)
	if span == nil {
		return nil
	}
	ext.DBStatement.Set(span, fmt.Sprintf("%v", c.SQL))
	if c.Err != nil {
		span.LogFields(log.String("err", c.Err.Error()))
	}
	span.Finish()
	return nil
}
