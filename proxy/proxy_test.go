package proxy

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/Meduzz/helper/nuts"
	"github.com/Meduzz/rpc"

	"github.com/Meduzz/helper/http/client"

	"github.com/Meduzz/rpc/api"
	"github.com/Meduzz/rpc/proxy/hub"
)

type Hello struct {
	Message string `json:"message"`
}

func TestMain(m *testing.M) {
	conn, err := nuts.Connect()

	if err != nil {
		panic(err)
	}

	rpc := rpc.NewRpc(conn)
	rpc.Handler("params", "", paramsHandler)
	rpc.Handler("error", "", errorHandler)
	rpc.Handler("noop", "", noopHandler)

	kidding := "127.0.0.1"

	pxy := NewProxy(nil, rpc)
	errorHub := pxy.Add(nil, "GET", "/error")
	noopHub := pxy.Add(nil, "GET", "/noop")
	missingHub := pxy.Add(nil, "GET", "/missing")
	deadHub := pxy.Add(nil, "GET", "/dead")
	authHub := pxy.Add(nil, "POST", "/auth")
	paramsHub := pxy.Add(nil, "POST", "/hello/:word")
	noobHub := pxy.Add(&kidding, "GET", "/noop")
	triggerHub := pxy.Add(nil, "GET", "/trigger")
	temporaryHub := pxy.Add(nil, "GET", "/temporary")

	paramsHub.SetRoute(func(req *http.Request, params map[string]string) *hub.Route {
		return &hub.Route{true, "params", true, 3}
	})

	errorHub.SetRoute(func(req *http.Request, params map[string]string) *hub.Route {
		return &hub.Route{true, "error", true, 3}
	})

	noopHub.SetRoute(func(req *http.Request, params map[string]string) *hub.Route {
		return &hub.Route{true, "noop", true, 3}
	})

	missingHub.SetRoute(func(req *http.Request, params map[string]string) *hub.Route {
		return &hub.Route{true, "", true, 3}
	})

	deadHub.SetRoute(func(req *http.Request, params map[string]string) *hub.Route {
		return &hub.Route{false, "", true, 3}
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
		return &hub.Route{true, "params", true, 3}
	})

	noobHub.SetRoute(func(req *http.Request, params map[string]string) *hub.Route {
		return &hub.Route{true, "noop", true, 3}
	})

	triggerHub.SetRoute(func(req *http.Request, params map[string]string) *hub.Route {
		return &hub.Route{true, "noop", false, 3}
	})

	temporaryHub.SetRoute(func(req *http.Request, params map[string]string) *hub.Route {
		return &hub.Route{false, "asdf", true, 3}
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
	req, err := client.POST("http://localhost:4000/auth", &Hello{"%s world"})

	if err != nil {
		fmt.Printf("Found an error createing the request: %s", err.Error())
		t.Fail()
	}

	req.Header("word", "Hello")

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
	req, err = client.POST("http://localhost:4000/auth", &Hello{"%s world"})

	if err != nil {
		fmt.Printf("Found an error createing the request: %s", err.Error())
		t.Fail()
	}

	req.Header("word", "Hello")

	req.Request().SetBasicAuth("asdf", "qwerty")

	res, err = req.Do(http.DefaultClient)

	if err != nil {
		fmt.Printf("Found an error reading the response: %s", err.Error())
		t.Fail()
	}

	if res.Code() != 200 {
		fmt.Printf("Response code was not 200 but %d.", res.Code())
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

func Test_TriggerHappy(t *testing.T) {
	req, err := client.GET("http://localhost:4000/trigger")

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
}

func Test_TemporarilyUnavailable(t *testing.T) {
	req, err := client.GET("http://localhost:4000/temporary")

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

	text := &Hello{}
	msg.Json(text)

	b := api.Builder()
	b.Header("result", "success")
	text.Message = fmt.Sprintf(text.Message, msg.Metadata["word"])
	b.Json(text)

	err := ctx.Reply(b.Message())

	if err != nil {
		fmt.Println(err)
	}
}

func errorHandler(ctx api.Context) {
	reply := api.NewErrorMessage("Error!")
	ctx.Reply(reply)
}

func noopHandler(ctx api.Context) {}
