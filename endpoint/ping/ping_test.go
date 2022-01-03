package ping // nolint: testpackage

func TestPingHandler(t *testing.T) {
    t.Parallel()

    req, err := http.NewRequestWithContext(context.Background(), "GET", "/ping", nil)
    if err != nil {
        t.Fatal(err)
    }

    w := httptest.NewRecorder()

    svc, err := smis.NewService(&http.Server{}, mux.NewRouter(), logrus.New())
    if err != nil {
        t.Fatal(err)
    }

    ep := &ping{svc: svc}
    handler := http.HandlerFunc(ep.pingHandler)
    handler.ServeHTTP(w, req)

    if status := w.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    if expected := "pong"; w.Body.String() != expected {
        t.Errorf("handler returned unexpected body: got %v want %v", w.Body.String(), expected)
    }
}

func TestInit(t *testing.T) {
    t.Parallel()

    router := mux.NewRouter()
    srv := &http.Server{
        Handler:      router,
        Addr:         fmt.Sprintf(":%d", 30000),
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }
    svc := &smis.Service{
        Log:    logrus.New(),
        Router: router,
        Server: srv,
    }

    if err := Init(svc); err != nil {
        t.Fatalf("init failed: %s", err)
    }

    err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
        pathTemplate, err := route.GetPathTemplate()
        if err != nil {
            return err // nolint: wrapcheck
        }

        if pathTemplate != "/ping" {
            t.Errorf("Expected single endpoint '/ping' but got '%s'", pathTemplate)
        }

        return nil
    })
    if err != nil {
        t.Fatalf("walk through routes failed: %s", err)
    }
}
