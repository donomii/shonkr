package main

import (
	"fmt"
	"github.com/go-gl/gl/v3.2-core/gl"
)

type Renderer struct {
	program     uint32
	vao         uint32
	vbo         uint32
	ebo         uint32
	texture     uint32
	samplerLoc  int32
	texW, texH  int
	initialized bool
}

// Fullscreen quad (positions, texcoords)
// NDC positions with texcoords mapped 0..1
var quadVertices = []float32{
	// x,   y,   u,  v
	-1.0, -1.0, 0.0, 0.0,
	1.0, -1.0, 1.0, 0.0,
	1.0, 1.0, 1.0, 1.0,
	-1.0, 1.0, 0.0, 1.0,
}

var quadIndices = []uint32{
	0, 1, 2,
	2, 3, 0,
}

const vertexSrc = `#version 150
in vec2 aPos;
in vec2 aUV;
out vec2 vUV;
void main() {
    vUV = aUV;
    gl_Position = vec4(aPos, 0.0, 1.0);
}`

const fragmentSrc = `#version 150
in vec2 vUV;
out vec4 fragColor;
uniform sampler2D uTex;
void main() {
    // Flip Y because our CPU buffer origin is top-left
    fragColor = texture(uTex, vec2(vUV.x, 1.0 - vUV.y));
}`

func (r *Renderer) Init() error {
	if r.initialized {
		return nil
	}

	vs, err := compileShader(vertexSrc, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	fs, err := compileShader(fragmentSrc, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	r.program = gl.CreateProgram()
	gl.AttachShader(r.program, vs)
	gl.AttachShader(r.program, fs)
	gl.LinkProgram(r.program)

	var status int32
	gl.GetProgramiv(r.program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLen int32
		gl.GetProgramiv(r.program, gl.INFO_LOG_LENGTH, &logLen)
		log := make([]byte, logLen)
		gl.GetProgramInfoLog(r.program, logLen, nil, &log[0])
		return fmt.Errorf("program link error: %s", string(log))
	}

	gl.GenVertexArrays(1, &r.vao)
	gl.BindVertexArray(r.vao)

	gl.GenBuffers(1, &r.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(quadVertices)*4, gl.Ptr(quadVertices), gl.STATIC_DRAW)

	gl.GenBuffers(1, &r.ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(quadIndices)*4, gl.Ptr(quadIndices), gl.STATIC_DRAW)

	posLoc := uint32(gl.GetAttribLocation(r.program, gl.Str("aPos\x00")))
	uvLoc := uint32(gl.GetAttribLocation(r.program, gl.Str("aUV\x00")))

	// stride = 4 floats per vertex (2 pos + 2 uv)
	stride := int32(4 * 4)
	gl.EnableVertexAttribArray(posLoc)
	gl.VertexAttribPointer(posLoc, 2, gl.FLOAT, false, stride, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(uvLoc)
	gl.VertexAttribPointer(uvLoc, 2, gl.FLOAT, false, stride, gl.PtrOffset(2*4))

	gl.GenTextures(1, &r.texture)
	gl.BindTexture(gl.TEXTURE_2D, r.texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	r.samplerLoc = gl.GetUniformLocation(r.program, gl.Str("uTex\x00"))
	r.initialized = true
	return nil
}

func (r *Renderer) UpdateTexture(pix []uint8, w, h int) {
	if !r.initialized {
		return
	}
	gl.BindTexture(gl.TEXTURE_2D, r.texture)
	if w != r.texW || h != r.texH {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(w), int32(h), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pix))
		r.texW, r.texH = w, h
	} else {
		gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pix))
	}
}

func (r *Renderer) Draw() {
	if !r.initialized {
		return
	}
	gl.UseProgram(r.program)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, r.texture)
	gl.Uniform1i(r.samplerLoc, 0)
	gl.BindVertexArray(r.vao)
	gl.DrawElements(gl.TRIANGLES, int32(len(quadIndices)), gl.UNSIGNED_INT, gl.PtrOffset(0))
}

func compileShader(src string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csrc, free := gl.Strs(src + "\x00")
	gl.ShaderSource(shader, 1, csrc, nil)
	free()
	gl.CompileShader(shader)
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLen int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLen)
		log := make([]byte, logLen)
		gl.GetShaderInfoLog(shader, logLen, nil, &log[0])
		return 0, fmt.Errorf("shader compile error: %s", string(log))
	}
	return shader, nil
}
