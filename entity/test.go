package entity

type WithoutJson struct {
	IdString string
	ID       string
}

type WithJson struct {
	IdString string `json:"id-string"`
	ID       string `json:"id"`
}

type WithoutForm struct {
	IdString string
	ID       string
}

type WithForm struct {
	IdString string `form:"id-string"`
	ID       string `form:"id"`
}
