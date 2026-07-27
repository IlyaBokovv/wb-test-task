package delivery

type Delivery struct {
	OrderUID string `json:"-" gorm:"primaryKey;size:64"`
	Name     string `json:"name" validate:"required"`
	Phone    string `json:"phone" validate:"required,len=10|len=11|len=12"`
	Zip      string `json:"zip" validate:"required"`
	City     string `json:"city" validate:"required"`
	Address  string `json:"address" validate:"required"`
	Region   string `json:"region" validate:"required"`
	Email    string `json:"email" validate:"omitempty,email"`
}
