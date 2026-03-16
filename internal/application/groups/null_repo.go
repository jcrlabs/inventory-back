package groups

type NullRepo struct{}

func NewNullRepo() *NullRepo { return &NullRepo{} }
