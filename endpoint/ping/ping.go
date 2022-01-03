package ping

type ping struct {
    svc *smis.Service
}

// Init initialises the ping endpoints.
func Init(svc *smis.Service) error {
    endpoint := &ping{svc: svc}
    _, err := svc.RegisterEndpoint("/ping", http.MethodGet, endpoint.pingHandler)

    return err // nolint: wrapcheck
}

func (p *ping) pingHandler(writer http.ResponseWriter, request *http.Request) {
    log := p.svc.NewLogForRequestID(request.Context())

    _, err := writer.Write([]byte("pong"))
    if err != nil {
        log.Errorf("ping failed: %s", err)
    }
}
