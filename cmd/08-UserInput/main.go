package main

import (
	"fmt"
	"math"

	//"math/rand"
	"os"
	"runtime"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// Returns a clojure which gets the next prime each call
func newP() func() int {
	n := 1
	return func() int {
		for {
			n++
			// Trial division as naïvely as possible.  For a candidate n,
			// numbers between 1 and n are checked to see if they divide n.
			// If no number divides n, n is prime.
			for f := 2; ; f++ {
				if f == n {
					return n
				}
				if n%f == 0 { // here is the trial division
					break
				}
			}
		}
	}
}

func createWindow(title string, width, height int) *glfw.Window {

	// Make sure width and height are not zero.
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

// Used to move the model along the z-axis
var zoom float32 = 0

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

	switch key {
	case glfw.KeyW:
		zoom += 0.1
	case glfw.KeyS:
		zoom -= 0.1
	}
	// if key == glfw.KeyW && (action == glfw.Press || action == glfw.Repeat) {
	// 	fmt.Println("W")
	// 	zoom += 0.1
	// }
	// if key == glfw.KeyS && action == glfw.Press {
	// 	fmt.Println("S")
	// 	zoom -= 0.1
	// }

}

func main() {

	// The thread running this, stays with this and only this.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	//              |
	// +-------------------------+
	// |                         |
	// |   Create a Window       |
	// |                         |
	// +-------------------------+
	//              |

	var windowWidth, windowHeight int = 1600, 1200
	win := createWindow("Hello OpenGL in Go", windowWidth, windowHeight)

	win.SetKeyCallback(keyCallback)

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
		uniform mat4 view;
		uniform mat4 model;
		
		in vec3 vert;
		out vec3 colour;
		vec3 doNotDraw = vec3(1000.0, 0.0, 0.0);
		
		void main() {
			// https://jsantell.com/model-view-projection
			// v′ = P⋅V⋅M⋅v
			gl_Position = projection * view * model * vec4(vert, 1);
		}

` + "\x00"

	var fragmentShader = `
		#version 430

		out vec4 outputColor;

		void main() {
			outputColor = vec4(1.0);
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
	// |    Projection Matrix    |
	// |                         |
	// +-------------------------+
	//              |

	// Field of View (along the Y axis)
	fovy := mgl32.DegToRad(45.0)
	// The aspect ratio
	aspectRatio := float32(windowWidth) / float32(windowHeight)
	// The near and far clipping distances
	var nearClip float32 = 0.01
	var farClip float32 = 12
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
	// |       View Matrix       |
	// |                         |
	// +-------------------------+
	//              |

	// This is where the eye/camera is positioned
	eye := mgl32.Vec3{0, 0, 1}
	// This is the point we are looking at
	lookingAt := mgl32.Vec3{0, 0, 0}
	// Up is in the positive Y direction
	thisWayIsUp := mgl32.Vec3{0, 1, 0}
	// LookAtV positions the camera
	view := mgl32.LookAtV(eye, lookingAt, thisWayIsUp)

	//              |
	// +-------------------------+
	// |                         |
	// | Link "camera" with      |
	// | the shader              |
	// |                         |
	// +-------------------------+
	//              |

	viewUniform := gl.GetUniformLocation(shader, gl.Str("view\x00"))
	gl.UniformMatrix4fv(viewUniform, 1, false, &view[0])

	//              |
	// +-------------------------+
	// |                         |
	// |      Model Matrix       |
	// |                         |
	// +-------------------------+
	//              |
	// Model is typically these three matrices multiplied in this order Translate.Rotate.Scale

	// This model is updated in the main rendering loop, below
	model := mgl32.Ident4()

	// Returns the location of a uniform variable
	modelUniform := gl.GetUniformLocation(shader, gl.Str("model\x00"))

	// Specify the value of a uniform variable for the current program object
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

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
	// |    Create some points   |
	// |                         |
	// +-------------------------+
	//              |

	type vec3 struct {
		x float32
		y float32
		z float32
	}

	const count int = 2000
	var samplePoints [count]vec3

	prime := newP()
	//p0 := float64(prime())
	for i := 0; i < count; i++ {
		p := float64(prime())
		//p := p1 - p0
		v := vec3{float32(p * math.Cos(p) / 100000), float32(p * math.Sin(p) / 100000), 0}
		samplePoints[i] = v
		//p0 = p1
	}

	//              |
	// +-------------------------+
	// |                         |
	// | Allocate memory on GPU  |
	// | and copy the vertices   |
	// | to that memory          |
	// |                         |
	// +-------------------------+
	//              |

	gl.BufferData(gl.ARRAY_BUFFER, int(unsafe.Sizeof(samplePoints)), unsafe.Pointer(&samplePoints), gl.STATIC_DRAW)

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
	coordinatesPerVertex := int32(unsafe.Sizeof(vec3{})) / int32(unsafe.Sizeof(float32(0)))
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

	gl.PointSize(2)

	var primitivesToDraw int32 = 2

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

		model = mgl32.Translate3D(0, 0, zoom)

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
		gl.DrawArrays(gl.POINTS, first, int32(len(samplePoints)))
		//gl.DrawArrays(gl.POINTS, first, primitivesToDraw%int32(count))
		primitivesToDraw++

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
