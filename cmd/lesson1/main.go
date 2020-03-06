package main

import (
	"fmt"

	"runtime"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func createWindow(title string, width, height int) *glfw.Window {
	runtime.LockOSThread()

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

		layout (location = 0) in vec4 vPosition;

		void main()
		{
			gl_Position = vPosition;
		}
` + "\x00"

	var fragmentShader = `
		#version 430 core

		out vec4 color;

		vec4 red = vec4(0.2, 0.0, 0.0, 1.0);

		void main() {
			color = red;
		}
	` + "\x00"

	shader, err := createShader(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}
	defer gl.DeleteProgram(shader)

	//              |
	// +-------------------------+
	// |                         |
	// |    Load the Shader      |
	// |                         |
	// +-------------------------+
	//              |

	gl.UseProgram(shader)

	//              |
	// +-------------------------+
	// |                         |
	// |   Define 2 Triangles    |
	// |   Requires 6 Vertices   |
	// |                         |
	// +-------------------------+
	//              |

	const numberOfVertices int32 = 6
	// The number of components per vertex attribute, X and Y
	const coordinatesPerVertex int32 = 2
	var vertices = [numberOfVertices][coordinatesPerVertex]float32{
		// Triangle 1
		{-0.90, -0.90},
		{0.85, -0.90},
		{-0.90, 0.85},
		// Triangle 2
		{0.90, -0.85},
		{0.90, 0.90},
		{-0.85, 0.90},
	}

	//              |
	// +-------------------------+
	// |                         |
	// |  Create one VAO         |
	// |                         |
	// +-------------------------+
	//              |

	var oneVAO int32 = 1
	var theVAO uint32

	// GenVertexArrays(GLsizei n, GLuint *arrays);
	// Copies "n" currently unused names (for use as vertex-array objects)
	// to "arrays". The names returned are marked as "used" for the purposes
	// of allocating additional buffer objects, and initialized with values
	// representing the default state of the collection of uninitialized vertex
	// arrays.

	gl.GenVertexArrays(oneVAO, &theVAO)

	// glBindVertexArray() does three things. When using the value array that
	// is other than zero and was returned from glGenVertexArrays(), a new
	// vertex-array object is created and assigned that name. When binding to a
	// previously created vertex-array object, that vertex array object becomes
	// active, which additionally affects the vertex array state stored in the
	// object. When binding to an array value of zero, OpenGL stops using
	// application-allocated vertex-array objects and returns to the default state
	// for vertex arrays.

	gl.BindVertexArray(theVAO)

	//              |
	// +-------------------------+
	// |                         |
	// |  Create one Buffer      |
	// |                         |
	// +-------------------------+
	//              |

	var oneBuffer int32 = 1
	var theArrayBuffer uint32

	gl.GenBuffers(oneBuffer, &theArrayBuffer)
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
	// +-------------------------+
	//              |

	var vPosition uint32 = 0
	// Size specifies the number of components per generic vertex attribute.
	gl.VertexAttribPointer(vPosition, coordinatesPerVertex, gl.FLOAT, false, 0, gl.PtrOffset(0))

	//              |
	// +-------------------------+
	// |                         |
	// | Use vPosition in Shader |
	// |                         |
	// +-------------------------+
	//              |

	gl.EnableVertexAttribArray(vPosition)

	//              |
	// +-------------------------+
	// |                         |
	// | Define a "Clear" colour |
	// | Black in this case      |
	// |                         |
	// +-------------------------+
	//              |

	// A structure for colour data
	type vec4 struct {
		r float32
		g float32
		b float32
		a float32
	}
	black := vec4{0, 0, 0, 1}

	//              |
	// +-------------------------+
	// |                         |
	// | Loop until the window   |
	// | is closed               |
	// |                         |
	// +-------------------------+
	//              |

	for !win.ShouldClose() {

		//              |
		// +-------------------------+
		// |                         |
		// |     Clear buffer 0      |
		// |                         |
		// +-------------------------+
		//              |

		const drawbuffer int32 = 0
		gl.ClearBufferfv(gl.COLOR, drawbuffer, &black.r)

		//              |
		// +-------------------------+
		// |                         |
		// |  Draw the vertices as   |
		// |  triangles              |
		// |                         |
		// +-------------------------+
		//              |

		// Specifie the first index in the enabled array.
		const first int32 = 0
		gl.DrawArrays(gl.TRIANGLES, first, numberOfVertices)
		// gl.PointSize(10)
		// gl.DrawArrays(gl.POINTS, first, numberOfVertices)

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
		// +-------------------------+
		//              |

		win.SwapBuffers()

		//              |
		// +-------------------------+
		// |                         |
		// | See what's going on     |
		// |                         |
		// +-------------------------+
		//              |

		glfw.PollEvents()
	}
}
