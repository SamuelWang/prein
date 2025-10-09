package models

import "fmt"

type Item struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Price float64 `json:"price"`
}

func (i *Item) Validate() error {
    if i.Name == "" {
        return fmt.Errorf("name is required")
    }
    if i.Price <= 0 {
        return fmt.Errorf("price must be greater than zero")
    }
    return nil
}