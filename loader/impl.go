package loader

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	resource "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"me/bldrec"

	capnp "capnproto.org/go/capnp/v3"
	zmq "github.com/pebbe/zmq4"
)

var glblContext context.Context
var glblMeterProvider *sdkmetric.MeterProvider
var glblInstruments map[string]any

func createDebitView() sdkmetric.View {
	// ---- provide view instead of instrument --------------------------------
	debitView := sdkmetric.NewView(
		sdkmetric.Instrument{
			Kind:        sdkmetric.InstrumentKindHistogram,
			Name:        "debit-histogram",
			Unit:        "ron",
			Description: "bank account debit in currency RON",
		},
		sdkmetric.Stream{
			Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: []float64{10, 50, 100, 500, 1000, 5000, 10000},
			},
		},
	)
	return debitView
}

func createResource() *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("spending-loader"),
		semconv.ServiceVersion("0.1.0"),
	)

}

func CreateMetricsPipeline(ctx context.Context) error {

	// ---- OTLP gRPC exporter ------------------------------------------------
	grpcMetricExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return err
	}
	// ---- OTLP console  (debug) exporter ------------------------------------------------
	consoleMetricExporter, err := stdoutmetric.New(stdoutmetric.WithWriter(os.Stdout),
		stdoutmetric.WithPrettyPrint())
	if err != nil {
		return err
	}

	glblMeterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(grpcMetricExporter)),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(consoleMetricExporter)),
		sdkmetric.WithResource(createResource()),
		//sdkmetric.WithView(createDebitView()),
	)
	otel.SetMeterProvider(glblMeterProvider)
	glblContext = ctx

	return nil
}

func CreateDebitInstrument() error {
	// ---- Create the histogram instrument ------------------------------------
	meter := otel.Meter("metrics_exercise")
	hist, err := meter.Float64Histogram(
		"debitHistogram",
		metric.WithUnit("ron"),
		metric.WithDescription("bank account debit in currency RON"),
	)
	if err != nil {
		log.Fatalf("failed to create histogram: %v", err)
	}
	glblInstruments = make(map[string]any)
	glblInstruments["debit-histogram"] = hist

	return nil
}

func ShutdownMetric(ctx context.Context) {
	if err := glblMeterProvider.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown MeterProvider: %s", err)
	}
}

func StartLoadBalancer(config map[string]string) {
	frontend, err := zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		log.Fatalf("failed to create frontend socket: %v", err)
	}
	defer frontend.Close()

	backend, err := zmq.NewSocket(zmq.DEALER)
	if err != nil {
		log.Fatalf("failed to create backend socket: %v", err)
	}
	defer backend.Close()

	port := config["frontend_port"]
	if err := frontend.Bind(fmt.Sprintf("tcp://localhost:%s", port)); err != nil {
		log.Fatalf("failed to bind frontend: %v", err)
	}

	if err := backend.Bind("tcp://localhost:5556"); err != nil {
		log.Fatalf("failed to bind backend: %v", err)
	}

	for i := 0; i < 5; i++ {
		go startWorker(i)
	}

	if err := zmq.Proxy(frontend, backend, nil); err != nil {
		log.Fatalf("proxy error: %v", err)
	}
}

func startWorker(id int) {
	socket, _ := zmq.NewSocket(zmq.REP)
	defer socket.Close()
	socket.Connect("tcp://localhost:5556")

	for {
		zmqMsgBytes, err := socket.RecvBytes(0)
		if err != nil {
			log.Printf("Worker %d: error receiving message: %v", id, err)
			continue
		}
		// Wrap in a Cap’n Proto message (read‑only)
		msg, err := capnp.Unmarshal(zmqMsgBytes)
		if err != nil {
			log.Printf("Worker %d: capnp message error: %v", id, err)
			continue
		}
		record, err := bldrec.ReadRootRecord(msg)
		if err != nil {
			log.Printf("Worker %d: read struct error: %v", id, err)
			continue
		}
		desc, _ := record.SDescription()
		fmt.Printf("Worker %d received: %s\n", id, desc)

		if hist, ok := glblInstruments["debit-histogram"].(metric.Float64Histogram); ok {
			hist.Record(glblContext, float64(record.FValue()))
		}
		socket.Send(fmt.Sprintf("Reply from worker %d", id), 0)
	}
}
