package entity

type Model struct {
	Name      string
	MaxTokens int
}

// Function to instace the struct
// it returns a model pointer
func NewModel(name string, maxTokens int) *Model {
	return &Model{
		Name:      name,
		MaxTokens: maxTokens,
	}
}

func (m *Model) GetMaxTokens() int {
	return m.MaxTokens
}

func (m *Model) GetName() string {
	return m.Name
}
