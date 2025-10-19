package lucirpc

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLuciRPC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Luci RPC Suite")
	defer GinkgoRecover()
}

var _ = Describe("Luci RPC", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	Context("auth", func() {
		It("should be login", func() {
			mux := http.NewServeMux()
			ts := httptest.NewServer(mux)
			defer ts.Close()
			u, err := url.Parse(ts.URL)
			Expect(err).To(BeNil())

			client, err := New("http://"+u.Host, "admin", "password", 1, true)
			Expect(err).To(BeNil())

			mux.HandleFunc(authPath, func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal(http.MethodPost))
				Expect(r.URL.Path).To(Equal(authPath))
				w.WriteHeader(http.StatusAccepted)
				_, err = w.Write([]byte(`{"result":"foobar"}`))
				Expect(err).To(BeNil())
			})

			err = client.auth(ctx)
			Expect(err).To(BeNil())
			Expect(client.token).To(Equal("foobar"))
		})

		It("should be unauthorized", func() {
			mux := http.NewServeMux()
			ts := httptest.NewServer(mux)
			defer ts.Close()
			u, err := url.Parse(ts.URL)
			Expect(err).To(BeNil())

			client, err := New("http://"+u.Host, "admin", "password", 1, true)
			Expect(err).To(BeNil())

			mux.HandleFunc(authPath, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			})

			err = client.auth(ctx)
			Expect(err).To(Equal(ErrHttpUnauthorized))
			Expect(client.token).To(Equal(""))
		})

		It("should be forbidden", func() {
			mux := http.NewServeMux()
			ts := httptest.NewServer(mux)
			defer ts.Close()
			u, err := url.Parse(ts.URL)
			Expect(err).To(BeNil())
			client, err := New("http://"+u.Host, "admin", "password", 1, true)
			Expect(err).To(BeNil())

			mux.HandleFunc(authPath, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
			})

			err = client.auth(ctx)
			Expect(err).To(Equal(ErrHttpForbidden))
			Expect(client.token).To(Equal(""))
		})

		It("should fail", func() {
			mux := http.NewServeMux()
			ts := httptest.NewServer(mux)
			defer ts.Close()
			u, err := url.Parse(ts.URL)
			Expect(err).To(BeNil())
			client, err := New("http://"+u.Host, "admin", "password", 1, true)
			Expect(err).To(BeNil())
			Expect(client).ToNot(BeNil())

			mux.HandleFunc(authPath, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})

			err = client.auth(ctx)
			Expect(err).To(Equal(fmt.Errorf("http status code: 500")))
		})

	})

	Context("uci", func() {
		It("should get", func() {
			mux := http.NewServeMux()
			ts := httptest.NewServer(mux)
			defer ts.Close()
			u, err := url.Parse(ts.URL)
			Expect(err).To(BeNil())
			client, err := New("http://"+u.Host, "admin", "password", 1, true)
			Expect(err).To(BeNil())
			Expect(client).ToNot(BeNil())

			expectedToken := "foobar"
			authCalled := false
			mux.HandleFunc(authPath, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err = w.Write([]byte(`{"result":"` + expectedToken + `"}`))
				Expect(err).To(BeNil())
			})

			expectedResp := "helloworld"
			mux.HandleFunc(uciPath, func(w http.ResponseWriter, r *http.Request) {
				// auth should be called
				if !authCalled {
					authCalled = true
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				Expect(r.URL.Path).To(Equal(uciPath))
				Expect(r.RequestURI).To(Equal(uciPath + "?auth=" + expectedToken))

				w.WriteHeader(http.StatusOK)
				_, err = w.Write([]byte(`{"result":"` + expectedResp + `"}`))
				Expect(err).To(BeNil())
			})

			resp, err := client.Uci(ctx, "get", []string{"network.lan.ipaddr"})
			Expect(err).To(BeNil())
			Expect(resp).To(Equal(expectedResp))
			Expect(authCalled).To(BeTrue())
			Expect(client.token).To(Equal(expectedToken))
		})
	})
})
