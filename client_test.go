package client_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"go.clever-cloud.dev/client"
)

func Test_client_Get(t *testing.T) {
	t.Parallel()

	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	log.SetOutput(os.Stdout)

	clever := client.New(
		client.WithLogger(log),
		client.WithAutoOauthConfig(),
	)

	type args struct {
		client *client.Client
		path   string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{{
		name: "main",
		args: args{
			client: clever,
			path:   "/v4/products/zones", // unauthenticated path
		},
		wantErr: false,
	}, {
		name: "main authenticated",
		args: args{
			client: clever,
			path:   "/v2/self",
		},
		wantErr: false,
	}, {
		name: "not found error",
		args: args{
			client: clever,
			path:   "/vx/test",
		},
		wantErr: true,
	}, {
		name: "unreacheable API",
		args: args{
			client: client.New(
				client.WithLogger(log),
				client.WithEndpoint("https://neverresolve"),
			),
			path: "/vx/test",
		},
		wantErr: true,
	}}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := client.Get[interface{}](context.Background(), tt.args.client, tt.args.path)
			if (res.Error() != nil) != tt.wantErr {
				t.Errorf("client.Get() error = %v, wantErr %v", res.Error().Error(), tt.wantErr)

				return
			}

			errStr := ""
			if res.Error() != nil {
				errStr = res.Error().Error()
			}
			t.Logf("response: code=%d, sozuId=%s, error=%s, payload=%+v", res.StatusCode(), res.SozuID(), errStr, res.Payload())
		})
	}
}

func Test_client_empty(t *testing.T) {
	t.Parallel()

	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	log.SetOutput(os.Stdout)

	c := client.New(
		client.WithLogger(log),
		client.WithAutoOauthConfig(),
	)

	res := client.Get[client.Nothing](context.Background(), c, "/")
	if res.HasError() {
		t.Errorf("expect empty response, err: %s", res.Error().Error())
	}
}

func Test_client_Payload(t *testing.T) {
	t.Parallel()

	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	log.SetOutput(os.Stdout)

	clever := client.New(
		client.WithLogger(log),
		client.WithAutoOauthConfig(),
	)

	type Self struct {
		ID             string        `json:"id"`
		Email          string        `json:"email"`
		Name           string        `json:"name"`
		Phone          string        `json:"phone"`
		Address        string        `json:"address"`
		City           string        `json:"city"`
		ZipCode        string        `json:"zipcode"`
		Country        string        `json:"country"`
		Avatar         *string       `json:"avatar"`
		CreationDate   int64         `json:"creationDate"`
		Lang           *string       `json:"lang"`
		EmailValidated bool          `json:"emailValidated"`
		OauthApps      []interface{} `json:"oauthApps"`
		Admin          bool          `json:"admin"`
		CanPay         bool          `json:"canPay"`
		PreferedMFA    *string       `json:"preferredMFA"`
		HasPassword    bool          `json:"hasPassword"`
	}

	res := client.Get[Self](context.Background(), clever, "/v2/self")
	if res.HasError() {
		t.Errorf("client.Get() error = %v", res.Error().Error())

		return
	}

	t.Logf("response: code=%d, sozuId=%s, type=%T", res.StatusCode(), res.SozuID(), res.Payload())

	if self := res.Payload(); self.ID == "" {
		t.Errorf("self.ID shoud not be empty")

		return
	}
}

func Test_client_Stream(t *testing.T) {
	t.Parallel()

	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	log.SetOutput(os.Stdout)

	clever := client.New(
		client.WithLogger(log),
		client.WithAutoOauthConfig(),
	)

	org := os.Getenv("CC_ORG")
	ng := os.Getenv("CC_NG")
	peer := os.Getenv("CC_NG_PEER")
	url := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups/%s/peers/%s/wireguard/configuration/stream", org, ng, peer)

	type logEntry struct{}

	res := client.Stream[logEntry](context.Background(), clever, url)
	if res.Error() != nil {
		t.Errorf("client.Stream() error = %v", res.Error().Error())

		return
	}

	for i := 0; i < 3; i++ {
		msg := <-res.Payload()
		t.Logf("MSG: %s", msg)

		if res.HasError() {
			t.Errorf("Stream.Payload() error = %v", res.Error().Error())

			return
		}
	}

	res.Close()
	res.Close() // we can call Close several times

	if res.Error() != nil {
		t.Errorf("client.Stream() error = %v", res.Error().Error())

		return
	}
}

func Test_client_StreamContext(t *testing.T) {
	t.Parallel()

	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	log.SetOutput(os.Stdout)

	clever := client.New(
		client.WithLogger(log),
		client.WithAutoOauthConfig(),
	)

	org := os.Getenv("CC_ORG")
	ng := os.Getenv("CC_NG")
	peer := os.Getenv("CC_NG_PEER")
	url := fmt.Sprintf("/v4/networkgroups/organisations/%s/networkgroups/%s/peers/%s/wireguard/configuration/stream", org, ng, peer)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	type logEntry struct{}

	res := client.Stream[logEntry](ctx, clever, url)
	if res.Error() != nil {
		t.Errorf("client.Stream() error = %v", res.Error().Error())

		return
	}

	for i := 0; i < 10; i++ {
		msg, ok := <-res.Payload()
		if !ok {
			t.Log("Payload() is closed")

			break
		}

		t.Logf("MSG: %s", msg)

		if res.HasError() {
			t.Errorf("Stream.Payload() error = %v", res.Error().Error())

			return
		}
	}

	res.Close()

	if res.Error() == nil {
		t.Errorf("client.Stream() expect context error")

		return
	}
}

// Simple Get.
func ExampleGet() {
	cc := client.New(client.WithAutoOauthConfig())

	res := client.Get[map[string]interface{}](context.Background(), cc, "/v2/self")
	if res.HasError() {
		panic(res.Error())
	}

	fmt.Printf("%+v\n", res.Payload())
}

// SSE.
func ExampleStream() {
	type logEntry map[string]interface{}

	cc := client.New(client.WithAutoOauthConfig())

	res := client.Stream[logEntry](context.Background(), cc, "/v4/logs")
	defer res.Close()

	for msg := range res.Payload() {
		fmt.Printf("%+v\n", msg)
	}
}
