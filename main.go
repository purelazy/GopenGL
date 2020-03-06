package main

import (
	"fmt"

	// "os"
	"runtime"
	"strings"
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

layout (location = 0) in vec4 vPosition;

void main()
{
	gl_Position = vPosition;
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

// Attribute IDs
type attributeID int

const (
	vPosition attributeID = 0
)

func main() {

	//              |
	//              |
	// +-------------------------+
	// |                         |
	// |   Create a Window       |
	// |                         |
	// +------------|------------+
	//              |

	win := getWindowFromGLFW("Hello GoPenGLSLang")

	//              |
	//              |
	// +-------------------------+
	// |                         |
	// |   Create the Shader     |
	// |                         |
	// +------------|------------+
	//              |

	shader, err := newProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}
	defer gl.DeleteProgram(shader)

	//              |
	//              |
	// +-------------------------+
	// |                         |
	// |   Load the Shader       |
	// |                         |
	// +------------|------------+
	//              |

	gl.UseProgram(shader)

	//              |
	//              |
	// +-------------------------+
	// |                         |
	// |   Create the Vertices   |
	// |                         |
	// +------------|------------+
	//              |

	var vertices = [6][2]float32{
		{-0.90, -0.90}, // Triangle 1
		{0.85, -0.90},
		{-0.90, 0.85},
		{0.90, -0.85}, // Triangle 2
		{0.90, 0.90},
		{-0.85, 0.90},
	}

	//              |
	//              |
	// +-------------------------+
	// |                         |
	// |  Create one VAO         |
	// |                         |
	// +------------|------------+
	//              |

	var one int32 = 1
	var theVAO uint32
	gl.GenVertexArrays(one, &theVAO)
	gl.BindVertexArray(theVAO)

	//              |
	//              |
	// +-------------------------+
	// |                         |
	// |  Create one Buffer      |
	// |                         |
	// +------------|------------+
	//              |

	var theArrayBuffer uint32
	gl.GenBuffers(one, &theArrayBuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, theArrayBuffer)

	//              |
	// +-------------------------+
	// |                         |
	// | Allocate memory on GPU  |
	// | and copy the vertices   |
	// | to that memory          |
	// |                         |
	// +-------------------------+
	//              |

	gl.BufferData(gl.ARRAY_BUFFER, int(unsafe.Sizeof(vertices)), unsafe.Pointer(&vertices), gl.STATIC_DRAW)

	//              |
	// +-------------------------+
	// |                         |
	// | Describe the format of  |
	// | the vertex data. Size,  |
	// | type, stride, etc       |
	// |                         |
	// +------------|------------+
	//              |

	var vPosition uint32 = 0
	gl.VertexAttribPointer(vPosition, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

	//              |
	//              |
	// +-------------------------+
	// |                         |
	// | Use vPosition in Shader |
	// |                         |
	// +------------|------------+
	//              |

	gl.EnableVertexAttribArray(vPosition)

	//              |
	//              |
	// +-------------------------+
	// |                         |
	// | Define "black"          |
	// |                         |
	// +------------|------------+
	//              |

	black := vec4{0, 0, 0, 1}

	//              |
	//              |
	// +-------------------------+
	// |                         |
	// | Loop until the window   |
	// | is closed               |
	// |                         |
	// +------------|------------+
	//              |

	for !win.ShouldClose() {
		// t0 := time.Now()

		//              |
		//              |
		// +-------------------------+
		// |                         |
		// |    Clear the screen     |
		// |                         |
		// +------------|------------+
		//              |

		gl.ClearBufferfv(gl.COLOR, drawBuffer, &black[0])

		//              |
		//              |
		// +-------------------------+
		// |                         |
		// |  Draw the vertices as   |
		// |  triangles              |
		// |                         |
		// +------------|------------+
		//              |

		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		//              |
		//              |
		// +-------------------------+
		// |                         |
		// |  Make what we have      |
		// |  drawn so far visible   |
		// |                         |
		// |  Swaps the front and    |
		// |  back buffers of the    |
		// |  window                 |
		// |                         |
		// +------------|------------+
		//              |

		win.SwapBuffers()

		//              |
		//              |
		// +-------------------------+
		// |                         |
		// | See what's going on     |
		// |                         |
		// +------------|------------+
		//              |

		glfw.PollEvents()
	}
}
