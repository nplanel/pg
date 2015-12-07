// +build js

package main

import (
	"bytes"
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/gopherjs/todomvc/utils"
	"github.com/gopherjs/webgl"
	"honnef.co/go/js/console"
	"html/template"
	"strconv"
	"time"
)

var VertexShadernew = `
        attribute  vec4 vPosition;
        attribute  vec4 vColor;
        varying vec4 fColor;

        uniform mat4 modelViewMatrix;
        uniform mat4 projectionMatrix;

        void main() 
        {
            gl_Position = projectionMatrix*modelViewMatrix*vPosition;
            fColor = vColor;
        } 
`
var FragmentShadernew = `
        precision mediump float;

        varying vec4 fColor;

        void
        main()
        {
            gl_FragColor = fColor;
        }
`

var VertexShader = `
        attribute  vec3 vPosition;

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;

void main() {
    gl_Position = projection * camera * model * vec4(vPosition, 1);
}
`
var FragmentShader = `
precision mediump float;

void main(){
  gl_FragColor = vec4(1,0,0,1);
}
`

var jQuery = jquery.NewJQuery //for convenience

const (
	KEY        = "TodoMVC4GopherJS"
	ENTER_KEY  = 13
	ESCAPE_KEY = 27
)

var gl *webgl.Context
var model mgl32.Mat4
var vPosition int
var program *js.Object
var modelViewMatrixLoc *js.Object

func maingl() *webgl.Context {
	document := js.Global.Get("document")
	canvas := document.Call("createElement", "canvas")
	canvas.Set("width", 320)
	canvas.Set("height", 200)
	document.Get("body").Call("appendChild", canvas)

	attrs := webgl.DefaultAttributes()
	attrs.Alpha = true
	console.Log(attrs)

	var err error
	gl, err = webgl.NewContext(canvas, attrs)
	if err != nil {
		js.Global.Call("alert", "Error: "+err.Error())
	}

	windowWidth, _ := strconv.ParseFloat(canvas.Get("width").String(), 32)
	windowHeight, _ := strconv.ParseFloat(canvas.Get("height").String(), 32)
	console.Log("wxh : %fx%f", windowWidth, windowHeight)

	vertShader := gl.CreateShader(gl.VERTEX_SHADER)
	gl.ShaderSource(vertShader, VertexShader)
	gl.CompileShader(vertShader)
	console.Assert(gl.GetShaderParameterb(vertShader, gl.COMPILE_STATUS) == true, "vertShader compile failed"+gl.GetShaderInfoLog(vertShader))

	fragShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	gl.ShaderSource(fragShader, FragmentShader)
	gl.CompileShader(fragShader)
	console.Assert(gl.GetShaderParameterb(fragShader, gl.COMPILE_STATUS) == true, "fragShader compile failed"+gl.GetShaderInfoLog(fragShader))

	program = gl.CreateProgram()
	gl.AttachShader(program, vertShader)
	gl.AttachShader(program, fragShader)
	gl.LinkProgram(program)
	console.Assert(gl.GetProgramParameterb(program, gl.LINK_STATUS) == true, "link program sharders failed"+gl.GetProgramInfoLog(program))

	g_vertex_buffer_datanew := []float32{
		-1.0, -1.0, -1.0, // triangle 1 : begin
		-1.0, -1.0, 1.0,
		-1.0, 1.0, 1.0, // triangle 1 : end
		1.0, 1.0, -1.0, // triangle 2 : begin
		-1.0, -1.0, -1.0,
		-1.0, 1.0, -1.0, // triangle 2 : end
		1.0, -1.0, 1.0,
		-1.0, -1.0, -1.0,
		1.0, -1.0, -1.0,
		1.0, 1.0, -1.0,
		1.0, -1.0, -1.0,
		-1.0, -1.0, -1.0,
		-1.0, -1.0, -1.0,
		-1.0, 1.0, 1.0,
		-1.0, 1.0, -1.0,
		1.0, -1.0, 1.0,
		-1.0, -1.0, 1.0,
		-1.0, -1.0, -1.0,
		-1.0, 1.0, 1.0,
		-1.0, -1.0, 1.0,
		1.0, -1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, -1.0, -1.0,
		1.0, 1.0, -1.0,
		1.0, -1.0, -1.0,
		1.0, 1.0, 1.0,
		1.0, -1.0, 1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, -1.0,
		-1.0, 1.0, -1.0,
		1.0, 1.0, 1.0,
		-1.0, 1.0, -1.0,
		-1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		-1.0, 1.0, 1.0,
		1.0, -1.0, 1.0,
	}

	g_vertex_buffer_datanew[0] = g_vertex_buffer_datanew[0]

	var g_vertex_buffer_data = []float32{
		-1.0, -1.0, 0.0,
		1.0, -1.0, 0.0,
		0.0, 1.0, 0.0,
	}
	g_vertex_buffer_data[0] = g_vertex_buffer_data[0]

	gl.UseProgram(program)

	cameraMatrixLoc := gl.GetUniformLocation(program, "camera")
	camera := mgl32.LookAtV(mgl32.Vec3{3, 3, 3}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(cameraMatrixLoc, false, camera[:])

	modelViewMatrixLoc = gl.GetUniformLocation(program, "model")
	model = mgl32.Ident4()
	gl.UniformMatrix4fv(modelViewMatrixLoc, false, (model)[:])

	projectionMatrixLoc := gl.GetUniformLocation(program, "projection")
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth/windowHeight), 0.1, 10.0)
	gl.UniformMatrix4fv(projectionMatrixLoc, false, projection[:])

	VertexBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, VertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, g_vertex_buffer_datanew, gl.STATIC_DRAW)
	vPosition := gl.GetAttribLocation(program, "vPosition")
	gl.EnableVertexAttribArray(vPosition)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, 0)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0.8, 0.3, 0.01, 0.7)

	return gl
}

var angle = float32(0.0)

var fps = 50
var period = time.Duration(1000/fps) * time.Millisecond
var gltimer *time.Timer

var start = time.Now()

var everyt = make(chan time.Duration)

func loopgl() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	model = mgl32.HomogRotate3D(float32(mgl32.DegToRad(angle)), mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(modelViewMatrixLoc, false, model[:])

	gl.DrawArrays(gl.TRIANGLES, 0, 12*3)
	//		gl.DisableVertexAttribArray(vPosition)
	//	angle += 0.1
	end := time.Now()
	every := end.Sub(start)
	everyt <- every
	start = end

	gltimer.Reset(period)
}

func main() {
	fmt.Println("app go main")
	console.Clear()
	console.Log("app go main")
	console.Trace()

	app := NewApp()
	app.bindEvents()
	app.initRouter()
	app.render()
	maingl()

	gltimer = time.AfterFunc(period, loopgl)
	gltimer = gltimer
	for {
		fmt.Println("call each time :", <-everyt)
	}
}

type ToDo struct {
	Id        string
	Text      string
	Completed bool
}
type App struct {
	todos           []ToDo
	todoTmpl        *template.Template
	footerTmpl      *template.Template
	todoAppJQuery   jquery.JQuery
	headerJQuery    jquery.JQuery
	mainJQuery      jquery.JQuery
	footerJQuery    jquery.JQuery
	sliderJQuery    jquery.JQuery
	newTodoJQuery   jquery.JQuery
	toggleAllJQuery jquery.JQuery
	todoListJQuery  jquery.JQuery
	countJQuery     jquery.JQuery
	clearBtnJQuery  jquery.JQuery
	filter          string
}

func NewApp() *App {
	somethingToDo := make([]ToDo, 0)
	utils.Retrieve(KEY, &somethingToDo)

	todoHtml := jQuery("#todo-template").Html()
	todoTmpl := template.Must(template.New("todo").Parse(todoHtml))

	footerHtml := jQuery("#footer-template").Html()
	footerTmpl := template.Must(template.New("footer").Parse(footerHtml))

	todoAppJQuery := jQuery("#todoapp")
	headerJQuery := todoAppJQuery.Find("#header")
	mainJQuery := todoAppJQuery.Find("#main")
	footerJQuery := todoAppJQuery.Find("#footer")
	newTodoJQuery := headerJQuery.Find("#new-todo")
	toggleAllJQuery := mainJQuery.Find("#toggle-all")
	todoListJQuery := mainJQuery.Find("#todo-list")
	countJQuery := footerJQuery.Find("#todo-count")
	clearBtnJQuery := footerJQuery.Find("#clear-completed")
	sliderJQuery := headerJQuery.Find("#slider")
	filter := "all"

	return &App{somethingToDo, todoTmpl, footerTmpl, todoAppJQuery, headerJQuery, mainJQuery, footerJQuery, sliderJQuery, newTodoJQuery, toggleAllJQuery, todoListJQuery, countJQuery, clearBtnJQuery, filter}
}
func (a *App) bindEvents() {

	a.newTodoJQuery.On(jquery.KEYUP, a.create)
	a.toggleAllJQuery.On(jquery.CHANGE, a.toggleAll)
	a.footerJQuery.On(jquery.CLICK, "#clear-completed", a.destroyCompleted)
	a.sliderJQuery.On("slide", a.slider)

	a.todoListJQuery.On(jquery.CHANGE, ".toggle", a.toggle)
	a.todoListJQuery.On(jquery.DBLCLICK, "label", a.edit)
	a.todoListJQuery.On(jquery.KEYUP, ".edit", a.blurOnEnter)
	a.todoListJQuery.On(jquery.FOCUSOUT, ".edit", a.update)
	a.todoListJQuery.On(jquery.CLICK, ".destroy", a.destroy)

	js.Global.Set("mydebug", mydebug)
}
func mydebug() *js.Object {
	console.Log("mydebug")
	o := js.Global.Get("slider_value")
	console.Log(o)
	return o
}
func (a *App) initRouter() {
	router := utils.NewRouter()
	router.On("/:filter", func(filter string) {
		a.filter = filter
		a.render()
	})
	router.Init("/all")
}
func (a *App) render() {
	todos := a.getFilteredTodos()

	var b bytes.Buffer
	a.todoTmpl.Execute(&b, todos)
	strtodoTmpl := b.String()

	a.todoListJQuery.SetHtml(strtodoTmpl)
	a.mainJQuery.Toggle(len(a.todos) > 0)
	a.toggleAllJQuery.SetProp("checked", len(a.getActiveTodos()) != 0)
	a.renderfooter()
	a.newTodoJQuery.Focus()
	utils.Store(KEY, a.todos)
}
func (a *App) renderfooter() {
	activeTodoCount := len(a.getActiveTodos())
	activeTodoWord := utils.Pluralize(activeTodoCount, "item")
	completedTodos := len(a.todos) - activeTodoCount
	filter := a.filter
	footerData := struct {
		ActiveTodoCount int
		ActiveTodoWord  string
		CompletedTodos  int
		Filter          string
	}{
		activeTodoCount, activeTodoWord, completedTodos, filter,
	}
	var b bytes.Buffer
	a.footerTmpl.Execute(&b, footerData)
	footerJQueryStr := b.String()

	a.footerJQuery.Toggle(len(a.todos) > 0).SetHtml(footerJQueryStr)
}
func (a *App) toggleAll(e jquery.Event) {
	checked := !a.toggleAllJQuery.Prop("checked").(bool)
	for idx := range a.todos {
		a.todos[idx].Completed = checked
	}
	a.render()
}
func (a *App) getActiveTodos() []ToDo {
	todosTmp := make([]ToDo, 0)
	for _, val := range a.todos {
		if !val.Completed {
			todosTmp = append(todosTmp, val)
		}
	}
	return todosTmp
}
func (a *App) getCompletedTodos() []ToDo {
	todosTmp := make([]ToDo, 0)
	for _, val := range a.todos {
		if val.Completed {
			todosTmp = append(todosTmp, val)
		}
	}
	return todosTmp
}
func (a *App) getFilteredTodos() []ToDo {
	switch a.filter {
	case "active":
		return a.getActiveTodos()
	case "completed":
		return a.getCompletedTodos()
	default:
		return a.todos
	}
}
func (a *App) destroyCompleted(e jquery.Event) {
	a.todos = a.getActiveTodos()
	a.filter = "all"
	a.render()
}
func (a *App) slider(e jquery.Event) {
	angle = float32(js.Global.Get("slider_value").Float())
	//	console.Log(angle)
}
func (a *App) indexFromEl(e jquery.Event) int {
	id := jQuery(e.Target).Closest("li").Data("id")
	for idx, val := range a.todos {
		if val.Id == id {
			return idx
		}
	}
	return -1
}
func (a *App) create(e jquery.Event) {
	val := jquery.Trim(a.newTodoJQuery.Val())
	if len(val) == 0 || e.KeyCode != ENTER_KEY {
		return
	}
	newToDo := ToDo{Id: utils.Uuid(), Text: val, Completed: false}
	a.todos = append(a.todos, newToDo)
	a.newTodoJQuery.SetVal("")
	a.render()
}
func (a *App) toggle(e jquery.Event) {
	idx := a.indexFromEl(e)
	a.todos[idx].Completed = !a.todos[idx].Completed
	a.render()
}
func (a *App) edit(e jquery.Event) {
	input := jQuery(e.Target).Closest("li").AddClass("editing").Find(".edit")
	input.SetVal(input.Val()).Focus()
}
func (a *App) blurOnEnter(e jquery.Event) {
	switch e.KeyCode {
	case ENTER_KEY:
		jQuery(e.Target).Blur()
	case ESCAPE_KEY:
		jQuery(e.Target).SetData("abort", "true").Blur()
	}
}
func (a *App) update(e jquery.Event) {

	thisJQuery := jQuery(e.Target)
	val := jquery.Trim(thisJQuery.Val())
	if thisJQuery.Data("abort") == "true" {
		thisJQuery.SetData("abort", "false")
		a.render()
		return
	}
	idx := a.indexFromEl(e)
	if len(val) > 0 {
		a.todos[idx].Text = val
	} else {
		a.todos = append(a.todos[:idx], a.todos[idx+1:]...)
	}
	a.render()
}
func (a *App) destroy(e jquery.Event) {
	idx := a.indexFromEl(e)
	a.todos = append(a.todos[:idx], a.todos[idx+1:]...)
	a.render()
}
