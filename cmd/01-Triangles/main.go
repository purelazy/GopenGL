package main

// /home/andre/go/src/GopenGL/cmd/01-Triangles/main.go

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	runtime.LockOSThread()
}

func createWindow(title string, width, height int) *glfw.Window {
	fmt.Println("Window width, height: ", width, height)
	if !(width != 0 && height != 0) {
		fmt.Println("Width and Height cannot be zero.")
		os.Exit(0)
	}

	if err := glfw.Init(); err != nil {
		panic(fmt.Errorf("could not initialize glfw: %v", err))
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	win, err := glfw.CreateWindow(width, height, title, nil, nil)

	if err != nil {
		panic(fmt.Errorf("could not create opengl renderer: %v", err))
	}

	win.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	return win
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	fmt.Println("compileShader")
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	// Check for errors
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		// How many bytes to allocate
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))

		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		fmt.Println("compileShader log")
		fmt.Println(log)

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func createShader(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		fmt.Println("Vertex shader did not compile")
		fmt.Println(err)
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		fmt.Println("Fragment shader did not compile")
		fmt.Println(err)
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func main() {
	fmt.Println("Main")
	fmt.Println("Define Vertices")

	type vec2 struct {
		x float32
		y float32
	}

	var vertices = [...]vec2{
		// Triangle 1
		{-0.90, -0.90},
		{0.85, -0.90},
		{-0.90, 0.85},
		// Triangle 2
		{0.90, -0.85},
		{0.90, 0.90},
		{-0.85, 0.90},
	}

	fmt.Println("Create the Window")
	var windowWidth, windowHeight int = 800, 600
	win := createWindow("Hello OpenGL in Go", windowWidth, windowHeight)

	// Vertex Shader
	var vertexShader = `
		#version 430 core

		layout (location = 0) in vec4 position;

		void main()
		{
			gl_Position = position;
		}
` + "\x00"

	// Fragment Shader
	var fragmentShader = `
		#version 430 core

		out vec4 color;

		vec4 red = vec4(0.2, 0.0, 0.0, 1.0);

		void main() {
			color = red;
		}
	` + "\x00"

	// Compile, link and load the shader program
	program, err := createShader(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}
	defer gl.DeleteProgram(program)
	gl.UseProgram(program)

	fmt.Println("Create and Bind the VAO")

	var oneVAO int32 = 1
	var theVAO uint32

	gl.GenVertexArrays(oneVAO, &theVAO)
	gl.BindVertexArray(theVAO)

	fmt.Println("Create and Bind the Array Buffer")

	var oneBuffer int32 = 1
	var theArrayBuffer uint32

	gl.GenBuffers(oneBuffer, &theArrayBuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, theArrayBuffer)

	fmt.Println("Copy data to the Array Buffer")
	gl.BufferData(gl.ARRAY_BUFFER, int(unsafe.Sizeof(vertices)), unsafe.Pointer(&vertices), gl.STATIC_DRAW)

	fmt.Println("Describe the data")
	var position uint32 = 0
	coordinatesPerVertex := int32(unsafe.Sizeof(vec2{})) / int32(unsafe.Sizeof(float32(0)))
	gl.VertexAttribPointer(position, coordinatesPerVertex, gl.FLOAT, false, 0, gl.PtrOffset(0))

	fmt.Println("Use the data")
	gl.EnableVertexAttribArray(position)

	fmt.Println("Set a background colour")
	type vec4 struct {
		r float32
		g float32
		b float32
		a float32
	}
	black := vec4{0, 0, 0, 1}

	fmt.Println("Clear the screen and draw the triangles")
	const drawbuffer int32 = 0
	gl.ClearBufferfv(gl.COLOR, drawbuffer, &black.r)

	const first int32 = 0
	gl.DrawArrays(gl.TRIANGLES, first, int32(len(vertices)))

	win.SwapBuffers()

	fmt.Println("Wait for the Close window click")
	for !win.ShouldClose() {
		glfw.PollEvents()
	}
}
