package div

var flexStyle = map[string]string{
	"display":         "flex",
	"flex-wrap":       "nowrap",
	"align-items":     "center",
	"justify-content": "center",
}

var outerStyle = map[string]string{
	"display":         "flex",
	"flex-direction":  "column",
	"align-items":     "center",
	"justify-content": "center",
}

var videoHtml = `<video src="%s" controls autoplay loop muted width="100%%" height="100%%"></video>`

const deFaultVideoHtml = `<video src="C:\Games\BlueTheBone Collection\audioed\Saber Padoru (Bra).mp4" controls autoplay loop muted width="100%" height="100%"></video>`
