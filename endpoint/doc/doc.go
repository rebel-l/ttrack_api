package doc

// Init initialises the doc endpoints.
func Init(svc *smis.Service) error {
  _, err := svc.RegisterFileServer("/doc", http.MethodGet, "endpoint/doc/web")

  return err // nolint: wrapcheck
}
