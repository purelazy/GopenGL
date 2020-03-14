package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
)

// All OpenGL and GLFW calls should be made same thread.
// A runtime.LockOsThread in the init() of your program is sufficient.
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

	// Use OpenGL 4.6 Core Profile
	// Window hints need to be set before the creation of the window and context
	// you wish to have the specified attributes. They function as additional
	// arguments to glfwCreateWindow.
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	// Allow it to be resized.
	glfw.WindowHint(glfw.Resizable, glfw.True)

	win, err := glfw.CreateWindow(width, height, title, nil, nil)

	if err != nil {
		panic(fmt.Errorf("could not create opengl renderer: %v", err))
	}

	return win
}

func main() {

	// Open a window
	var windowWidth, windowHeight int = 800, 600
	win := createWindow("Hello OpenGL", windowWidth, windowHeight)

	// Poll for window close
	for !win.ShouldClose() {
		glfw.PollEvents()
	}
}
