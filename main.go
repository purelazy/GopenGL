package main

import (
	"fmt"
	"math"

	// "os"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	windowWidth  = 960
	windowHeight = 540
)

func getWindowFromGLFW(title string) *glfw.Window {
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		panic(fmt.Errorf("could not initialize glfw: %v", err))
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	win, err := glfw.CreateWindow(800, 600, title, nil, nil)

	if err != nil {
		panic(fmt.Errorf("could not create opengl renderer: %v", err))
	}

	win.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	// version := gl.GoStr(gl.GetString(gl.VERSION))
	// fmt.Println("OpenGL version", version)

	return win
}

type vec3 [3]float32
type vec4 [4]float32

const drawBuffer int32 = 0

var vertexShader = `
#version 430 core

// This is a vertex attribute
layout (location = 0) in vec4 offset;

void main() {

vec4 vertices[3] = vec4[3] (
		vec4(0.25, -0.25, 0.0, 1.0),
		vec4(-0.25, -0.25, 0.0, 1.0),
		vec4(0.25, 0.25, 0.0, 1.0));

	
	vertices[gl_VertexID].x *= offset.x;
	

	// gl_VertexID 
	gl_Position = vertices[gl_VertexID] + offset;
}
` + "\x00"

var fragmentShader = `
#version 430 core

out vec4 color;

void main() {
    color = vec4(0.0, 0.8, 1.0, 1.0);
}
` + "\x00"

func compileShader(source string, shaderType uint32) (uint32, error) {
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

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		fmt.Println("Vertex Compile Error")
		fmt.Println(err)
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
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

// The following gets me names for Vertex Array IDs, the count and storage for the names
type vertexArrayID int

const (
	triangles       vertexArrayID = 0
	numVertexArrays vertexArrayID = 1
)

var vertexArrays [numVertexArrays]uint32

// The following gets me names for Buffer IDs, the count and storage for the names
type bufferID int

const (
	arrayBuffer bufferID = 0
	numBuffers  bufferID = 1
)

var buffers [numBuffers]uint32

// Attribute IDs
type attributeID int

const (
	vPosition attributeID = 0
)

func prepareVerticies(shader uint32) {

	// Opaque VAO
	gl.GenVertexArrays(int32(numVertexArrays), &vertexArrays[0])
	gl.BindVertexArray(vertexArrays[triangles])

	// Vertices
	var vertices = [6][2]float32{
		{-0.90, -0.90}, // Triangle 1
		{0.85, -0.90},
		{-0.90, 0.85},
		{0.90, -0.85}, // Triangle 2
		{0.90, 0.90},
		{-0.85, 0.90},
	}

	// Copy vertices to the GPU
	gl.GenBuffers(int32(numBuffers), &buffers[0])
	gl.BindBuffer(gl.ARRAY_BUFFER, buffers[arrayBuffer])
	gl.BufferData(gl.ARRAY_BUFFER, int(unsafe.Sizeof(vertices)), unsafe.Pointer(&vertices), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(shader, gl.Str("vert\x00")))
	gl.VertexAttribPointer(vertAttrib, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(vertAttrib)
}

func main() {

	win := getWindowFromGLFW("Hello GoPenGLSLang")

	prepareVerticies()

	//gl.ClearColor(0, 0.5, 1.0, 1.0)

	// Configure the vertex and fragment shaders
	shader, err := newProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}
	defer gl.DeleteProgram(shader)

	prepareVerticies(shader)

	var vertexArray uint32

	// Create a vertex array so we can draw
	gl.GenVertexArrays(1, &vertexArray)
	gl.BindVertexArray(vertexArray)
	defer gl.DeleteVertexArrays(1, &vertexArray)

	gl.PointSize(64)

	//                                       Background Colour
	colour := vec4{0, 0, 0, 1}
	clipOffset := vec4{0, 0, 0, 0}

	radianPerSecond := math.Pi * 2
	var angle float64 = 0

	for !win.ShouldClose() {
		// gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		colour[0] = float32(math.Sin(angle)*0.5 + 0.5)
		colour[1] = float32(math.Cos(angle)*0.5 + 0.5)

		clipOffset[0] = float32(math.Sin(angle) * 0.5)
		clipOffset[1] = float32(math.Cos(angle) * 0.5)

		t0 := time.Now()

		gl.ClearBufferfv(gl.COLOR, drawBuffer, &colour[0])

		gl.UseProgram(shader)

		// Send a float to the vertex shader
		// func gl.VertexAttrib4fv(index uint32, v *float32)
		gl.VertexAttrib4fv(0, &clipOffset[0])

		// gl.DrawArrays(gl.POINTS, 0, 1)
		gl.DrawArrays(gl.TRIANGLES, 0, 3)

		win.SwapBuffers()
		glfw.PollEvents()

		t1 := time.Now()

		// dt for each frame
		delta := t1.Sub(t0).Milliseconds()
		//fmt.Printf("FPS = %v\n", 1/(float32(delta)/1000))

		angle += (radianPerSecond / 60) * float64(delta) / 1000.0
	}
}
