package proxy

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/Meduzz/helper/http/client"

	"github.com/Meduzz/rpc/api"
	"github.com/Meduzz/rpc/proxy/hub"
	"github.com/Meduzz/rpc/transports"
)

type Hello struct {
	Message string `json:"message"`
}

func TestMain(m *testing.M) {
	server, _ := transports.NewLocalRpcServer("testing")
	server.RegisterHandler("params", paramsHandler)
	server.RegisterHandler("error", errorHandler)
	server.RegisterHandler("noop", noopHandler)
	client := transports.NewLocalRpcClient(server.(*transports.LocalRpcServer))

	kidding := "127.0.0.1"

	pxy := NewProxy(nil, client)
	errorHub := pxy.Add(nil, "GET", "/error")
	noopHub := pxy.Add(nil, "GET", "/noop")
	missingHub := pxy.Add(nil, "GET", "/missing")
	deadHub := pxy.Add(nil, "GET", "/dead")
	authHub := pxy.Add(nil, "POST", "/auth")
	paramsHub := pxy.Add(nil, "POST", "/hello/:word")
	noobHub := pxy.Add(&kidding, "GET", "/noop")

	paramsHub.SetRoute(func(req *http.Request, params map[string]string) *hub.Route {
		return &hub.Route{true, "params", true}
	})

	errorHub.SetRoute(func(req *http.Request, params map[string]string) *hub.Route {
		return &hub.Route{true, "error", true}
	})

	noopHub.SetRoute(func(req *http.Request, params map[string]string) *hub.Route {
		return &hub.Route{true, "noop", true}
	})

	missingHub.SetRoute(func(req *http.Request, params map[string]string) *hub.Route {
		return &hub.Route{true, "", true}
	})

	deadHub.SetRoute(func(req *http.Request, params map[string]string) *hub.Route {
		return &hub.Route{false, "", true}
	})

	authHub.SetFilter(func(req *http.Request) (*http.Request, error) {
		_, _, ok := req.BasicAuth()

		if !ok {
			return nil, fmt.Errorf("No authentication present")
		}

		return req, nil
	}, func(err error) int {
		return 401
	})

	authHub.SetRoute(func(req *http.Request, params map[string]string) *hub.Route {
		return &hub.Route{true, "params", true}
	})

	noobHub.SetRoute(func(req *http.Request, params map[string]string) *hub.Route {
		return &hub.Route{true, "noop", true}
	})

	go pxy.Start(":4000")

	os.Exit(m.Run())
}

func Test_NoResponseStartsNoFires(t *testing.T) {
	req, err := client.GET("http://localhost:4000/noop")

	if err != nil {
		fmt.Printf("Found an error createing the request: %s", err.Error())
		t.Fail()
	}

	res, err := req.Do(http.DefaultClient)

	if err != nil {
		fmt.Printf("Found an error reading the response: %s", err.Error())
		t.Fail()
	}

	if res.Code() != 503 {
		fmt.Printf("Code was not 503, but %d.", res.Code())
		t.Fail()
	}

	if err != nil {
		fmt.Printf("Reading body of response threw error: %s.", err.Error())
		t.Fail()
	}
}

func Test_DeadFunc(t *testing.T) {
	req, err := client.GET("http://localhost:4000/dead")

	if err != nil {
		fmt.Printf("Found an error createing the request: %s", err.Error())
		t.Fail()
	}

	res, err := req.Do(http.DefaultClient)

	if err != nil {
		fmt.Printf("Found an error reading the response: %s", err.Error())
		t.Fail()
	}

	if res.Code() != 404 {
		fmt.Printf("Response code was not 404 but %d.", res.Code())
		t.Fail()
	}
}

func Test_AuthFilter(t *testing.T) {
	req, err := client.POST("http://localhost:4000/auth", &Hello{"world"})

	if err != nil {
		fmt.Printf("Found an error createing the request: %s", err.Error())
		t.Fail()
	}

	res, err := req.Do(http.DefaultClient)

	if err != nil {
		fmt.Printf("Found an error reading the response: %s", err.Error())
		t.Fail()
	}

	if res.Code() != 401 {
		fmt.Printf("Response code was not 401 but %d.", res.Code())
		t.Fail()
	}

	// now with basic auth set.
	req, err = client.POST("http://localhost:4000/auth", &Hello{"world"})

	if err != nil {
		fmt.Printf("Found an error createing the request: %s", err.Error())
		t.Fail()
	}

	req.Request().SetBasicAuth("asdf", "qwerty")

	res, err = req.Do(http.DefaultClient)

	if err != nil {
		fmt.Printf("Found an error reading the response: %s", err.Error())
		t.Fail()
	}

	if res.Code() != 200 {
		fmt.Printf("Response code was not 401 but %d.", res.Code())
		t.Fail()
	}
}

func Test_MissingFunc(t *testing.T) {
	req, err := client.GET("http://localhost:4000/missing")

	if err != nil {
		fmt.Printf("Found an error createing the request: %s", err.Error())
		t.Fail()
	}

	res, err := req.Do(http.DefaultClient)

	if err != nil {
		fmt.Printf("Found an error reading the response: %s", err.Error())
		t.Fail()
	}

	if res.Code() != 404 {
		fmt.Printf("Response code was not 404 but %d.", res.Code())
		t.Fail()
	}
}

func Test_Params(t *testing.T) {
	req, err := client.POST("http://localhost:4000/hello/world", &Hello{"Hello %s!"})

	if err != nil {
		fmt.Printf("Found an error createing the request: %s", err.Error())
		t.Fail()
	}

	res, err := req.Do(http.DefaultClient)

	if err != nil {
		fmt.Printf("Found an error reading the response: %s", err.Error())
		t.Fail()
	}

	if res.Code() != 200 {
		fmt.Printf("Response code was not 200 but %d.", res.Code())
		t.Fail()
	}

	struckt := &Hello{}
	err = res.Body(struckt)

	if err != nil {
		fmt.Printf("Digging out the Hello from response threw error: %s.", err.Error())
		t.Fail()
	}

	if struckt.Message != "Hello world!" {
		fmt.Printf("The response was not Hello world! but: %s.", struckt.Message)
		t.Fail()
	}
}

func Test_UnknownHosts(t *testing.T) {
	req, err := client.GET("http://127.0.0.1:4000/noop")

	if err != nil {
		fmt.Printf("Found an error createing the request: %s", err.Error())
		t.Fail()
	}

	res, err := req.Do(http.DefaultClient)

	if err != nil {
		fmt.Printf("Found an error reading the response: %s", err.Error())
		t.Fail()
	}

	if res.Code() != 503 {
		fmt.Printf("Response code was not 503 but %d.", res.Code())
		t.Fail()
	}
}

func paramsHandler(ctx api.Context) {
	msg, _ := ctx.Body()

	// mother of all nullpointers...
	reply := api.NewBytesMessage([]byte(fmt.Sprintf(string(msg.Body), msg.Metadata["word"])))

	ctx.Reply(reply)
}

func errorHandler(ctx api.Context) {
	reply := api.NewErrorMessage("Error!")
	ctx.Reply(reply)
}

func noopHandler(ctx api.Context) {}
