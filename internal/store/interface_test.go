package store

import (
	"reflect"
	"testing"

	"github.com/xonoxc/scopion/internal/model"
)

func TestStorageInterfaceHasAppendSpan(t *testing.T) {
	var s Storage
	tpe := reflect.TypeOf(&s).Elem()
	method, ok := tpe.MethodByName("AppendSpan")
	if !ok {
		t.Fatal("Storage interface missing AppendSpan method")
	}
	if method.Type.NumIn() != 1 {
		t.Errorf("AppendSpan should have 1 input (span), got %d", method.Type.NumIn())
	}
	if method.Type.In(0) != reflect.TypeOf(model.Span{}) {
		t.Errorf("AppendSpan input should be model.Span, got %v", method.Type.In(0))
	}
	if method.Type.NumOut() != 1 {
		t.Errorf("AppendSpan should have 1 output (error), got %d", method.Type.NumOut())
	}
}

func TestStorageInterfaceHasGetTraceSpans(t *testing.T) {
	var s Storage
	tpe := reflect.TypeOf(&s).Elem()
	method, ok := tpe.MethodByName("GetTraceSpans")
	if !ok {
		t.Fatal("Storage interface missing GetTraceSpans method")
	}
	if method.Type.NumIn() != 1 {
		t.Errorf("GetTraceSpans should have 1 input (traceID), got %d", method.Type.NumIn())
	}
	if method.Type.NumOut() != 2 {
		t.Errorf("GetTraceSpans should have 2 outputs, got %d", method.Type.NumOut())
	}
	if method.Type.Out(0) != reflect.TypeOf([]model.Span{}) {
		t.Errorf("GetTraceSpans first output should be []model.Span, got %v", method.Type.Out(0))
	}
}

func TestStorageInterfaceGetTracesSignature(t *testing.T) {
	var s Storage
	tpe := reflect.TypeOf(&s).Elem()
	method, ok := tpe.MethodByName("GetTraces")
	if !ok {
		t.Fatal("Storage interface missing GetTraces method")
	}
	if method.Type.NumIn() != 1 {
		t.Errorf("GetTraces should have 1 input (limit), got %d", method.Type.NumIn())
	}
	if method.Type.NumOut() != 2 {
		t.Errorf("GetTraces should have 2 outputs, got %d", method.Type.NumOut())
	}
	if method.Type.Out(0) != reflect.TypeOf([]model.TraceInfo{}) {
		t.Errorf("GetTraces first output should be []model.TraceInfo, got %v", method.Type.Out(0))
	}
}
