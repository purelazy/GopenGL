package main

import (
	"fmt"

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

func createShader(vertexShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		fmt.Println("Vertex shader did not compile")
		fmt.Println(err)
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	// gl.LinkProgram(program)

	// var status int32
	// gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	// if status == gl.FALSE {
	// 	var logLength int32
	// 	gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

	// 	log := strings.Repeat("\x00", int(logLength+1))
	// 	gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

	// 	return 0, fmt.Errorf("failed to link program: %v", log)
	// }

	// gl.DeleteShader(vertexShader)

	return program, nil
}

func main() {

	//              |
	// +-------------------------+
	// |                         |
	// |   Create a Window       |
	// |                         |
	// +-------------------------+
	//              |

	const windowWidth int = 800
	const windowHeight int = 600
	win := createWindow("Hello OpenGL in Go", windowWidth, windowHeight)
	defer win.Destroy()

	//              |
	// +-------------------------+
	// |                         |
	// |   Create the Shader     |
	// |  (Compile & Link GLSL)  |
	// |                         |
	// +-------------------------+
	//              |

	var vertexShader = `
    #version 430 core

    in float inValue;
    out float outValue;

    void main()
    {
        outValue = sqrt(inValue);
    }
` + "\x00"

	shader, err := createShader(vertexShader)
	if err != nil {
		panic(err)
	}
	defer gl.DeleteProgram(shader)

	// Str takes a null-terminated Go string and returns its GL-compatible address.
	// This function reaches into Go string storage in an unsafe way so the caller
	// must ensure the string is not garbage collected.
	names := "outValue\x00"
	uint8Name := gl.Str(names)
	gl.TransformFeedbackVaryings(shader, 1, &uint8Name, gl.INTERLEAVED_ATTRIBS)

	// --------------------------

	gl.LinkProgram(shader)

	var status int32
	gl.GetProgramiv(shader, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(shader, logLength, nil, gl.Str(log))

		err = fmt.Errorf("failed to link program: %v", log)
		fmt.Println(err)
	}

	gl.DeleteShader(shader)

	gl.UseProgram(shader)

	// Create VAO
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// Create input VBO and vertex format
	data := [...]float32{1.0, 2.0, 3.0, 4.0, 5.0}

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, int(unsafe.Sizeof(data)), unsafe.Pointer(&data), gl.STATIC_DRAW)

	inputAttrib := gl.GetAttribLocation(shader, gl.Str("inValue\x00"))
	gl.EnableVertexAttribArray(uint32(inputAttrib))
	gl.VertexAttribPointer(uint32(inputAttrib), 1, gl.FLOAT, false, 0, unsafe.Pointer(nil))

	// Create transform feedback buffer
	var tbo uint32
	gl.GenBuffers(1, &tbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, tbo)
	gl.BufferData(gl.ARRAY_BUFFER, int(unsafe.Sizeof(data)), unsafe.Pointer(nil), gl.STATIC_READ)

	// Perform feedback transform
	gl.Enable(gl.RASTERIZER_DISCARD)

	gl.BindBufferBase(gl.TRANSFORM_FEEDBACK_BUFFER, 0, tbo)

	gl.BeginTransformFeedback(gl.POINTS)

	gl.DrawArrays(gl.POINTS, 0, 5)

	gl.EndTransformFeedback()

	gl.Disable(gl.RASTERIZER_DISCARD)

	gl.Flush()

	// Fetch and print results
	var feedback [5]float32
	gl.GetBufferSubData(gl.TRANSFORM_FEEDBACK_BUFFER, 0, int(unsafe.Sizeof(feedback)), unsafe.Pointer(&feedback))

	fmt.Println(feedback[0], feedback[1], feedback[2], feedback[3], feedback[4])

	gl.DeleteProgram(shader)
	gl.DeleteShader(shader)

	gl.DeleteBuffers(1, &tbo)
	gl.DeleteBuffers(1, &vbo)

	gl.DeleteVertexArrays(1, &vao)

}
