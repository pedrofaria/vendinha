package print

type DebugService struct{}

func NewDebugService() *DebugService {
	return &DebugService{}
}

func (s *DebugService) Print(data []byte) error {
	println(string(data))
	return nil
}
