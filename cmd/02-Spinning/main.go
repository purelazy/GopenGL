package main

// /home/andre/go/src/GopenGL/cmd/01-Triangles/main.go

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func init() {
	runtime.LockOSThread()
}

func createWindow(title string, width, height int) *glfw.Window {
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
	// |        Window           |
	// |                         |
	// +-------------------------+
	//              |

	var windowWidth, windowHeight int = 800, 600
	win := createWindow("Hello OpenGL in Go", windowWidth, windowHeight)

	//              |
	// +-------------------------+
	// |                         |
	// |        Vertices         |
	// |                         |
	// +-------------------------+
	//              |

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

	//              |
	// +-------------------------+
	// |                         |
	// |         Shader          |
	// |                         |
	// +-------------------------+
	//              |

	var vertexShader = `
		#version 430 core

		layout (location = 0) in vec4 position;

		uniform mat4 projection;
		uniform mat4 view;
		uniform mat4 model;

		void main()
		{
			gl_Position = projection * view * model * position;
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

	//              |
	// +-------------------------+
	// |                         |
	// |      Vertex Array       |
	// |                         |
	// +-------------------------+
	//              |

	var oneBuffer int32 = 1
	var theArrayBuffer uint32

	gl.GenBuffers(oneBuffer, &theArrayBuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, theArrayBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, int(unsafe.Sizeof(vertices)), unsafe.Pointer(&vertices), gl.STATIC_DRAW)

	//              |
	// +-------------------------+
	// |                         |
	// |      Vertex Array       |
	// |      Description        |
	// |                         |
	// +-------------------------+
	//              |

	var oneVAO int32 = 1
	var theVAO uint32

	gl.GenVertexArrays(oneVAO, &theVAO)
	gl.BindVertexArray(theVAO)

	var position uint32 = 0
	coordinatesPerVertex := int32(unsafe.Sizeof(vec2{})) / int32(unsafe.Sizeof(float32(0)))
	gl.VertexAttribPointer(position, coordinatesPerVertex, gl.FLOAT, false, 0, gl.PtrOffset(0))

	// Enable this attribute in the shader
	gl.EnableVertexAttribArray(position)

	//              |
	// +-------------------------+
	// |                         |
	// |          Model          |
	// |                         |
	// +-------------------------+
	//              |

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	//              |
	// +-------------------------+
	// |                         |
	// |       View Matrix       |
	// |                         |
	// +-------------------------+
	//              |

	// This is where the camera is positioned
	eye := mgl32.Vec3{3, 3, 3}
	// This is the point at which the camera is looking
	lookingAt := mgl32.Vec3{0, 0, 0}
	// Up is in the positive Y direction
	thisWayIsUp := mgl32.Vec3{0, 1, 0}
	// LookAtV positions the camera based on these 3 things
	view := mgl32.LookAtV(eye, lookingAt, thisWayIsUp)

	viewLocation := gl.GetUniformLocation(program, gl.Str("view\x00"))
	gl.UniformMatrix4fv(viewLocation, 1, false, &view[0])

	//              |
	// +-------------------------+
	// |                         |
	// |    Projection Matrix    |
	// |                         |
	// +-------------------------+
	//              |

	// Field of View (along the Y axis)
	fovy := mgl32.DegToRad(45.0)
	// The aspect ratio
	aspectRatio := float32(windowWidth) / float32(windowHeight)
	// The near and far clipping distances
	var nearClip float32 = 0.1
	var farClip float32 = 10
	// Perspective generates a Perspective Matrix.
	projection := mgl32.Perspective(fovy, aspectRatio, nearClip, farClip)

	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	// Background colour
	type vec4 struct {
		r float32
		g float32
		b float32
		a float32
	}
	black := vec4{0, 0, 0, 1}

	// Angle, angular velocity,
	angle := 0.0
	omega := 2 * math.Pi
	previousTime := glfw.GetTime()

	// Poll for window close
	for !win.ShouldClose() {

		// Clear screen
		const drawbuffer int32 = 0
		gl.ClearBufferfv(gl.COLOR, drawbuffer, &black.r)

		//              |
		// +-------------------------+
		// |                         |
		// |      Rotate Model       |
		// |                         |
		// +-------------------------+
		//              |

		time := glfw.GetTime()
		dt := time - previousTime
		previousTime = time
		angle += omega * dt

		model = mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

		//              |
		// +-------------------------+
		// |                         |
		// |          Draw           |
		// |                         |
		// +-------------------------+
		//              |

		const first int32 = 0
		gl.DrawArrays(gl.TRIANGLES, first, int32(len(vertices)))
		win.SwapBuffers()

		glfw.PollEvents()
	}
}
