package main

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

func createWindow(title string, width, height int) *glfw.Window {
	fmt.Println("Window width, height: ", width, height)
	if !(width != 0 && height != 0) {
		fmt.Println("Width and Height cannot be zero.")
		os.Exit(0)
	}
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		panic(fmt.Errorf("could not initialize glfw: %v", err))
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Create a window
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

	var windowWidth, windowHeight int = 800, 600
	win := createWindow("Hello OpenGL in Go", windowWidth, windowHeight)

	//              |
	// +-------------------------+
	// |                         |
	// |   Create the Shader     |
	// |   (Compile & Link)      |
	// |                         |
	// +-------------------------+
	//              |

	var vertexShader = `
		#version 430

		uniform mat4 projection;
		uniform mat4 camera;
		uniform mat4 model;
		
		in vec3 vert;
		
		void main() {
			gl_Position = projection * camera * model * vec4(vert, 1);
		}

` + "\x00"

	var fragmentShader = `
		#version 430

		out vec4 outputColor;

		void main() {
			// This pixel is green
			outputColor = vec4(0.0, 1.0, 0.0, 1.0);
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
	// |  Describe camera lens   |
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

	//              |
	// +-------------------------+
	// |                         |
	// | Link "projection" with  |
	// | the shader              |
	// |                         |
	// +-------------------------+
	//              |

	projectionUniform := gl.GetUniformLocation(shader, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	//              |
	// +-------------------------+
	// |                         |
	// |   Position the Camera   |
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
	camera := mgl32.LookAtV(eye, lookingAt, thisWayIsUp)
	// glGetUniformLocation gets the location of a uniform within a program

	//              |
	// +-------------------------+
	// |                         |
	// | Link "camera" with      |
	// | the shader              |
	// |                         |
	// +-------------------------+
	//              |

	cameraUniform := gl.GetUniformLocation(shader, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

	//              |
	// +-------------------------+
	// |                         |
	// |          Model          |
	// |                         |
	// +-------------------------+
	//              |

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(shader, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	//              |
	// +-------------------------+
	// |                         |
	// |   Define 2 Triangles    |
	// |   Requires 6 Vertices   |
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
	coordinatesPerVertex := int32(unsafe.Sizeof(vec2{})) / int32(unsafe.Sizeof(float32(0)))
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
	type colour4 struct {
		r float32
		g float32
		b float32
		a float32
	}
	black := colour4{0, 0, 0, 1}

	//              |
	// +-------------------------+
	// |                         |
	// | Let's rotate the model  |
	// | by omega (in radians)   |
	// |                         |
	// +-------------------------+
	//              |

	// Start at angle 0
	angle := 0.0
	// Rotate at one revolution (2 * Pi radians) per second
	omega := 2 * math.Pi
	// GetTime returns the time elapsed since GLFW was started
	// previousTime is used to calculate the time interval (dt) between frames
	previousTime := glfw.GetTime()

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
		// | dt: time between frames |
		// |                         |
		// +-------------------------+
		//              |

		// Update
		time := glfw.GetTime()
		dt := time - previousTime
		previousTime = time

		angle += omega * dt

		model = mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})

		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

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
		gl.DrawArrays(gl.TRIANGLES, first, int32(len(vertices)))

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
		// | Poll mouse and keyboard |
		// |                         |
		// +-------------------------+
		//              |

		// Without this, clicking the close window button would not be detected
		// and you would need to use "Control-C" to stop the program.
		glfw.PollEvents()
	}
}
