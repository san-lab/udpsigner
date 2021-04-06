package templates

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

type Renderer struct {
	Templates *template.Template
}

func NewRenderer() *Renderer {
	r := &Renderer{}
	r.LoadTemplates()
	return r
}

const templatedir = "./templates"
const templatesuffix = ".tmpl"

//Taken out of the constructor with the idae of forced template reloading
func (r *Renderer) LoadTemplates() {
	var allFiles []string
	files, err := ioutil.ReadDir(templatedir)
	if err != nil {
		log.Println(err)
	}
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, templatesuffix) {
			allFiles = append(allFiles, templatedir+"/"+filename)
		}
	}
	r.Templates, err = template.ParseFiles(allFiles...) //parses all .tmpl files in the 'Templates' folder
	if err != nil {
		log.Println(err)
	}
}

func (r *Renderer) RenderResponse(w io.Writer, data *RenderData) error {
	err := r.Templates.ExecuteTemplate(w, data.TemplateName, data)
	if err != nil {
		log.Println(err)
	}
	return err

}

//This is a try to bring some uniformity to passing data to the Templates
//The "RenderData" container is a wrapper for the header/body/footer containers
type RenderData struct {
	Error        error
	TemplateName string
	HeaderData   interface{ HeaderData }
	BodyData     interface{}
	FooterData   interface{}
	Client       interface{}
}

type HeaderData interface {
	GetRefresh() (interval int)
	SetRefresh(int)
}
